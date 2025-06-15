package anime

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

type Jkanime struct{}

func (j Jkanime) GetLatestEpisodes() ([]LatestEpisode, error) {
	var episodes []LatestEpisode

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64)..."),
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{Parallelism: 5, Delay: 500 * time.Millisecond})

	c.OnHTML("#animes .card a", func(e *colly.HTMLElement) {
		slug := strings.Split(e.Attr("href"), "/")[3]
		img := e.ChildAttr("img", "src")
		title := e.ChildText("h5")
		epText := e.ChildText(".badge-primary")
		epParts := strings.Fields(epText)
		episode := ""
		if len(epParts) > 1 {
			episode = epParts[1]
		}

		episodes = append(episodes, LatestEpisode{
			Slug:    slug,
			Img:     img,
			Title:   title,
			Episode: episode,
		})
	})

	err := c.Visit("https://jkanime.net/")
	if err != nil {
		return nil, err
	}

	c.Wait()
	return episodes, nil
}

func (j Jkanime) GetAnime(slug string) (*Anime, error) {
	if slug == "" {
		return nil, errors.New("slug cannot be empty")
	}

	anime := &Anime{
		AdditionalInfo: make(map[string]interface{}),
	}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64)..."),
	)

	// Título
	c.OnHTML(".anime_info h3", func(e *colly.HTMLElement) {
		anime.Title = strings.TrimSpace(e.Text)
	})

	// Sinopsis
	c.OnHTML(".anime_info .scroll", func(e *colly.HTMLElement) {
		anime.Synopsis = strings.TrimSpace(e.Text)
	})

	// Imagen
	c.OnHTML(".anime_pic img", func(e *colly.HTMLElement) {
		anime.Img = e.Attr("src")
	})

	// Información adicional tipo: 'Género', 'Estado', 'Estudio', etc.
	c.OnHTML(".card-bod ul li", func(e *colly.HTMLElement) {
		key := ""
		values := []string{}

		e.DOM.Contents().Each(func(i int, s *goquery.Selection) {
			if goquery.NodeName(s) == "span" && key == "" {
				// es el label, lo usamos como clave
				key = strings.Trim(strings.TrimSuffix(s.Text(), ":"), " ")
				key = strings.ToLower(key)
			} else {
				text := strings.TrimSpace(s.Text())
				if text != "" && text != "," {
					values = append(values, text)
				}
			}
		})

		if key != "" {
			if len(values) == 1 {
				anime.AdditionalInfo[key] = values[0]
			} else if len(values) > 1 {
				anime.AdditionalInfo[key] = values
			}
		}
	})

	err := c.Visit("https://jkanime.net/" + slug)
	if err != nil {
		return nil, err
	}

	return anime, nil
}

func (j Jkanime) GetEpisodes(slug string, page int) (*Episode, error) {

	if slug == "" {
		return nil, errors.New("slug cannot be empty")
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var episode Episode

	if page == 0 {
		page = 1
	}

	var url string

	if page == 1 {
		url = fmt.Sprintf("https://jkanime.net/%s/", slug)
	} else {
		url = fmt.Sprintf("https://jkanime.net/%s/#pag%d", slug, page)
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`
			(() => ({
				total_pages:  document.querySelectorAll('.anime__pagination .option').length,
				total_episodes: parseInt(document.querySelector('#uep')?.href.split('/')[4]),
				last_episode: parseInt(document.querySelector('#uep')?.href.split('/')[4]),
				episodes: Array.from(document.querySelectorAll('#episodes-content .anime__item')).map(item => ({
					title: document.querySelector('.anime_info h3').textContent,
					img: item.querySelector('.anime__item__pic').dataset.setbg,
					slug: item.querySelector('a').href.split('/')[3],
					episode: item.querySelector('a').href.split('/')[4]
				}))
			}))()
		`, &episode),
	)
	if err != nil {
		return nil, err
	}

	episode.Page = page

	return &episode, nil
}

func (j Jkanime) GetServers(slug, episode string) ([]Server, error) {
	if slug == "" {
		return nil, errors.New("slug cannot be empty")
	}

	if episode == "" {
		return nil, errors.New("episode cannot be empty")
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var servers []Server

	err := chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf("https://jkanime.net/%s/%s", slug, episode)),
		chromedp.Evaluate(`(() => {
			const desu = document.querySelector('#btn-show-0').textContent
			const magi = document.querySelector('#btn-show-1').textContent

			const srcUrls = video.map(iframe => {
			const match = iframe.match(/src="([^"]+)"/);
			return match ? match[1] : null;
			});

			const excludedServers = ['Mega', 'Mediafire', 'Mixdrop', 'Mp4upload', 'SaveFiles'];

			let videos = [
				{
					server: desu,
					remote: btoa(encodeURIComponent(srcUrls[0]))
				},
				{
					server: magi,
					remote: btoa(encodeURIComponent(srcUrls[1]))
				},
				...servers
			]

			return videos
				.filter(video => !excludedServers.includes(video.server))
				.map(({ server, remote }) => ({ server, remote }));
		})()`, &servers),
	)

	if err != nil {
		return nil, err
	}

	return servers, nil
}

func (j Jkanime) GetStreaming(server, slug string) (string, error) {
	if server == "" {
		return "", errors.New("server cannot be empty")
	}

	if slug == "" {
		return "", errors.New("slug cannot be empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(slug)
	if err != nil {
		return "", err
	}

	decodedStr, err := url.QueryUnescape(string(decoded))
	if err != nil {
		panic(err)
	}

	fmt.Println("Decoded URL:", decodedStr)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var script string

	switch server {
	case "Desu":
		script = `parts.segments.swarmId`
	case "Magi":
		script = `player.options_.sources[0].src`
	case "Streamwish":
		script = `player.getConfig().playlist[0].file`
	case "Vidhide":
		script = `player.getConfig().playlist[0].file`
	case "Filemoon":
		script = `jwplayer().getConfig().playlist[0].file`
	case "VOE":
		script = `jwplayer().getConfig().playlist[0].file`
	case "Streamtape":
		script = `player.source`
	default:
		return "", errors.New("unsupported server")
	}

	var streaming string

	err = chromedp.Run(ctx,
		chromedp.Navigate(decodedStr),
		chromedp.Evaluate(script, &streaming),
	)

	if err != nil {
		return "", err
	}

	return streaming, nil
}

func (j Jkanime) GetSearch(name string, page int) ([]Anime, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	var results []Anime

	c := colly.NewCollector()

	c.OnHTML(".anime__item", func(e *colly.HTMLElement) {
		anime := Anime{
			Title:          strings.TrimSpace(e.ChildText("h5")),
			Img:            e.ChildAttr(".anime__item__pic", "data-setbg"),
			Synopsis:       "",
			AdditionalInfo: map[string]interface{}{},
		}

		href := e.ChildAttr("a", "href")
		u, err := url.Parse(href)
		if err == nil {
			segments := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(segments) > 0 {
				anime.Slug = segments[0]
			}
		}

		firstLi := e.ChildText("ul li")
		if firstLi != "" {
			anime.AdditionalInfo["estado"] = strings.TrimSpace(firstLi)
		}

		tipo := e.ChildText("li.anime")
		if tipo != "" {
			anime.AdditionalInfo["tipo"] = strings.TrimSpace(tipo)
		}

		results = append(results, anime)
	})

	searchURL := fmt.Sprintf("https://jkanime.net/buscar/%s", url.PathEscape(name))
	if page > 1 {
		searchURL += fmt.Sprintf("/%d", page)
	}

	if err := c.Visit(searchURL); err != nil {
		return nil, err
	}

	return results, nil
}
