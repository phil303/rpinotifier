package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
	Categories  []string `xml:"category"`
	GUID        string   `xml:"guid"`
	PubDate     string   `xml:"pubDate"`
}

const (
	feedURL     = "https://rpilocator.com/feed/?country=US&cat=PI4"
	pushoverURL = "https://api.pushover.net/1/messages.json"
	age         = 5 * time.Minute
	apiToken    = ""
	userKey     = ""
	device      = ""
)

func main() {
	response, err := http.Get(feedURL)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var rss Rss
	err = xml.Unmarshal(data, &rss)
	if err != nil {
		log.Fatal(err)
	}

	var items []string
	for _, item := range rss.Channel.Items {
		pubDate, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			log.Fatal(err)
		}

		if time.Since(pubDate) < age {
			items = append(items, " • "+item.Description+" Visit "+item.Link)
		}
	}

	if len(items) == 0 {
		return
	}

	title := "🚨 Raspberry Pis In Stock 🚨"
	message := url.QueryEscape(strings.Join(items, "\n"))
	out := fmt.Sprintf("device=%s&token=%s&user=%s&title=%s&message=%s", device, apiToken, userKey, title, message)
	resp, err := http.Post(pushoverURL, "application/x-www-form-urlencoded", strings.NewReader(out))
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		log.Fatalf("Getting bad status codes from Pushover: %d\n", resp.StatusCode)

	}
}
