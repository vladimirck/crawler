package main

import (
	"errors"
	"io"
	"slices"

	//"io"
	"net/url"
	"strings"

	"fmt"
	"net/http"

	"golang.org/x/net/html"
)

// normalizeURL takes a URL string and returns a normalized version
// suitable for comparison. It converts host to lowercase,
// removes default ports (80 for http, 443 for https), removes
// fragments, removes trailing slashes from the path, and removes the scheme.
// The final format is host/path (or just host if path is empty).
// NOTE: This is an updated implementation based on the requirements.
func normalizeURL(rawURL string) string {
	// Handle URLs starting with "//" by assuming http
	if strings.HasPrefix(rawURL, "//") {
		rawURL = "http:" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" { // Scheme isn't strictly needed anymore for the output format
		// Return original or empty string for invalid/unparsable URLs
		// Returning original for simplicity here. Consider returning "" or error
		// if the input cannot be meaningfully normalized to host/path.
		// Let's try to extract host/path even if scheme is missing
		if u != nil && u.Host != "" {
			host := strings.ToLower(u.Host)
			path := u.Path
			if len(path) > 1 && strings.HasSuffix(path, "/") {
				path = strings.TrimSuffix(path, "/")
			}
			if path == "/" || path == "" {
				return host
			}
			return host + path // Path will start with /
		}
		return rawURL // Fallback for truly unparsable input
	}

	// 1. Lowercase host (scheme is not part of the output)
	// url.Parse already lowercases host, but being explicit is fine.
	host := strings.ToLower(u.Host)

	// 2. Remove default ports from the host string
	port := u.Port()
	scheme := strings.ToLower(u.Scheme) // Need scheme to check default ports
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		hostname := u.Hostname() // Get host without port
		host = hostname
	} else if port != "" {
		// Ensure non-default ports are kept in the host string
		host = u.Host // Use the original Host which includes the port
	}

	// 3. Fragment is ignored as it's not part of host/path output

	// 4. Normalize path: remove trailing slash, unless path is just "/"
	path := u.Path
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	// If the path is empty or just "/", return only the host
	if path == "/" || path == "" {
		return host
	}

	// Return host + path. Note: u.Path includes the leading slash.
	return strings.ToLower(host + path)
}

func getLinks(node *html.Node) []string {
	//fmt.Printf("node type: %v, node data: %v\n", node.Type, node.Data)
	links := []string{}
	if node.Type == html.ElementNode && node.Data == "a" {
		//fmt.Println("Econtré uno!")
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				links = append(links, attr.Val)
			}
		}
	}

	for child := range node.ChildNodes() {
		//fmt.Printf("***childs nodes**** type: %v, data: %v\n", child.Type, child.Data)
		child_links := getLinks(child)
		if len(child_links) > 0 {
			links = slices.Concat(links, child_links)
		}
	}

	return links
}

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	//fmt.Println("I am here")

	reader := strings.NewReader(htmlBody)
	htmlNode, err := html.Parse(reader)
	baseError := false
	linkError := false

	if err != nil {
		fmt.Println("Error, no se pudo parsear")
		return []string{}, errors.New("the Documents could not be parsed")
	}

	urls := getLinks(htmlNode)
	baseURL, err := url.Parse(rawBaseURL)

	if err != nil {
		//fmt.Printf("error parsing base link %s: %v\n", rawBaseURL, err)
		baseError = true
	}

	resultURL := []string{}

	for _, link := range urls {
		u, err := url.Parse(link)
		if err != nil {
			fmt.Printf("error parsing link %s: %v\n", link, err)
			linkError = true
		}

		//fmt.Printf("link: %v\n\t\t->Host: %v\n\t\t->Path: %v\n\n", link, u.Host, u.Path)

		if u.Host == "" {
			if baseURL != nil {
				resultURL = append(resultURL, baseURL.ResolveReference(u).String())
			}
		} else {
			resultURL = append(resultURL, u.String())
		}
	}

	var errLink error = nil
	if baseError {
		errLink = errors.New("the base link was not valid")
	} else if linkError {
		errLink = errors.New("the some links were not valid")
	}

	return resultURL, errLink
}

func getHTML(rawURL string) (string, error) {
	res, err := http.Get(rawURL)

	if err != nil {
		return "", err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", errors.New(res.Status)
	}

	contentType := res.Header.Get("Content-Type")

	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		return "", errors.New("invalid content type\n")
	}

	contentHTML, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return "", errors.New("error decoding the body\n")
	}

	return string(contentHTML), nil
}

func sameDomain(baseURL, otherURL string) bool {

	base, err := url.Parse(strings.ToLower(baseURL))
	if err != nil {
		return false
	}

	other, err := url.Parse(strings.ToLower(otherURL))
	if err != nil {
		return false
	}

	if base.Hostname() == "" || other.Hostname() == "" {
		return false
	}

	if base.Hostname() == other.Hostname() {
		return true
	}

	return false
}

func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int) {

	if !sameDomain(rawBaseURL, rawCurrentURL) {
		return
	}

	normURL := normalizeURL(rawCurrentURL)

	if _, ok := pages[normURL]; ok {
		pages[normURL]++
		return
	} else {
		pages[normURL] = 1
	}

	fmt.Printf("Entering at URL %s\n\n", normURL)
	htmlText, err := getHTML(rawCurrentURL)

	if err != nil {
		fmt.Printf("The URL %s no responding: %v\n", normURL, err)
		return
	}

	allURLs, err := getURLsFromHTML(htmlText, rawBaseURL)

	if len(allURLs) == 0 {
		fmt.Printf("%v", err)
		return
	}

	for _, link := range allURLs {
		crawlPage(rawBaseURL, link, pages)
	}
}
