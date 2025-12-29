package main

import (
	"context"

	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// HTTPClientFactory creates a new HTTP client for an agent
type HTTPClientFactory func() *http.Client

// NewHTTPClientFactory creates an HTTP client with cookie support
func NewHTTPClientFactory() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Printf("Failed to create cookie jar: %v", err)
		return &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
}

// ExecuteAction executes a single action on a webpage
type ActionExecutor struct {
	client  *http.Client
	parser  *HTMLParser
	baseURL *url.URL
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(client *http.Client, baseURL string) (*ActionExecutor, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &ActionExecutor{
		client:  client,
		parser:  NewHTMLParser(),
		baseURL: parsedURL,
	}, nil
}

// ExecuteActionResult is the result of executing an action
type ExecuteActionResult struct {
	HTML       string
	NewURL     string
	StatusCode int
	Error      error
}

// ExecuteAction executes an action and returns the resulting HTML
func (e *ActionExecutor) ExecuteAction(ctx context.Context, action GeminiDecisionResponse, currentURL string) ExecuteActionResult {


	switch action.Action {
	case "click":
		return e.executeClick(ctx, action, currentURL)
	case "type":
		return e.executeType(ctx, action, currentURL)
	case "wait":
		return e.executeWait()
	case "go_back":
		return ExecuteActionResult{
			Error: fmt.Errorf("go_back should be handled by agent, not executor"),
		}
	default:
		return ExecuteActionResult{
			Error: fmt.Errorf("unknown action: %s", action.Action),
		}
	}
}

// executeClick executes a click action
func (e *ActionExecutor) executeClick(ctx context.Context, action GeminiDecisionResponse, currentURL string) ExecuteActionResult {
	// Fetch current page
	resp, err := e.fetchWithRetry(ctx, currentURL)
	if err != nil {
		return ExecuteActionResult{Error: err}
	}
	defer resp.Body.Close()

	// Parse to find element
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ExecuteActionResult{Error: fmt.Errorf("parse HTML: %w", err)}
	}

	// Try to find the element and its href
	element := doc.Find(action.Selector)
	if element.Length() == 0 {
		return ExecuteActionResult{Error: fmt.Errorf("element not found: %s", action.Selector)}
	}

	// Check if it's a link with href
	if href, exists := element.Attr("href"); exists && href != "" {
		// Resolve URL
		targetURL, err := e.resolveURL(currentURL, href)
		if err != nil {
			return ExecuteActionResult{Error: fmt.Errorf("resolve URL: %w", err)}
		}

		// Follow the link
		linkResp, err := e.fetchWithRetry(ctx, targetURL)
		if err != nil {
			return ExecuteActionResult{Error: err}
		}
		defer linkResp.Body.Close()

		bodyBytes, _ := io.ReadAll(linkResp.Body)
		return ExecuteActionResult{
			HTML:       string(bodyBytes),
			NewURL:     linkResp.Request.URL.String(),
			StatusCode: linkResp.StatusCode,
		}
	}

	// Check if it's a button within a form
	form := element.Closest("form")
	if form.Length() > 0 {
		return e.submitForm(ctx, form, currentURL, action)
	}

	// Try clicking a button (might need to follow onclick, but for HTTP-only we do best effort)
	return ExecuteActionResult{
		Error: fmt.Errorf("click action not supported for element: %s", action.Selector),
	}
}

// executeType executes a type action
func (e *ActionExecutor) executeType(ctx context.Context, action GeminiDecisionResponse, currentURL string) ExecuteActionResult {
	// Fetch current page
	resp, err := e.fetchWithRetry(ctx, currentURL)
	if err != nil {
		return ExecuteActionResult{Error: err}
	}
	defer resp.Body.Close()

	// Parse to find input
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ExecuteActionResult{Error: fmt.Errorf("parse HTML: %w", err)}
	}

	input := doc.Find(action.Selector)
	if input.Length() == 0 {
		return ExecuteActionResult{Error: fmt.Errorf("input not found: %s", action.Selector)}
	}

	// Find nearest form
	form := input.Closest("form")
	if form.Length() == 0 {
		return ExecuteActionResult{Error: fmt.Errorf("no form found for input: %s", action.Selector)}
	}

	// Submit form with the input value
	return e.submitForm(ctx, form, currentURL, action)
}

