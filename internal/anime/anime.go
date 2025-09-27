package anime

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
	"github.com/chromedp/chromedp"
)

type Jkanime struct{}

func (j Jkanime) GetLatestEpisodes() ([]LatestEpisode, error) {
	var episodes []LatestEpisode

	client, _ := cloudscraper.Init(false, false)
	res, err := client.Get("https://jkanime.net/", make(map[string]string), "")

	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Body))
	if err != nil {
		return nil, err
	}

	doc.Find("#animes .card").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".card-title").Text()
		link, _ := s.Find("a").Attr("href")
		slug := strings.Split(link, "/")[3]
		img, _ := s.Find("img").Attr("src")
		epText := s.Find(".badge-primary").Text()
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

	return episodes, nil
}

func (j Jkanime) GetAnime(slug string) (*Anime, error) {
	if slug == "" {
		return nil, errors.New("slug cannot be empty")
	}

	anime := &Anime{
		AdditionalInfo: make(map[string]interface{}),
	}

	client, _ := cloudscraper.Init(false, false)
	res, err := client.Get("https://jkanime.net/"+slug, make(map[string]string), "")
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Body))
	if err != nil {
		return nil, err
	}

	// Título
	anime.Title = strings.TrimSpace(doc.Find(".anime_info h3").Text())

	// Sinopsis
	anime.Synopsis = strings.TrimSpace(doc.Find(".anime_info .scroll").Text())

	// Imagen
	anime.Img, _ = doc.Find(".anime_pic img").Attr("src")

	// Información adicional tipo: 'Género', 'Estado', 'Estudio', etc.
	doc.Find(".card-bod ul li").Each(func(i int, s *goquery.Selection) {
		key := ""
		values := []string{}

		s.Contents().Each(func(i int, s *goquery.Selection) {
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
		chromedp.WaitVisible("#episodes-content .anime__item", chromedp.ByQuery),
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
		script = `dp.options.video.url`
	case "Magi":
		script = `player.options_.sources[0].src`
	case "Streamwish":
		script = `
			(() => {
				const file = jwplayer().getConfig().playlist[0].file;
				let url;

				if (file.startsWith("http://") || file.startsWith("https://")) {
					url = file;
				} else {
					url = window.location.origin + file;
				}
			
				return url;
			})()
		`
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

	client, _ := cloudscraper.Init(false, false)
	res, err := client.Get(fmt.Sprintf("https://jkanime.net/buscar/%s", url.PathEscape(name)), make(map[string]string), "")
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Body))
	if err != nil {
		return nil, err
	}

	doc.Find(".anime__item").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("h5").Text())
		img, _ := s.Find(".anime__item__pic").Attr("data-setbg")
		link, _ := s.Find("a").Attr("href")
		slug := strings.Split(link, "/")[3]

		firstLi := s.Find("ul li").First().Text()
		tipo := s.Find("li.anime").Text()

		additionalInfo := make(map[string]interface{})
		if firstLi != "" {
			additionalInfo["estado"] = strings.TrimSpace(firstLi)
		}
		if tipo != "" {
			additionalInfo["tipo"] = strings.TrimSpace(tipo)
		}

		results = append(results, Anime{
			Title:          title,
			Img:            img,
			Slug:           slug,
			Synopsis:       "",
			AdditionalInfo: additionalInfo,
		})
	})

	return results, nil
}
