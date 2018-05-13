package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

type pageData struct {
	Page     string `json:"page"`
	Metadata string `json:"data"`
}

// Crawler structure calls collys crawler and
//	its configured to extract microdata and JSON-LD metadata.
type Crawler struct {
	C              *colly.Collector
	BaseURL        *url.URL
	SkipQueries    bool
	MaxDepth       int
	AllowedDomains []string
	PagesData      []pageData
	Filter         string
	QueryWord      string
}

// Init setup the initial configuration for the crawler
// based on the parameter given when the crawler instance is created.
func (cw *Crawler) Init() {
	cacheDir := fmt.Sprintf("bioschemas_gocrawlit_cache/%s_cache", cw.BaseURL.Host)

	cw.C = colly.NewCollector(
		// MaxDepth is 1, so only the links on the scraped page
		// is visited, and no further links are followed
		colly.MaxDepth(cw.MaxDepth),

		colly.AllowedDomains(cw.AllowedDomains...),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir(cacheDir),

		// MaxDepth is 1, so only the links on the scraped page
		// is visited, and no further links are followed
		colly.MaxDepth(cw.MaxDepth),

		//Visit only root url and urls
		colly.URLFilters(
			regexp.MustCompile(cw.Filter),
		),
	)

	cw.C.OnError(func(r *colly.Response, err error) {
		log.WithFields(log.Fields{
			"URL":      r.Request.URL,
			"RespCode": r.StatusCode,
			"Error":    err,
		}).Error("Failed request")
	})

	cw.C.OnHTML(`script[type="application/ld+json"]`, func(e *colly.HTMLElement) {
		log.Warn("Script found ", e.Request.URL)
		log.Debug(e.Text)

		cw.PagesData = append(cw.PagesData, pageData{e.Request.URL.String(), e.Text})
	})

	cw.C.OnHTML(`html`, func(e *colly.HTMLElement) {
		child := e.DOM.Find(`[itemtype^='http://schema.org']`)

		if child.Length() > 0 {
			log.Warn("Found itemtype schema.org", e.Request.URL)
			html, err := e.DOM.Html()
			if err != nil {
				log.Error("Error getting HTML")
			}

			json, err := extractMicrodata(html, cw.BaseURL)

			if err != nil {
				log.Error("Error calling extractMicrodata ", err)
				return
			}
			cw.PagesData = append(cw.PagesData, pageData{e.Request.URL.String(), string(json)})
		}

		//time.Sleep(1 * time.Second)
	})

	cw.C.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		u, err := url.Parse(e.Request.AbsoluteURL(link))
		if err != nil {
			log.Error("Error parsing URL ", err)
		}

		re := regexp.MustCompile(fmt.Sprintf("^?%s=.*", cw.QueryWord))
		if len(cw.QueryWord) > 0 && re.MatchString(u.RawQuery) {
			cw.C.Visit(u.String())
			return
		}

		if cw.SkipQueries {

			u.RawQuery = ""
			log.Debug("Found link ", u.String())
			cw.C.Visit(u.String())
			return
		}

		log.Debug("Found link ", u.String())
		cw.C.Visit(u.String())

	})

	cw.C.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		log.Info(e.Text)

		u, err := url.Parse(e.Text)
		if err != nil {
			log.Error("Error parsing URL ", err)
		}
		log.Info(u)
		cw.C.Visit(u.String())
	})

	// Before making a request print "Visiting ..."
	cw.C.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Accept", "text/html")
		log.Info("Visiting ", r.URL.String())
	})
}

// Start visits the url given as entry point starting
// starting the crawling process.
func (cw *Crawler) Start() {
	cw.C.Visit(cw.BaseURL.String())
}

func extractMicrodata(html string, baseURL *url.URL) ([]byte, error) {
	var json []byte

	p := NewParser(strings.NewReader(html), baseURL)
	data, err := p.Parse()
	if err != nil {
		log.Error("Error parsing microdata from HTML ", html)
		return json, err
	}

	json, err = data.JSON()
	if err != nil {
		log.Error("Error getting JSON from microdata HTML ")
		return json, err
	}

	return json, nil

}

// ToJSONfile creates a file with the information
// store on PagesData which contains the microdata and JSON-LD metadata
// extracted from the pages visited.
func (cw *Crawler) ToJSONfile() error {
	nq := cw.BaseURL
	nq.RawQuery = ""
	p := strings.Replace(cw.BaseURL.Path, "/", "_", -1)
	f := fmt.Sprintf("%s%sschema.json", cw.BaseURL.Host, p)

	j, err := json.Marshal(cw.PagesData)
	if err != nil {
		log.Error("Error at marshalling pagesData to json ", err)
		return err
	}

	log.Info("Creating file ", f)
	err = ioutil.WriteFile(f, j, 0644)
	if err != nil {
		log.Error("Error writing output file ", err)
		return err
	}
	return nil
}
