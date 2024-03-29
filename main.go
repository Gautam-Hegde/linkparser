package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	urlparser "parser/urlParser"
	"strings"
)

func main() {

	http.HandleFunc("/parse", parseHandler)

	fmt.Println("Server started on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	link := strings.TrimSpace(string(body))
	if link == "" {
		http.Error(w, "Empty link provided", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(link)
	if err != nil {
		http.Error(w, "Invalid URL provided", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(link)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching URL: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Received non-200 status code: %d", resp.StatusCode), http.StatusInternalServerError)
		return
	}

	urls, err := urlparser.ParseHTML(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing HTML: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	cleanedURLs := make([]map[string]interface{}, 0)
	for _, url := range urls {
		cleanedURL := make(map[string]interface{})
		if url.Href != "" {
			if !strings.HasPrefix(url.Href, "http") { // Check if Href is relative
				domain := parsedURL.Scheme + "://" + parsedURL.Host
				if strings.HasPrefix(url.Href, "/") { // Check if Href starts with a slash
					url.Href = domain + url.Href
				} else {
					url.Href = domain + "/" + url.Href
				}
			}
			cleanedURL["Href"] = url.Href
		}
		if url.Content != "" {
			cleanedURL["Content"] = url.Content
		}
		if len(url.Images) > 0 {
			cleanedURL["Images"] = url.Images
		}
		if len(url.Emails) > 0 {
			cleanedURL["Emails"] = url.Emails
		}
		if len(cleanedURL) > 0 {
			cleanedURLs = append(cleanedURLs, cleanedURL)
		}
	}

	jsonResponse, err := json.Marshal(cleanedURLs)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
