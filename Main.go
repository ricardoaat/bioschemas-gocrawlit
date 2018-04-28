package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"

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
		crawl(*u)
	}
}

func crawl(u string) error {

	if u == "" {
		log.Error("Empty URL")
		return errors.New("The URL must not be empty")
	}
	url, err := url.Parse(u)
	if err != nil {
		return err
	}
	log.Debug("Parsed host ", url.Host)

	fn := fmt.Sprintf("%s_schema.yaml", url.Host)
	fout, err := os.Create(fn)
	if err != nil {
		log.Error("Fail to create file. Check your file path and permissions")
		return err
	}
	defer fout.Close()

	c := colly.NewCollector(

		colly.AllowedDomains(url.Host, fmt.Sprintf("www.%s", url.Host)),
		colly.MaxDepth(2),
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir(fmt.Sprintf("./%s_cache", url.Host)),
	)

	c.OnHTML(`script[type="application/ld+json"]`, func(e *colly.HTMLElement) {
		log.Warn("Script found ", e.Request.URL)
		log.Info(e.Text)
		fout.WriteString(fmt.Sprintf(`%s\n%s`, e.Request.URL, e.Text))
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
