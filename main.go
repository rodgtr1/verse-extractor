package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// RSS is the top-level structure for an RSS feed
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

// Channel represents the channel element in an RSS feed
type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

// Item represents an individual item in an RSS feed
type Item struct {
	Title          string `xml:"title"`
	Link           string `xml:"link"`
	Description    string `xml:"description"`
	PubDate        string `xml:"pubDate"`
	ContentEncoded string `xml:"encoded"`
}

// fetchVerse gets the latest verse from the Fighter Verses RSS feed
func fetchVerse() (string, error) {
	// URL is hardcoded to Fighter Verses RSS feed
	feedURL := "https://www.fighterverses.com/blog-feed.xml"

	// Fetch the RSS feed
	resp, err := http.Get(feedURL)
	if err != nil {
		return "", fmt.Errorf("error fetching RSS feed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received status code %d", resp.StatusCode)
	}

	// Read the response body
	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Parse the XML
	var rss RSS
	err = xml.Unmarshal(xmlData, &rss)
	if err != nil {
		return "", fmt.Errorf("error parsing XML: %v", err)
	}

	if len(rss.Channel.Items) == 0 {
		return "", fmt.Errorf("no items found in the feed")
	}

	// Get the most recent item (first item in the feed)
	mostRecentItem := rss.Channel.Items[0]

	// Extract blockquote using regex
	re := regexp.MustCompile(`<blockquote>.*?</blockquote>`)
	blockquote := re.FindString(mostRecentItem.ContentEncoded)

	if blockquote == "" {
		return "", fmt.Errorf("no blockquote found in the most recent item")
	}

	// Remove HTML tags and return only the text
	cleanText := removeHTMLTags(blockquote)
	return cleanText, nil
}

// removeHTMLTags removes HTML tags from a string
func removeHTMLTags(html string) string {
	// First, replace some common entities
	html = strings.ReplaceAll(html, "&apos;", "'")
	html = strings.ReplaceAll(html, "&quot;", "\"")
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")

	// Replace <br> and variants with newlines
	html = strings.ReplaceAll(html, "<br>", "\n")
	html = strings.ReplaceAll(html, "<br/>", "\n")
	html = strings.ReplaceAll(html, "<br />", "\n")

	// Remove all HTML tags
	re := regexp.MustCompile("<[^>]*>")
	return strings.TrimSpace(re.ReplaceAllString(html, ""))
}

// verseHandler handles the request for a verse
func verseHandler(w http.ResponseWriter, r *http.Request) {
	verse, err := fetchVerse()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching verse: %v", err), http.StatusInternalServerError)
		return
	}

	// Set plain text content type
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Write verse as plain text
	w.Write([]byte(verse))
}

// healthHandler provides a simple health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	// Get port from environment variable or default to 8081
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Define routes
	http.HandleFunc("/verse", verseHandler)
	http.HandleFunc("/health", healthHandler)

	// Start the server
	log.Printf("Starting verse extractor service on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