// executeWait executes a wait action
func (e *ActionExecutor) executeWait() ExecuteActionResult {
	// For HTTP-only testing, wait means just pause briefly
	time.Sleep(time.Second)

	// Return error so agent knows to re-fetch current page
	return ExecuteActionResult{
		Error: fmt.Errorf("wait: agent should refetch current page"),
	}
}

// submitForm submits a form
func (e *ActionExecutor) submitForm(ctx context.Context, form *goquery.Selection, currentURL string, action GeminiDecisionResponse) ExecuteActionResult {
	// Get form action
	actionURL, _ := form.Attr("action")
	method, _ := form.Attr("method")
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	// Resolve action URL
	targetURL, err := e.resolveURL(currentURL, actionURL)
	if err != nil {
		return ExecuteActionResult{Error: fmt.Errorf("resolve form action: %w", err)}
	}

	// Build form data
	formData := url.Values{}

	// Collect all form inputs
	form.Find("input, select, textarea").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		if name == "" {
			return
		}

		tag := s.Get(0).Data
		inputType, _ := s.Attr("type")

		// Skip submit buttons (unless it's the clicked one)
		if inputType == "submit" || inputType == "button" {
			return
		}

		var value string
		if action.Selector != "" && generateSelectorForNode(s.Get(0)) == action.Selector {
			value = action.TextInput
		} else {
			// Use default value
			value, _ = s.Attr("value")
			if tag == "select" {
				// Use first option value
				s.Find("option").First().Each(func(i int, opt *goquery.Selection) {
					if val, exists := opt.Attr("value"); exists && val != "" {
						value = val
					} else {
						value = strings.TrimSpace(opt.Text())
					}
				})
			}
		}

		// Skip checkboxes and radio buttons that aren't checked
		if (inputType == "checkbox" || inputType == "radio") && !s.Is(":checked") {
			return
		}

		formData.Set(name, value)
	})

	// Submit form
	var req *http.Request
	if method == "POST" {
		req, err = http.NewRequestWithContext(ctx, "POST", targetURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return ExecuteActionResult{Error: fmt.Errorf("create POST request: %w", err)}
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequestWithContext(ctx, "GET", targetURL+"?"+formData.Encode(), nil)
		if err != nil {
			return ExecuteActionResult{Error: fmt.Errorf("create GET request: %w", err)}
		}
	}

	// Set headers
	req.Header.Set("User-Agent", "SwarmTest/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")

	// Execute
	resp, err := e.client.Do(req)
	if err != nil {
		return ExecuteActionResult{Error: fmt.Errorf("submit form: %w", err)}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	return ExecuteActionResult{
		HTML:       string(bodyBytes),
		NewURL:     resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
	}
}

// fetchWithRetry fetches a URL with retry logic
func (e *ActionExecutor) fetchWithRetry(ctx context.Context, urlStr string) (*http.Response, error) {
	var lastErr error
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "SwarmTest/1.0")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")

		resp, err := e.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// Check for rate limiting or server errors
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server returned %d", resp.StatusCode)
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// resolveURL resolves a potentially relative URL against a base URL
func (e *ActionExecutor) resolveURL(base, rel string) (string, error) {
	if rel == "" || rel == "#" {
		return base, nil
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	relURL, err := url.Parse(rel)
	if err != nil {
		return "", err
	}

	return baseURL.ResolveReference(relURL).String(), nil
}

// generateSelectorForNode generates a selector for a node (simplified version)
func generateSelectorForNode(node *html.Node) string {
	if node == nil {
		return ""
	}
	// Simplified - in production you'd match the logic in html_parser.go
	if node.Parent != nil {
		idx := 1
		for c := node.Parent.FirstChild; c != nil; c = c.NextSibling {
			if c == node {
				return node.Data + ":nth-child(" + fmt.Sprint(idx) + ")"
			}
			if c.Type == html.ElementNode {
				idx++
			}
		}
	}
	return node.Data
}
