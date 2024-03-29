package urlparser

import (
	"io"
	"net/mail"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type URL struct {
	Href    string
	Content string
	Images  []Image
	Emails  []string
}

type Image struct {
	Src string
	Alt string
}

func ParseHTML(r io.Reader) ([]URL, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return parseNodes(doc), nil
}

func parseNodes(n *html.Node) []URL {
	var urls []URL
	if isAnchorNode(n) {
		urls = append(urls, parseAnchorNode(n))
	}

	urls = append(urls, parseChildNodes(n)...)

	return urls
}

func parseAnchorNode(n *html.Node) URL {
	url := URL{
		Href:    getHref(n),
		Content: getText(n),
		Images:  getImages(n),
		Emails:  getEmails(n),
	}

	return url
}

func parseChildNodes(n *html.Node) []URL {
	var urls []URL
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		urls = append(urls, parseNodes(c)...)
	}

	return urls
}

func isAnchorNode(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "a"
}

func getHref(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}

	return ""
}

func getText(n *html.Node) string {
	var text strings.Builder

	switch n.Type {
	case html.TextNode:
		text.WriteString(n.Data)
	case html.ElementNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			text.WriteString(getText(c))
		}
	}

	return strings.TrimSpace(text.String()) //ignore newline atm
}

func getImages(n *html.Node) []Image {
	var images []Image
	if n.Type == html.ElementNode && n.Data == "img" {
		img := parseImage(n)
		images = append(images, img)
	}

	images = append(images, parseChildImages(n)...)

	return images
}

func parseChildImages(n *html.Node) []Image {
	var images []Image
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		images = append(images, getImages(c)...)
	}

	return images
}

func parseImage(n *html.Node) Image {
	var img Image
	for _, attr := range n.Attr {
		if attr.Key == "src" {
			img.Src = attr.Val
		} else if attr.Key == "alt" {
			img.Alt = attr.Val
		}
	}

	return img
}

func getEmails(n *html.Node) []string {
	var emails []string
	if isAnchorNode(n) {
		emails = parseEmailLink(n)
	}

	text := getText(n)
	emailMatches := extractEmails(text)
	emails = append(emails, emailMatches...)

	emails = append(emails, parseChildEmails(n)...)

	return emails
}

func parseChildEmails(n *html.Node) []string {
	var emails []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		emails = append(emails, getEmails(c)...)
	}

	return emails
}

func parseEmailLink(n *html.Node) []string {
	var emails []string
	for _, attr := range n.Attr {
		if attr.Key == "href" && strings.HasPrefix(attr.Val, "mailto:") {
			email := strings.TrimPrefix(attr.Val, "mailto:")
			emails = append(emails, email)
			break
		}
	}

	return emails
}

func extractEmails(text string) []string {
	emailRegex := regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)
	emailMatches := emailRegex.FindAllString(text, -1)

	var validEmails []string
	for _, email := range emailMatches {
		_, err := mail.ParseAddress(email)
		if err == nil {
			validEmails = append(validEmails, email)
		}
	}

	return validEmails
}
