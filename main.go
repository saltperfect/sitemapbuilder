package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	link "github.com/saltperfect/linkparser"
)

/*
	1.
*/
const xmlAttr string = "www.hello.com"

type loc struct {
	Value string `xml:"loc"`
}

type urlSet struct {
	Urls  []loc  `xml:"url"`
	UrlNs string `xml:"xmlns,attr"`
}

func main() {
	targetUrl := flag.String("url", "https://gophercises.com", "target url whose sitemap needs to be made")
	maxDepth := flag.Int("maxDepth", 5, "maximum depth to to which you need to traverse")
	flag.Parse()
	links := bfs(*targetUrl, *maxDepth)
	var urls []loc
	for _, link := range links {
		urls = append(urls, loc{Value: link})
	}
	urlSet := urlSet{
		Urls:  urls,
		UrlNs: xmlAttr,
	}
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "  ")
	err := enc.Encode(urlSet)
	if err != nil {
		panic(err)
	}
	fmt.Println()
}

func bfs(urlStr string, maxDepth int) []string {
	seen := make(map[string]struct{})
	var q map[string]struct{}
	nq := map[string]struct{}{
		urlStr: {},
	}
	for i := 0; i < maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})
		if len(q) == 0 {
			break
		}
		for url := range q {
			if _, ok := seen[url]; !ok {
				seen[url] = struct{}{}
				for _, link := range get(url) {
					if _, ok := seen[link]; !ok {
						nq[link] = struct{}{}
					}
				}
			}
		}
	}
	ret := make([]string, 0, len(seen))
	for url := range seen {
		ret = append(ret, url)
	}
	return ret
}

func get(urlStr string) []string {
	resp, err := http.Get(urlStr)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	reqUrl := resp.Request.URL
	baseUrl := url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}
	base := baseUrl.String()

	return filter(hrefs(resp.Body, base), withPrefix(base))
}

func hrefs(r io.Reader, base string) []string {
	links, _ := link.Parse(r)
	var hrefs []string
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			hrefs = append(hrefs, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			hrefs = append(hrefs, l.Href)
		}
	}
	return hrefs
}

func filter(links []string, keepFn func(string) bool) []string {
	var ret []string
	for _, link := range links {
		if keepFn(link) {
			ret = append(ret, link)
		}
	}
	return ret
}

func withPrefix(prefix string) func(string) bool {
	return func(s string) bool {
		return strings.HasPrefix(s, prefix)
	}
}
