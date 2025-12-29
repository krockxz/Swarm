package main

import (
	"io"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// HTMLParser parses HTML and extracts interactive elements
type HTMLParser struct{}

// NewHTMLParser creates a new HTML parser
func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

// ParseHTML parses HTML content and returns a StrippedPage
func (p *HTMLParser) ParseHTML(url string, htmlContent io.Reader) (*StrippedPage, error) {
	doc, err := goquery.NewDocumentFromReader(htmlContent)
	if err != nil {
		return nil, err
	}

	page := &StrippedPage{
		URL:         url,
		Title:       doc.Find("title").First().Text(),
		InteractiveElements: []Element{},
		Timestamp:   time.Now(),
	}

	// Extract meta description
	if desc, exists := doc.Find("meta[name='description']").Attr("content"); exists {
		page.Description = truncateString(desc, 500)
	} else {
		// Try og:description
		if desc, exists := doc.Find("meta[property='og:description']").Attr("content"); exists {
			page.Description = truncateString(desc, 500)
		}
	}

	// Extract interactive elements
	page.InteractiveElements = p.extractElements(doc)

	return page, nil
}

// extractElements extracts all interactive elements from the page
func (p *HTMLParser) extractElements(doc *goquery.Document) []Element {
	elements := []Element{}
	elementID := 0

	// Links and buttons with href or onclick
	doc.Find("a, button, [onclick], [role='button']").Each(func(i int, s *goquery.Selection) {
		selector := generateSelector(s)

		// Check if it's a link
		if tag := s.Get(0).Data; tag == "a" {
			href, _ := s.Attr("href")
			text := strings.TrimSpace(s.Text())

			if href != "" || s.HasClass("btn") || s.HasClass("button") {
				elements = append(elements, Element{
					ID:       generateElementID(elementID),
					Type:     "link",
					Text:     truncateString(text, 100),
					Selector: selector,
					Href:     href,
				})
				elementID++
			}
		} else {
			// Button
			text := strings.TrimSpace(s.Text())
			if text == "" {
				text, _ = s.Attr("value")
				text, _ = s.Attr("aria-label")
			}

			elements = append(elements, Element{
				ID:       generateElementID(elementID),
				Type:     "button",
				Text:     truncateString(text, 100),
				Selector: selector,
			})
			elementID++
		}
	})

	// Input fields
	doc.Find("input, textarea, select").Each(func(i int, s *goquery.Selection) {
		selector := generateSelector(s)

		inputType, _ := s.Attr("type")
		if inputType == "" {
			inputType = "text"
		}

		// Skip hidden inputs
		if inputType == "hidden" {
			return
		}

		name, _ := s.Attr("name")
		placeholder, _ := s.Attr("placeholder")

		elements = append(elements, Element{
			ID:          generateElementID(elementID),
			Type:        "input",
			Selector:    selector,
			Name:        name,
			Placeholder: placeholder,
			InputType:   inputType,
		})
		elementID++
	})

	// Forms
	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		selector := generateSelector(s)
		action, _ := s.Attr("action")
		method, _ := s.Attr("method")
		if method == "" {
			method = "GET"
		}

		elements = append(elements, Element{
			ID:       generateElementID(elementID),
			Type:     "form",
			Text:     method + " " + action,
			Selector: selector,
		})
		elementID++
	})

	return elements
}

// generateSelector generates a simple CSS selector for an element
func generateSelector(s *goquery.Selection) string {
	node := s.Get(0)
	if node == nil {
		return ""
	}

	var parts []string

	// Walk up the tree to build a selector
	for n := node; n != nil; n = n.Parent {
		if n.Type == html.ElementNode {
			tag := n.Data
			var classPart, idPart string

			// Add ID if present
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val != "" {
					idPart = "#" + attr.Val
					break
				}
			}

			// Add class if no ID
			if idPart == "" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && attr.Val != "" {
						classes := strings.Fields(attr.Val)
						if len(classes) > 0 {
							classPart = "." + strings.Join(classes, ".")
							break
						}
					}
				}
			}

			// Build the part
			part := tag
			if idPart != "" {
				part += idPart
			} else if classPart != "" {
				part += classPart
			} else {
				// Use nth-child if no class or id
				if n.Parent != nil {
					idx := 1
					for c := n.Parent.FirstChild; c != nil; c = c.NextSibling {
						if c == n {
							part = tag + ":nth-child(" + string(rune('0'+idx)) + ")"
							break
						}
						if c.Type == html.ElementNode {
							idx++
						}
					}
				}
			}

			parts = append([]string{part}, parts...)

			// If we have an ID, we can stop
			if idPart != "" {
				break
			}
		}
	}

	if len(parts) > 3 {
		parts = parts[len(parts)-3:]
	}

	return strings.Join(parts, " ")
}

// generateElementID generates a unique element ID
func generateElementID(num int) string {
	return "elem_" + string(rune('a'+num%26))
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
