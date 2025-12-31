package utils

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"swarmtest/internal/models"
)

// SharedBrowserPool is the global instance
var SharedBrowserPool *BrowserPool

// BrowserPool manages a shared Chrome instance
type BrowserPool struct {
	allocCtx context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
}

// NewBrowserPool creates a new browser pool
func NewBrowserPool(headless bool) (*BrowserPool, error) {
	// Check if chrome is available
	path, err := exec.LookPath("google-chrome")
	if err != nil {
		path, err = exec.LookPath("chromium")
		if err != nil {
			return nil, fmt.Errorf("chrome not found")
		}
	}
	log.Printf("Found chrome at %s", path)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Start a shared browser instance
	ctx, _ := chromedp.NewContext(allocCtx)
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start browser: %w", err)
	}

	return &BrowserPool{
		allocCtx: allocCtx,
		cancel:   cancel,
	}, nil
}

// GetContext creates a new tab context
func (p *BrowserPool) GetContext() (context.Context, context.CancelFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return chromedp.NewContext(p.allocCtx)
}

// Close shuts down the browser
func (p *BrowserPool) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}

// BrowserExecutor executes actions in a browser
type BrowserExecutor struct {
	pool *BrowserPool
	ctx  context.Context
	cancel context.CancelFunc
}

// NewBrowserExecutor creates a new executor for an agent
func NewBrowserExecutor(pool *BrowserPool) *BrowserExecutor {
	ctx, cancel := pool.GetContext()
	return &BrowserExecutor{
		pool:   pool,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Close cleans up the tab
func (e *BrowserExecutor) Close() {
	if e.cancel != nil {
		e.cancel()
	}
}

// ExecuteAction executes an action
func (e *BrowserExecutor) ExecuteAction(ctx context.Context, action models.GeminiDecisionResponse, currentURL string) ExecuteActionResult {
	// Use executor's context directly - no timeout wrapper to avoid cancellation issues
	var htmlContent string
	var newURL string
	
	switch action.Action {
	case "visit":
		if err := chromedp.Run(e.ctx,
			chromedp.Navigate(currentURL),
			chromedp.WaitReady("body"),
			chromedp.OuterHTML("html", &htmlContent),
			chromedp.Location(&newURL),
		); err != nil {
			return ExecuteActionResult{Error: err}
		}

	case "click":
		if err := chromedp.Run(e.ctx,
			chromedp.Click(action.Selector, chromedp.NodeVisible),
			chromedp.WaitReady("body"),
			chromedp.Sleep(1*time.Second), // Wait for hydration/animations
			chromedp.OuterHTML("html", &htmlContent),
			chromedp.Location(&newURL),
		); err != nil {
			return ExecuteActionResult{Error: err}
		}

	case "type":
		if err := chromedp.Run(e.ctx,
			chromedp.SendKeys(action.Selector, action.TextInput, chromedp.NodeVisible),
			chromedp.OuterHTML("html", &htmlContent),
			chromedp.Location(&newURL),
		); err != nil {
			return ExecuteActionResult{Error: err}
		}

	case "wait":
		if err := chromedp.Run(e.ctx,
			chromedp.Sleep(2*time.Second),
			chromedp.OuterHTML("html", &htmlContent),
			chromedp.Location(&newURL),
		); err != nil {
			return ExecuteActionResult{Error: err}
		}
		
	default:
		// Fallback for getting status
		if err := chromedp.Run(e.ctx,
			chromedp.OuterHTML("html", &htmlContent),
			chromedp.Location(&newURL),
		); err != nil {
			return ExecuteActionResult{Error: err}
		}
	}

	return ExecuteActionResult{
		HTML:   htmlContent,
		NewURL: newURL,
		// StatusCode is hard to get in simple chromedp, assuming 200 if successful
		StatusCode: 200, 
	}
}

// CaptureDOM captures current DOM state
func (e *BrowserExecutor) CaptureDOM(ctx context.Context) (string, string, error) {
	var htmlContent, urlStr string
	err := chromedp.Run(e.ctx,
		chromedp.OuterHTML("html", &htmlContent),
		chromedp.Location(&urlStr),
	)
	if err != nil {
		log.Printf("CaptureDOM failed: %v", err)
		return "", "", err
	}
	log.Printf("CaptureDOM succeeded, URL: %s, HTML length: %d", urlStr, len(htmlContent))
	return htmlContent, urlStr, nil
}

// GetInteractableElements returns interactive nodes (simplified)
func (e *BrowserExecutor) GetInteractableElements(ctx context.Context) ([]*cdp.Node, error) {
	var nodes []*cdp.Node
	err := chromedp.Run(e.ctx,
		chromedp.Nodes("a, button, input, select, textarea", &nodes, chromedp.ByQueryAll),
	)
	return nodes, err
}
