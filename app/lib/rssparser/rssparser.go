package rssparser

import (
	"bytes"
	"ereader-rss/app/lib/utils"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-epub"
	"github.com/gofiber/fiber/v2/log"

	"encoding/xml"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Generator     string `xml:"generator"`
	WebMaster     string `xml:"webMaster"`
	Languague     string `xml:"languague"`
	LastBuildDate string `xml:"lastBuildDate"`
	Ttl           int    `xml:"ttl"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Guid        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Category    string `xml:"category"`
}

// Returns an RSS struct from the xmlData
func ReadRSS(xmlData *[]byte) (*RSS, error) {
	var rss RSS
	err := xml.Unmarshal(*xmlData, &rss)
	if err != nil {
		return &RSS{}, err
	}

	return &rss, nil
}

// Remove classes and remove semantic and interactive tags
func CleanHtml(d string) (string, error) {
	reader := strings.NewReader(d)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return "", err
	}

	// Unwrap every node and remove their divs
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s.Parent()) == "div" {
			s.Unwrap()
		}
		s.RemoveClass()
		otherUnwrapDiv(s, doc)
	})

	// Remove unknown tags
	doc.Find("body").Find("*").Each(func(i int, s *goquery.Selection) {

		switch goquery.NodeName(s) {
		case "a", "abbr", "address", "article", "b",
			"blockquote", "body", "cite", "code", "em",
			"figcaption", "figure", "footer", "h1", "h2",
			"h3", "h4", "h5", "h6", "head", "header", "hr",
			"html", "i", "img", "li", "link", "main", "nav",
			"ol", "p", "pre", "section", "span", "strong", "table",
			"tbody", "td", "th", "thead", "time", "title", "tr", "ul",
			"br", "picture", "source":
			break
		default:
			s.Remove()
		}
	})

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		s.AddClass("image")
	})

	cleanHtml, err := doc.Html()
	if err != nil {
		return "", err
	}

	return cleanHtml, nil
}

func otherUnwrapDiv(c *goquery.Selection, doc *goquery.Document) {
	c.Find("*").Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s.Parent()) == "div" {
			s.Unwrap()
		}
		s.RemoveClass()
		otherUnwrapDiv(s, doc)
	})
}

type Page struct {
	Title string
	Body  *string
}

func (rss *RSS) CreateEpub() (*[]byte, error) {
	e, err := epub.NewEpub(rss.Channel.Title)
	if err != nil {
		return nil, err
	}
	e.SetTitle(rss.Channel.Title)
	e.SetDescription(rss.Channel.Description)

	cssPath, err := addCustomCSS(e)
	if err != nil {
		cssPath = ""
	}

	pages, err := processRSSItems(rss, e)
	if err != nil {
		return nil, err
	}

	err = addTOCAndPages(e, pages, cssPath)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	_, err = e.WriteTo(&buffer)
	if err != nil {
		return nil, err
	}
	bytesWritten := buffer.Bytes()
	return &bytesWritten, nil
}

func addCustomCSS(e *epub.Epub) (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := path.Join(workDir, "/app/lib/rssparser/", "styles.css")
	return e.AddCSS(path, "")
}

func processRSSItems(rss *RSS, e *epub.Epub) ([]Page, error) {
	var pages []Page
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, item := range rss.Channel.Items {
		wg.Add(1)
		go func(item Item) {
			defer wg.Done()
			page, err := processRSSItem(item, rss, e)
			if err != nil {
				log.Error("Error processing RSS item:", err)
				return
			}
			mu.Lock()
			pages = append(pages, page)
			mu.Unlock()
		}(item)
	}
	wg.Wait()
	return pages, nil
}

func processRSSItem(item Item, rss *RSS, e *epub.Epub) (Page, error) {
	cleanHtml, err := CleanHtml(item.Description)
	if err != nil {
		return Page{}, err
	}

	// Its easier to parse the HTML creating a doc again
	reader := strings.NewReader(cleanHtml)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return Page{}, err
	}

	// Find and replace all img's src with an internal link to the image
	// that is compressed
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists {
			log.Error("src not found at: ", item.Title)
			return
		}
		log.Debug("making request to: ", src)
		path, err := utils.CompressImage(src)
		if err != nil {
			return
		}
		e.AddImage(path, utils.GetMD5Hash(src))
	})

	// Add a header for each page that is similar as seen in the website
	headerHTML := fmt.Sprintf(`
        <div class='header'>
            <h1>%s</h1>
            <p class='header-item'>
                <span class='last-build-date'>%s</span>
            </p>
            <p class='header-item'>
                <a href='%s'>%s</a>
            </p>
            <div class='bar strong'></div>
        </div>`,
		item.Title,
		rss.Channel.LastBuildDate,
		item.Link,
		item.Link,
	)
	doc.Find("body").BeforeHtml(headerHTML)

	html, err := doc.Html()
	if err != nil {
		return Page{}, err
	}

	log.Debug("page parsed: ", item.Title)
	return Page{Title: item.Title, Body: &html}, nil
}

func addTOCAndPages(e *epub.Epub, pages []Page, cssPath string) error {
	tocPage := "<nav epub:type='toc' id='toc'><h1>Table of Contents</h1><ul>"
	for _, page := range pages {
		tocPage += fmt.Sprintf(
			"<li class='item'><a href='%s.xhtml'>%s</a></li>",
			page.Title,
			page.Title,
		)
	}
	tocPage += "</ul></nav>"
	e.AddSection(tocPage, "Table of Contents", "Table of Contents", cssPath)

	for _, page := range pages {
		_, err := e.AddSection(*page.Body, page.Title, page.Title, cssPath)
		if err != nil {
			log.Error("error writing page: ", err)
			return err
		}
	}
	return nil
}
