package main

import (
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Rss struct {
	Channel struct {
		Items []struct {
			Description string `xml:"description"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

type Config struct {
	Pushover struct {
		APIToken string `yaml:"token"`
		UserKey  string `yaml:"user"`
		Device   string `yaml:"device"`
	} `yaml:"pushover"`
}

const (
	feedURL     = "https://rpilocator.com/feed/?country=US&cat=%s"
	pushoverURL = "https://api.pushover.net/1/messages.json"
	age         = 10 * time.Minute
)

var piOptions = []string{"pi4", "pizero2"}

func main() {
	var piVersion string

	if len(os.Args) != 2 {
		log.Fatalf("Require raspberry version as argument (%s).", strings.Join(piOptions, ", "))
	}

	for _, v := range piOptions {
		if os.Args[1] == v {
			piVersion = os.Args[1]
		}
	}

	if piVersion == "" {
		log.Fatalf("Require one of (%s)", strings.Join(piOptions, ", "))
	}

	config, err := readConfig("./config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	response, err := http.Get(fmt.Sprintf(feedURL, piVersion))
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
		log.Printf("Query for %s successful. No items found.\n", piVersion)
		return
	}

	out := fmt.Sprintf(
		"device=%s&token=%s&user=%s&title=%s&message=%s",
		config.Pushover.Device,
		config.Pushover.APIToken,
		config.Pushover.UserKey,
		"🚨 Raspberry Pis In Stock 🚨",
		url.QueryEscape(strings.Join(items, "\n")),
	)
	resp, err := http.Post(pushoverURL, "application/x-www-form-urlencoded", strings.NewReader(out))
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		log.Fatalf("Getting bad status codes from Pushover: %d\n", resp.StatusCode)
	}

	log.Printf("Query for %s successful. Items found.\n", piVersion)
}

func readConfig(filename string) (Config, error) {
	var config Config

	execPath, err := os.Executable()
	if err != nil {
		return config, err
	}

	absPath := filepath.Join(filepath.Dir(execPath), filename)
	f, err := os.Open(absPath)
	if err != nil {
		return config, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
