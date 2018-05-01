package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	microdata "github.com/bioschemas/bioschemas-gocrawlit/getmicrodata"
	"github.com/gocolly/colly"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

var (
	version   string
	buildDate string
)

func logInit(d bool) {

	logfile := "biocrawlit.log"
	fmt.Println("Loging to " + logfile)
	log.SetOutput(os.Stdout)
	if d {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetFormatter(&log.TextFormatter{})
	pathMap := lfshook.PathMap{
		log.DebugLevel: logfile,
		log.InfoLevel:  logfile,
		log.ErrorLevel: logfile,
		log.WarnLevel:  logfile,
		log.PanicLevel: logfile,
	}
	log.AddHook(lfshook.NewHook(
		pathMap,
		&log.JSONFormatter{},
	))
}

func main() {

	d := flag.Bool("d", false, "Sets up the log level to debug")
	v := flag.Bool("v", false, "Returns the binary version and built date info")
	u := flag.String("u", "", "Url to crawl and extract markup")

	flag.Parse()

	logInit(*d)

	log.Info("--------------Init program--------------")
	log.Info(fmt.Sprintf("Version: %s Build Date: %s", version, buildDate))

	if !*v {
		if err := crawl(*u); err != nil {
			log.Error(err)
		}
		//test(*u)
	}
}

func test(u string) {
	html := `
	`

	url, err := url.Parse(u)
	if err != nil {
		log.Error("Error parsing URL")
	}
	log.Info("Parsed host ", url.Host)

	p := microdata.NewParser(strings.NewReader(html), url)
	data, err := p.Parse()
	if err != nil {
		log.Error("Error parsing microdata from HTML ", html)
	}

	json, err := data.JSON()
	if err != nil {
		log.Error("Error getting JSON from microdata HTML ")
	}

	out := fmt.Sprintf("%s", json)

	fmt.Println(out)

}

func crawl(u string) error {
	log.Info("URL to crawl ", u)

	if u == "" {
		log.Error("Empty URL")
		return errors.New("The URL must not be empty")
	}
	baseURL, err := url.Parse(u)
	if err != nil {
		return err
	}
	log.Info("Parsed host ", baseURL.Host)

	fn := fmt.Sprintf("%s_schema.yaml", baseURL.Host)
	fout, err := os.Create(fn)
	if err != nil {
		log.Error("Fail to create file. Check your file path and permissions")
		return err
	}
	defer fout.Close()

	cacheDir := fmt.Sprintf(".bioschemas_gocrawlit_cache/%s_cache", baseURL.Host)

	c := colly.NewCollector(

		colly.AllowedDomains(baseURL.Host, fmt.Sprintf("www.%s", baseURL.Host)),
		colly.MaxDepth(2),
		//colly.Async(true),
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir(cacheDir),

		// Visit only root url and urls
		colly.URLFilters(
			regexp.MustCompile(u),
		),
	)

	// Parallelism can be controlled also by spawning fixed
	// number of go routines.
	//c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Error("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.OnHTML(`script[type="application/ld+json"]`, func(e *colly.HTMLElement) {
		log.Warn("Script found ", e.Request.URL)
		log.Info(e.Text)
		fout.WriteString(fmt.Sprintf(`%s\n%s`, e.Request.URL, e.Text))
	})

	c.OnHTML(`html`, func(e *colly.HTMLElement) {
		child := e.DOM.Find(`[itemtype^='http://schema.org']`)

		if child.Length() > 0 {
			log.Warn("Found itemtype bioschemas")
			html, err := e.DOM.Html()
			if err != nil {
				log.Error("Error getting HTML")
			}

			fout.WriteString(fmt.Sprintf("\n%s - itemtype %s\n", e.Request.URL, e.Attr("itemtype")))

			json, err := extractMicrodata(html, baseURL)

			if err != nil {
				log.Error("Error calling extractMicrodata ", err)
				return
			}
			fout.Write(json)
		}

		time.Sleep(1 * time.Second)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		log.WithFields(log.Fields{
			"Text": e.Text,
			"Link": link,
		}).Debug("Link found")

		c.Visit(e.Request.AbsoluteURL(link))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Accept", "text/html")
		log.Info("Visiting ", r.URL.String())
	})

	c.Visit(u)

	return nil
}

func extractMicrodata(html string, baseURL *url.URL) ([]byte, error) {
	var json []byte

	p := microdata.NewParser(strings.NewReader(html), baseURL)
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
