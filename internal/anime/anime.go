package anime

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/chromedp/chromedp"
)

type Jkanime struct{}

func (j Jkanime) GetLatestEpisodes() ([]LatestEpisode, error) {
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

	var animes []LatestEpisode

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://jkanime.net/"),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('#animes .card a')).map(anime => ({
				slug: anime.href.split('/')[3],
				img: anime.querySelector('img').src,
				title: anime.querySelector('h5').textContent,
				episode: anime.querySelector('.badge-primary').textContent.replace(/\s+/g, ' ').trim().split(' ')[1]
			}));
		`, &animes),
	)
	if err != nil {
		return nil, err
	}

	return animes, nil
}

func (j Jkanime) GetAnime(slug string) (*Anime, error) {

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

	var anime Anime

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://jkanime.net/"+slug),
		chromedp.Evaluate(`
			(() => ({
				title: document.querySelector('.anime_info h3').textContent,
				synopsis: document.querySelector('.anime_info .scroll').textContent,
				img: document.querySelector('.anime_pic img').src
			}))()
		`, &anime),
		chromedp.Evaluate(AddionalAnimeInfoJsCode, &anime.AdditionalInfo),
	)
	if err != nil {
		return nil, err
	}

	return &anime, nil
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

	var result []Anime

	err := chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf("https://jkanime.net/buscar/%s", name)),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.anime__item')).map(item => ({
			title: item.querySelector('h5').textContent,
			slug: item.querySelector('a').href.split('/')[3],
			img: document.querySelector('.anime__item__pic').dataset.setbg,
			additional_info: {
				estado: item.querySelector('li').textContent,
				tipo: item.querySelector('li.anime').textContent.trim(),
			}
		}))`, &result),
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}
