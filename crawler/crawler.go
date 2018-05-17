package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
)

type pageData struct {
	Page     string                 `json:"page"`
	Metadata map[string]interface{} `json:"data"`
}

// Crawler structure calls collys crawler and
//	its configured to extract microdata and JSON-LD metadata.
type Crawler struct {
	Index          string
	C              *colly.Collector
	BaseURL        *url.URL
	UseElastic     bool
	SkipQueries    bool
	MaxDepth       int
	AllowedDomains []string
	PagesData      []pageData
	Filter         string
	QueryWord      string
	ElasticClient  *elastic.Client
	OutputFileName string
	Client         *elastic.Client
	OutFile        *os.File
}

// Init setup the initial configuration for the crawler
// based on the parameter given when the crawler instance is created.
func (cw *Crawler) Init() {
	f, err := os.Create(cw.OutputFileName)
	if err != nil {
		log.Error("Error opening file ", f)
	}
	cw.OutFile = f

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

		var res map[string]interface{}
		if err := json.Unmarshal([]byte(e.Text), &res); err != nil {
			log.Error("Error getting MAP from microdata json result ")
		}
		pageData := pageData{e.Request.URL.String(), res}

		if cw.UseElastic {
			cw.sendToElastic(pageData)
		}

		cw.sendToJSONfile(pageData)
	})

	cw.C.OnHTML(`html`, func(e *colly.HTMLElement) {
		child := e.DOM.Find(`[itemtype^='http://schema.org']`)

		if child.Length() > 0 {
			log.Warn("Found itemtype schema.org", e.Request.URL)
			html, err := e.DOM.Html()
			if err != nil {
				log.Error("Error getting HTML")
			}

			res, err := extractMicrodata(html, cw.BaseURL)
			if err != nil {
				log.Error("Error calling extractMicrodata ", err)
				return
			}

			pageData := pageData{e.Request.URL.String(), res}

			if cw.UseElastic {
				cw.sendToElastic(pageData)
			}

			cw.sendToJSONfile(pageData)
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
	defer cw.OutFile.Close()
	cw.C.Visit(cw.BaseURL.String())
}

func (cw *Crawler) sendToElastic(p pageData) {
	ctx := context.Background()
	data := p.Metadata
	data["page"] = p.Page
	_, err := cw.Client.Index().Index(cw.Index).Type("page").BodyJson(data).Do(ctx)
	if err != nil {
		log.Panic("Error indexig ", p.Page)
	}
}

func (cw *Crawler) sendToJSONfile(p pageData) {
	data := p.Metadata
	data["page"] = p.Page

	j, err := json.Marshal(data)
	if err != nil {
		log.Error("Error at marshalling lemap to json ", err)
	}

	_, err = cw.OutFile.WriteString(string(j) + "\n")
	if err != nil {
		log.Error("Error writing output file line ", err)
	}
}

// ElasticInit sets up the initial configuration for the
//	crawler's elastic interface
func (cw *Crawler) ElasticInit() error {

	ctx := context.Background()

	c, err := elastic.NewClient(
		elastic.SetSniff(false),
	)
	if err != nil {
		log.Panic("Error creating elastic client ", err)
		return err
	}

	cw.Client = c

	inf, code, err := cw.Client.Ping("http://127.0.0.1:9200").Do(ctx)
	if err != nil {
		log.Panic("Error Pinging elastic client ", err)
		return err
	}
	log.Info(fmt.Sprintf("Elasticsearch returned with code %d and version %s\n", code, inf.Version.Number))

	ex, err := cw.Client.IndexExists(cw.Index).Do(ctx)
	if err != nil {
		log.Panic("Error fetchin index existence ", err)
		return err
	}

	if !ex {
		in, err := cw.Client.CreateIndex(cw.Index).Do(ctx)
		log.Info("Creating index " + cw.Index)
		if err != nil {
			log.Panic("Error creating index  ", err)
		}

		if !in.Acknowledged {
			log.Error("Index creation not acknowledged")
		}
	}

	return nil
}

func extractMicrodata(html string, baseURL *url.URL) (map[string]interface{}, error) {
	var res map[string]interface{}

	p := NewParser(strings.NewReader(html), baseURL)
	data, err := p.Parse()
	if err != nil {
		log.Error("Error parsing microdata from HTML ", html)
		return res, err
	}

	r, err := data.JSON()
	if err != nil {
		log.Error("Error getting JSON from microdata HTML ")
		return res, err
	}

	if err := json.Unmarshal(r, &res); err != nil {
		log.Error("Error getting MAP from microdata json result ")
		return res, err
	}
	return res, nil

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
