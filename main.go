package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
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

const url = "https://rpilocator.com/feed/?country=US&cat=PI4"

func main() {
	// Create/Open SQLite database
	db, err := sql.Open("sqlite3", "rss.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create "items" table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS items (guid TEXT PRIMARY KEY)")
	if err != nil {
		log.Fatal(err)
	}

	// Create an index on the "guid" column if it doesn't exist
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_guid ON items (guid)")
	if err != nil {
		log.Fatal(err)
	}

	// Fetch XML data
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var rss Rss
	err = xml.Unmarshal(data, &rss)
	if err != nil {
		log.Fatal(err)
	}

	// Create a slice to store GUIDs from XML
	xmlGUIDs := make([]string, len(rss.Channel.Items))

	for i, item := range rss.Channel.Items {
		xmlGUIDs[i] = item.GUID
	}

	// Query existing GUIDs from the database using the list of XML GUIDs
	query := fmt.Sprintf("SELECT guid FROM items WHERE guid IN (?" + strings.Repeat(",?", len(xmlGUIDs)-1) + ")")
	queryParams := make([]any, len(xmlGUIDs))
	for i, p := range xmlGUIDs {
		queryParams[i] = any(p)
	}
	rows, err := db.Query(query, queryParams...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Create a map to track existing GUIDs
	existingGUIDs := make(map[string]bool)
	for rows.Next() {
		var guid string
		err = rows.Scan(&guid)
		if err != nil {
			log.Fatal(err)
		}
		existingGUIDs[guid] = true
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Print new GUIDs not found in the database
	for _, guid := range xmlGUIDs {
		if _, exists := existingGUIDs[guid]; !exists {
			fmt.Println("New GUID found:", guid)

			// Insert the new GUIDs into the database
			_, err := db.Exec("INSERT INTO items (guid) VALUES (?)", guid)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
