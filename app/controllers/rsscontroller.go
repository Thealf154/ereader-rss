package lib

import (
	"ereader-rss/app/lib/rssparser"
	"ereader-rss/app/lib/utils"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

var client = memcache.New(os.Getenv("MEMCACHE_HOST") + ":" + os.Getenv("MEMCACHE_PORT"))

type Item struct {
	Title string
	Link  string
}

func GetPageFromRSS() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		rssUrl := cleanUrl(ctx.Query("url"))
		index, err := strconv.Atoi(ctx.Query("index"))

		if err != nil {
			return err
		}

		rssFeed, err := getRss(rssUrl)
		if err != nil {
			return err
		}

		item := rssFeed.Channel.Items[index]
		url, _ := url.Parse(rssUrl)

		cleanHtml, err := rssparser.CleanHtml(item.Description)

		if err != nil {
			return err
		}

		return ctx.Render("rss", fiber.Map{
			"RSSLanguague":     rssFeed.Channel.Languague,
			"RSSLastBuildDate": rssFeed.Channel.LastBuildDate,
			"ItemTitle":        item.Title,
			"ItemHost":         url.Host,
			"ItemLink":         url,
			"ItemCategory":     item.Category,
			"ItemDescription":  cleanHtml,
		})
	}
}

func GetListFromRSS() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		rssUrl := cleanUrl(ctx.Query("url"))
		rss, err := getRss(rssUrl)

		if err != nil {
			return err
		}

		items := []Item{}
		for i, item := range rss.Channel.Items {
			link := "/rss/page?url=" + rssUrl + "&index=" + strconv.Itoa(i)
			items = append(items, Item{Title: item.Title, Link: link})
		}

		url, _ := url.Parse(cleanUrl(rss.Channel.Link))

		return ctx.Render("rssList", fiber.Map{
			"RSSTitle":         rss.Channel.Title,
			"RSSLink":          url,
			"RSSDescription":   rss.Channel.Description,
			"RSSHost":          url.Host,
			"RSSLastBuildDate": rss.Channel.LastBuildDate,
			"Items":            items,
			"DownloadEpubLink": "/rss/epub/download.epub?url=" + rssUrl,
		})
	}

}

func GetRssAsEpub() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		rssUrl := cleanUrl(ctx.Query("url"))
		rssFeed, err := getRss(rssUrl)

		if err != nil {
			return err
		}

		epub, err := rssFeed.CreateEpub()

		if err != nil {
			return err
		}

		ctx.Set("Content-Type", "application/epub+zip")
		return ctx.Send(*epub)
	}
}

func cleanUrl(rawUrl string) string {
	rawUrl = strings.Trim(strings.Replace(rawUrl, "\n", "", -1), " ")

	if !strings.HasPrefix(rawUrl, "http://") && !strings.HasPrefix(rawUrl, "https://") {
		rawUrl = "http://" + rawUrl
	}

	return rawUrl
}

func getRss(rssUrl string) (*rssparser.RSS, error) {
	cache, err := getRssCache(rssUrl)

	if err != nil {
		//cache miss
		if err == memcache.ErrCacheMiss {
			log.Debug("Cache miss")
		}

		urlData, err := utils.ReadBytesFromUrl(rssUrl)
		if err != nil {
			return nil, err
		}

		data, err := rssparser.ReadRSS(urlData)
		if err != nil {
			return nil, err
		}

		setRssCache(rssUrl, urlData)
		return data, nil
	}

	//cache hit
	data, err := rssparser.ReadRSS(&cache.Value)
	if err != nil {
		return nil, err
	}

	log.Debug("Cache hit")

	return data, nil
}

func setRssCache(key string, data *[]byte) error {
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	seconds := int32(endOfDay.Sub(now).Seconds())

	cacheErr := client.Set(&memcache.Item{
		Key:        key,
		Value:      *data,
		Expiration: seconds,
	})

	if cacheErr != nil {
		log.Errorf("error setting cache: %s", cacheErr)
		return cacheErr
	}

	return nil
}

func getRssCache(key string) (*memcache.Item, error) {
	val, err := client.Get(key)
	if err != nil {
		return nil, err
	}
	return val, nil
}
