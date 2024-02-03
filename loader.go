package restify

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const HttpRequestTimeout = time.Second * 60

type RequestConfig func(*http.Request)

// WithHeaders configures additional headers in the request used in LoadContent
func WithHeaders(headers map[string]string) RequestConfig {
	return func(request *http.Request) {
		for k, v := range headers {
			request.Header.Set(k, v)
		}
	}
}

// LoadFile retrieves the HTML content from the given file URL.
func LoadBuffer(buffer []byte) (*html.Node, error) {
	root, err := html.Parse(strings.NewReader(string(buffer)))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse buffer: %w", err)
	}

	return root, nil
}

func LoadReader(reader io.Reader) (*html.Node, error) {
	root, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse reader: %w", err)
	}

	return root, nil
}

func LoadFile(url *url.URL, userAgent string, configs ...RequestConfig) (*html.Node, error) {

	// open the file as a io.reader
	filePointer, err := os.Open(url.Path)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult

	root, err := html.Parse(filePointer)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse file: %w", err)
	}

	return root, nil
}

// LoadContent retrieves the HTML content from the given url.
// The userAgent is optional, but if provided should conform with https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
func LoadContent(url *url.URL, userAgent string, configs ...RequestConfig) (*html.Node, error) {
	if url.Scheme == "file" {
		return LoadFile(url, userAgent, configs...)
	}

	request, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to request: %w", err)
	}

	request.Header.Set("accept", "*/*")
	if userAgent != "" {
		request.Header.Set("user-agent", userAgent)
	}
	for _, config := range configs {
		config(request)
	}

	http.DefaultClient.Timeout = HttpRequestTimeout
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve response: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse response body: %w", err)
	}

	return root, nil
}

// FindSubsetById locates the HTML node within the given root that has an id attribute of given value.
// If the node is not found, then ok will be false.
func FindSubsetById(root *html.Node, id string) (n *html.Node, ok bool) {
	return scrape.Find(root, scrape.ById(id))
}

// FindSubsetByClass locates the HTML nodes with the given root that have the given className.
func FindSubsetByClass(root *html.Node, className string) []*html.Node {
	return scrape.FindAll(root, scrape.ByClass(className))
}

// FindSubsetByAttributeName retrieves the HTML nodes that have the requested
// attribute, regardless of their values.
func FindSubsetByAttributeName(root *html.Node, attribute string) []*html.Node {
	return FindSubsetByAttributeNameValue(root, attribute, "")
}

// FindSubsetByAttributeNameValue retrieves the HTML nodes that have the requested attribute with a specific value.
func FindSubsetByAttributeNameValue(root *html.Node, attribute string, value string) []*html.Node {
	return scrape.FindAll(root, matchByAttribute(attribute, value))
}

// FindSubsetByTagName retrieves the HTML nodes with the given tagName
func FindSubsetByTagName(root *html.Node, tagName string) []*html.Node {
	return scrape.FindAll(root, scrape.ByTag(atom.Lookup([]byte(tagName))))
}

func matchByAttribute(key, value string) scrape.Matcher {
	return func(node *html.Node) bool {
		if node.Type == html.ElementNode {
			result := scrape.Attr(node, key)
			if result != "" && (value == "" || value == result) {
				return true
			}
		}
		return false
	}
}
