package main // Use the actual package name where normalizeURL resides

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// TestNormalizeURL tests the normalizeURL function with various inputs.
func TestNormalizeURL(t *testing.T) {
	// Define test cases using a table
	testCases := []struct {
		name     string // Name of the test case
		input    string // Input URL string
		expected string // Expected normalized URL string (host/path format)
	}{
		// User specified cases
		{
			name:     "User Case 1: HTTPS with trailing slash",
			input:    "https://blog.boot.dev/path/",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "User Case 2: HTTPS without trailing slash",
			input:    "https://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "User Case 3: HTTP with trailing slash",
			input:    "http://blog.boot.dev/path/",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "User Case 4: HTTP without trailing slash",
			input:    "http://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},

		// Original cases adapted to new format
		{
			name:     "Simple HTTP",
			input:    "http://example.com",
			expected: "example.com", // Path is empty
		},
		{
			name:     "Simple HTTPS with slash",
			input:    "https://example.com/",
			expected: "example.com", // Path becomes empty after removing slash
		},
		{
			name:     "Uppercase Scheme",
			input:    "HTTP://example.com",
			expected: "example.com",
		},
		{
			name:     "Uppercase Host",
			input:    "http://EXAMPLE.COM",
			expected: "example.com",
		},
		{
			name:     "HTTP with default port",
			input:    "http://example.com:80/path",
			expected: "example.com/path",
		},
		{
			name:     "HTTPS with default port and trailing slash",
			input:    "https://example.com:443/path/",
			expected: "example.com/path",
		},
		{
			name:     "HTTP with non-default port",
			input:    "http://example.com:8080",
			expected: "example.com:8080",
		},
		{
			name:     "HTTPS with non-default port and path",
			input:    "https://example.com:8443/path/",
			expected: "example.com:8443/path",
		},
		{
			name:     "Path without trailing slash",
			input:    "http://example.com/path/to/resource",
			expected: "example.com/path/to/resource",
		},
		{
			name:     "Path with trailing slash",
			input:    "http://example.com/path/to/resource/",
			expected: "example.com/path/to/resource",
		},
		{
			name:     "Root path only",
			input:    "http://example.com/",
			expected: "example.com",
		},
		{
			name:     "URL with fragment",
			input:    "http://example.com/path#section1",
			expected: "example.com/path", // Fragment ignored
		},
		{
			name:     "URL with query parameters",
			input:    "http://example.com/path?a=1&b=2",
			expected: "example.com/path", // Query params ignored
		},
		{
			name:     "URL with query and fragment",
			input:    "http://example.com/path?a=1#section",
			expected: "example.com/path", // Both ignored
		},
		{
			name:     "URL with trailing slash, query, and fragment",
			input:    "https://example.com:443/path/?a=1#section",
			expected: "example.com/path", // Default port removed, trailing slash removed, query/fragment ignored
		},
		{
			name:     "Scheme-relative URL",
			input:    "//Example.com/path/",
			expected: "example.com/path", // Assumes http for port check (doesn't matter here), removes trailing slash
		},
		{
			name:     "URL with username/password",
			input:    "https://user:pass@example.com/path",
			expected: "example.com/path", // User info not preserved
		},
		{
			name:     "Path is only slash",
			input:    "http://example.com/",
			expected: "example.com", // Path becomes empty
		},
		// Add more test cases as needed
		// {
		//  name:     "Invalid URL",
		//  input:    "htp:/invalid-url",
		//  expected: "htp:/invalid-url", // Or "" depending on error handling
		// },
		{
			name:     "No Scheme, No Path",
			input:    "example.com",
			expected: "example.com", // Should parse host correctly
		},
		{
			name:     "No Scheme, With Path",
			input:    "example.com/path",
			expected: "example.com/path", // Should parse host and path
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			// Call the function under test
			actual := normalizeURL(tc.input)

			// Compare the actual result with the expected result
			if actual != tc.expected {
				t.Errorf("normalizeURL(%q)\n  got: %q\n want: %q", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestGetURLsFromHTML(t *testing.T) {
	testCases := []struct {
		name         string   // Nombre del caso de prueba
		htmlBody     string   // Cuerpo HTML de entrada
		baseURL      string   // URL base para resolver URLs relativas
		expectedURLs []string // URLs absolutas esperadas (ordenadas)
		expectError  bool     // Si se espera un error
	}{
		{
			name:         "Sin enlaces",
			htmlBody:     `<html><body><p>Sin enlaces aquí.</p></body></html>`,
			baseURL:      "https://example.com",
			expectedURLs: []string{},
			expectError:  false,
		},
		{
			name:         "Enlace absoluto simple",
			htmlBody:     `<html><body><a href="https://blog.boot.dev">Boot.dev Blog</a></body></html>`,
			baseURL:      "https://example.com",
			expectedURLs: []string{"https://blog.boot.dev"},
			expectError:  false,
		},
		{
			name:         "Enlace relativo simple",
			htmlBody:     `<html><body><a href="/path/to/page">Página Relativa</a></body></html>`,
			baseURL:      "https://example.com",
			expectedURLs: []string{"https://example.com/path/to/page"},
			expectError:  false,
		},
		{
			name: "Múltiples enlaces (absolutos y relativos)",
			htmlBody: `
				<html>
					<body>
						<a href="https://other.com/page1">Absoluto</a>
						<p>Algo de texto</p>
						<a href="/relative/page2">Relativo</a>
						<a href="https://example.com/another/path">Otro Absoluto</a>
						<a href="justafile.html">Archivo Relativo</a>
					</body>
				</html>`,
			baseURL: "https://example.com/base/", // Base URL con path
			expectedURLs: []string{
				"https://example.com/another/path",
				"https://example.com/base/justafile.html", // Relativo a /base/
				"https://example.com/relative/page2",      // Relativo a la raíz del host
				"https://other.com/page1",
			},
			expectError: false,
		},
		{
			name:         "Enlace relativo sin slash inicial",
			htmlBody:     `<html><body><a href="subfolder/page">Subcarpeta</a></body></html>`,
			baseURL:      "https://example.com/current/path/",
			expectedURLs: []string{"https://example.com/current/path/subfolder/page"},
			expectError:  false,
		},
		{
			name:         "Enlace relativo con ..",
			htmlBody:     `<html><body><a href="../otherfolder/page">Otra Carpeta</a></body></html>`,
			baseURL:      "https://example.com/current/path/",
			expectedURLs: []string{"https://example.com/current/otherfolder/page"},
			expectError:  false,
		},
		{
			name:         "Enlace a la raíz",
			htmlBody:     `<html><body><a href="/">Inicio</a></body></html>`,
			baseURL:      "https://example.com/current/path/",
			expectedURLs: []string{"https://example.com/"},
			expectError:  false,
		},
		{
			name:         "URL base inválida",
			htmlBody:     `<html><body><a href="/path">Relativo</a></body></html>`,
			baseURL:      ":invalid-url:",
			expectedURLs: []string{},
			expectError:  true,
		},
		{
			name:         "HTML vacío",
			htmlBody:     "",
			baseURL:      "https://example.com",
			expectedURLs: []string{},
			expectError:  false,
		},
		{
			name: "Ignorar fragmentos y queries (si la normalización no es parte de getURLs)",
			htmlBody: `<html><body>
						<a href="/path?query=1">Con Query</a>
						<a href="/path#fragment">Con Fragmento</a>
					   </body></html>`,
			baseURL: "https://example.com",
			// Asumiendo que getURLsFromHTML solo extrae y resuelve, sin normalizar quitando query/fragment
			expectedURLs: []string{
				"https://example.com/path#fragment",
				"https://example.com/path?query=1",
			},
			expectError: false,
		},
		{
			name:     "URL base con path sin slash final",
			htmlBody: `<html><body><a href="page.html">Página</a></body></html>`,
			baseURL:  "https://example.com/folder", // Sin slash final
			// El comportamiento estándar de url.Parse es resolver relativo al 'directorio' padre
			expectedURLs: []string{"https://example.com/page.html"},
			expectError:  false,
		},
		{
			name:         "URL base con path con slash final",
			htmlBody:     `<html><body><a href="page.html">Página</a></body></html>`,
			baseURL:      "https://example.com/folder/", // Con slash final
			expectedURLs: []string{"https://example.com/folder/page.html"},
			expectError:  false,
		},

		// Añadir más casos de prueba según sea necesario
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Llama a la función bajo prueba (Asegúrate de que esté disponible en el paquete)
			actualURLs, err := getURLsFromHTML(tc.htmlBody, tc.baseURL)

			// Verifica si se esperaba un error y si ocurrió (o no)
			if tc.expectError {
				if err == nil {
					t.Errorf("Se esperaba un error, pero no ocurrió ninguno.")
				}
				// Puedes añadir verificaciones más específicas sobre el tipo de error si es necesario
				return // No continuar si se esperaba un error
			}
			if err != nil {
				t.Errorf("No se esperaba un error, pero ocurrió: %v", err)
				return // No continuar si ocurrió un error inesperado
			}

			// Ordena ambos slices (actual y esperado) para una comparación consistente
			sort.Strings(actualURLs)
			sort.Strings(tc.expectedURLs)

			// Compara los resultados
			if !reflect.DeepEqual(actualURLs, tc.expectedURLs) {
				t.Errorf("getURLsFromHTML(%q, %q)\n  got: %v\n want: %v", tc.htmlBody, tc.baseURL, actualURLs, tc.expectedURLs)
			}
		})
	}
}

// Mock Handler Helper
func createMockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// --- Test Cases ---

func TestGetHTML_HappyPath(t *testing.T) {
	expectedHTML := "<html><body><h1>Hello</h1></body></html>"
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, expectedHTML)
	})
	defer server.Close()

	html, err := getHTML(server.URL)

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	// Trim potential trailing newline added by Fprintln
	if strings.TrimSpace(html) != expectedHTML {
		t.Errorf("Expected HTML '%s', but got '%s'", expectedHTML, html)
	}
}

func TestGetHTML_HappyPathWithCharset(t *testing.T) {
	expectedHTML := "<html><body><h1>Charset</h1></body></html>"
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		// Common variation
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, expectedHTML)
	})
	defer server.Close()

	html, err := getHTML(server.URL)

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if strings.TrimSpace(html) != expectedHTML {
		t.Errorf("Expected HTML '%s', but got '%s'", expectedHTML, html)
	}
}

func TestGetHTML_HappyPathCaseInsensitive(t *testing.T) {
	expectedHTML := "<html><body><h1>Case Test</h1></body></html>"
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		// Case variation
		w.Header().Set("Content-Type", "TEXT/HTML")
		fmt.Fprintln(w, expectedHTML)
	})
	defer server.Close()

	html, err := getHTML(server.URL)

	if err != nil {
		t.Fatalf("Expected no error for case-insensitive header, but got: %v", err)
	}
	if strings.TrimSpace(html) != expectedHTML {
		t.Errorf("Expected HTML '%s', but got '%s'", expectedHTML, html)
	}
}

func TestGetHTML_ErrorInvalidContentType_JSON(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"key": "value"}`)
	})
	defer server.Close()

	html, err := getHTML(server.URL)

	if err == nil {
		t.Fatalf("Expected an error for content type application/json, but got nil")
	}
	// Assuming the error message indicates the issue
	if !strings.Contains(strings.ToLower(err.Error()), "invalid content type") && !strings.Contains(strings.ToLower(err.Error()), "content-type") {
		t.Errorf("Expected error message to contain 'invalid content type', but got: %v", err)
	}
	if html != "" {
		t.Errorf("Expected empty HTML string on error, but got: %s", html)
	}
}

func TestGetHTML_ErrorInvalidContentType_PlainText(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "This is plain text.")
	})
	defer server.Close()

	html, err := getHTML(server.URL)

	if err == nil {
		t.Fatalf("Expected an error for content type text/plain, but got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid content type") && !strings.Contains(strings.ToLower(err.Error()), "content-type") {
		t.Errorf("Expected error message to contain 'invalid content type', but got: %v", err)
	}
	if html != "" {
		t.Errorf("Expected empty HTML string on error, but got: %s", html)
	}
}

func TestGetHTML_ErrorInvalidURL(t *testing.T) {
	invalidURLs := []string{
		"",                                   // Empty
		"htp://google.com",                   // Invalid scheme
		"just string",                        // Not a URL
		"://google.com",                      // Missing scheme
		"http://invalid url with spaces.com", // Spaces (needs encoding)
	}

	for _, url := range invalidURLs {
		t.Run(fmt.Sprintf("URL_%s", url), func(t *testing.T) {
			html, err := getHTML(url)
			if err == nil {
				t.Errorf("Expected an error for invalid URL '%s', but got nil", url)
			}
			// The specific error might vary (e.g., url.Parse error)
			// We just check that *an* error occurred.
			if html != "" {
				t.Errorf("Expected empty HTML string on error, but got: %s", html)
			}
		})
	}
}

func TestGetHTML_ErrorServerError(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // 500
		fmt.Fprintln(w, "Internal Server Error")
	})
	defer server.Close()

	html, err := getHTML(server.URL)

	if err == nil {
		t.Fatalf("Expected an error for 500 status code, but got nil")
	}
	// Check if error indicates non-2xx status
	if !strings.Contains(err.Error(), "status code") && !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error message to mention non-2xx status or 500, but got: %v", err)
	}
	if html != "" {
		t.Errorf("Expected empty HTML string on error, but got: %s", html)
	}
}

func TestGetHTML_ErrorNotFound(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound) // 404
		fmt.Fprintln(w, "Not Found")
	})
	defer server.Close()

	html, err := getHTML(server.URL)

	if err == nil {
		t.Fatalf("Expected an error for 404 status code, but got nil")
	}
	if !strings.Contains(err.Error(), "status code") && !strings.Contains(err.Error(), "404") {
		t.Errorf("Expected error message to mention non-2xx status or 404, but got: %v", err)
	}
	if html != "" {
		t.Errorf("Expected empty HTML string on error, but got: %s", html)
	}
}

// Note: Testing actual network errors like DNS resolution failure or connection
// refused is harder with httptest. These often require integration tests or
// more complex mocking of the network stack itself. A simple proxy might be
// trying a URL that's syntactically valid but guaranteed not to resolve or connect.
func TestGetHTML_ErrorNetwork(t *testing.T) {
	// This URL is syntactically valid but unlikely to be served, simulating a network issue.
	// Use with caution, as environment might affect this.
	unreachableURL := "http://127.0.0.1:9999/unreachable"
	// Or a non-resolvable domain: "http://domain.invalid/"

	html, err := getHTML(unreachableURL)

	if err == nil {
		t.Fatalf("Expected a network error for unreachable URL '%s', but got nil", unreachableURL)
	}
	// Network errors can vary (connection refused, context deadline exceeded, no such host)
	// We just expect *some* error.
	if html != "" {
		t.Errorf("Expected empty HTML string on network error, but got: %s", html)
	}
}

func TestSameDomain(t *testing.T) {
	testCases := []struct {
		name     string
		baseURL  string
		otherURL string
		expected bool
	}{
		// Happy Path - Same Domain
		{
			name:     "Exact Match HTTP",
			baseURL:  "http://example.com",
			otherURL: "http://example.com",
			expected: true,
		},
		{
			name:     "Exact Match HTTPS",
			baseURL:  "https://example.com",
			otherURL: "https://example.com",
			expected: true,
		},
		{
			name:     "Same Domain Different Path",
			baseURL:  "http://example.com/path1",
			otherURL: "http://example.com/path2",
			expected: true,
		},
		{
			name:     "Same Domain Different Scheme",
			baseURL:  "http://example.com",
			otherURL: "https://example.com",
			expected: true,
		},
		{
			name:     "Same Domain Different Port",
			baseURL:  "http://example.com:8080",
			otherURL: "http://example.com:9090",
			expected: true, // Assuming port is ignored
		},
		{
			name:     "Same Domain Base Omits Standard Port",
			baseURL:  "http://example.com",
			otherURL: "http://example.com:80",
			expected: true, // Assuming port is ignored
		},
		{
			name:     "Same Domain Other Omits Standard Port",
			baseURL:  "https://example.com:443",
			otherURL: "https://example.com",
			expected: true, // Assuming port is ignored
		},
		{
			name:     "Same Domain Different Query",
			baseURL:  "http://example.com?q=1",
			otherURL: "http://example.com?q=2",
			expected: true,
		},
		{
			name:     "Same Domain Different Fragment",
			baseURL:  "http://example.com#frag1",
			otherURL: "http://example.com#frag2",
			expected: true,
		},
		{
			name:     "Same Domain Different UserInfo",
			baseURL:  "http://user1@example.com",
			otherURL: "http://user2@example.com",
			expected: true, // Assuming userinfo ignored
		},
		{
			name:     "Same Domain Case Insensitive",
			baseURL:  "http://Example.com",
			otherURL: "http://example.com",
			expected: true,
		},
		{
			name:     "Same Domain www",
			baseURL:  "http://www.example.com",
			otherURL: "https://www.example.com/path",
			expected: true,
		},
		{
			name:     "Same Domain IP Address",
			baseURL:  "http://192.168.1.1/a",
			otherURL: "http://192.168.1.1:8080/b",
			expected: true,
		},

		// Different Domains
		{
			name:     "Different Domain Simple",
			baseURL:  "http://example.com",
			otherURL: "http://google.com",
			expected: false,
		},
		{
			name:     "Different TLD",
			baseURL:  "http://example.com",
			otherURL: "http://example.org",
			expected: false,
		},
		{
			name:     "Different Subdomain (www vs non-www)",
			baseURL:  "http://www.example.com",
			otherURL: "http://example.com",
			expected: false, // Assuming exact host match required
		},
		{
			name:     "Different Subdomain (other)",
			baseURL:  "http://app.example.com",
			otherURL: "http://api.example.com",
			expected: false,
		},
		{
			name:     "Different IP Address",
			baseURL:  "http://192.168.1.1",
			otherURL: "http://192.168.1.2",
			expected: false,
		},
		{
			name:     "IP vs Hostname",
			baseURL:  "http://127.0.0.1",
			otherURL: "http://localhost", // Resolve differently at parse time
			expected: false,
		},

		// Invalid / Edge Cases
		{
			name:     "Base URL Invalid",
			baseURL:  "://invalid-url",
			otherURL: "http://example.com",
			expected: false, // Cannot parse base host
		},
		{
			name:     "Other URL Invalid",
			baseURL:  "http://example.com",
			otherURL: "http://invalid host name.com", // space in host
			expected: false,                          // Cannot parse other host
		},
		{
			name:     "Both URLs Invalid",
			baseURL:  "://invalid-1",
			otherURL: "::invalid-2",
			expected: false,
		},
		{
			name:     "Base URL Empty",
			baseURL:  "",
			otherURL: "http://example.com",
			expected: false, // Cannot parse base host
		},
		{
			name:     "Other URL Empty",
			baseURL:  "http://example.com",
			otherURL: "",
			expected: false, // Cannot parse other host
		},
		{
			name:     "Both URLs Empty",
			baseURL:  "",
			otherURL: "",
			expected: false,
		},
		{
			name:     "Base URL Relative Path",
			baseURL:  "/path/only",
			otherURL: "http://example.com",
			expected: false, // Base has no host
		},
		{
			name:     "Other URL Relative Path",
			baseURL:  "http://example.com",
			otherURL: "/path/only",
			expected: false, // Other has no host
		},
		{
			name:     "Base URL Scheme Missing",
			baseURL:  "example.com", // url.Parse treats this as Path without Scheme/Host
			otherURL: "http://example.com",
			expected: false, // Base has no host
		},
		{
			name:     "Other URL Scheme Missing",
			baseURL:  "http://example.com",
			otherURL: "example.com", // url.Parse treats this as Path without Scheme/Host
			expected: false,         // Other has no host
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := sameDomain(tc.baseURL, tc.otherURL)
			if actual != tc.expected {
				t.Errorf("sameDomain(%q, %q) = %v; want %v", tc.baseURL, tc.otherURL, actual, tc.expected)
			}
		})
	}
}

// --- Mock Server Setup ---

// Helper to create HTML content easily (same as before)
func createHTML(title string, links []string) string {
	var linkTags []string
	for _, link := range links {
		linkTags = append(linkTags, fmt.Sprintf(`<a href="%s">%s</a>`, link, link))
	}
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><title>%s</title></head>
<body>
<h1>%s</h1>
%s
</body>
</html>
`, title, title, strings.Join(linkTags, "\n"))
}

// --- Test Case ---

func TestCrawlPage_InternalLinksOnly_SchemeLessNormalized(t *testing.T) {
	// Map to store HTML content for different paths
	pageContents := make(map[string]string)

	// --- Setup Mock Server ---
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Handle potential trailing slash inconsistency in requests
		if path != "/" && strings.HasSuffix(path, "/") {
			path = strings.TrimSuffix(path, "/")
		}

		content, exists := pageContents[path]
		if !exists {
			if !strings.Contains(path, ".") && path != "/favicon.ico" {
				t.Logf("Mock server request for undefined internal path: %s (returning empty)", path)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "<html><body>Empty page or Not Found</body></html>")
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, content)
	}))
	defer server.Close()

	// --- Define Expected Normalized URLs (keys in the final map) ---
	// Format: host:port/path (lowercase path, no trailing slash)
	expectedInternalKeys := make(map[string]bool)

	// Get base details AFTER server starts
	baseURL, _ := url.Parse(server.URL) // Still useful for getting the test Host+Port
	expectedHostPort := baseURL.Host    // e.g., 127.0.0.1:xxxx

	// Define expected keys based on the assumed scheme-less normalization
	expectedInternalKeys[expectedHostPort] = true          // Root page (no path)
	expectedInternalKeys[expectedHostPort+"/pagea"] = true // Page A
	expectedInternalKeys[expectedHostPort+"/pageb"] = true // Page B
	expectedInternalKeys[expectedHostPort+"/pagec"] = true // Page C (link found, even if page empty/404)

	// --- Define Page Content ---
	// Links should still be standard URLs; crawlPage/normalizeURL handle conversion
	pageContents["/"] = createHTML("Index", []string{
		"/pageA",              // -> host:port/pagea
		"pageB/",              // -> host:port/pageb
		server.URL + "/pageA", // -> host:port/pagea
		server.URL + "/",      // -> host:port
		"pageA",               // -> host:port/pagea
		"http://external.com", // External (ignore)
		"#fragment",           // Fragment only (should resolve to base -> host:port)
	})
	pageContents["/pageA"] = createHTML("Page A", []string{
		"../pageB",       // -> host:port/pageb
		server.URL + "/", // -> host:port
		"/pageA",         // -> host:port/pagea (test self-link normalization)
	})
	pageContents["/pageB"] = createHTML("Page B", []string{
		"http://another-external.org", // External (ignore)
		"pageC",                       // -> host:port/pagec
	})
	// pageC is not defined, simulating empty/404

	// --- Execute the Crawl ---
	pages := make(map[string]int)
	crawlPage(server.URL, server.URL, pages) // Start crawl from the base URL

	// --- Assertions ---
	foundInternalKeys := make(map[string]bool)
	foundMalformedKeys := make(map[string]bool) // Keys that don't seem to match expected format

	t.Logf("--- Pages Map Contents (Expected Host:Port: %s) ---", expectedHostPort)
	for k, v := range pages {
		t.Logf("Key: %s, Count: %d", k, v)
		// Check if the key starts with the expected host:port to categorize
		// This assumes external URLs are correctly excluded *before* storing.
		// If an external URL *was* normalized and stored (incorrectly),
		// it would fail the "unexpected internal key" check later if its host part differs,
		// or fail the "expected key" check if it matches no expected internal key.
		if strings.HasPrefix(k, expectedHostPort) {
			foundInternalKeys[k] = true
		} else {
			// This case should ideally not happen if external URLs are filtered correctly
			// It suggests either a bug where external URLs are stored, or a malformed internal key
			t.Errorf("Found key '%s' in map that does NOT start with expected host:port '%s'. This indicates external URL storage bug or malformed key.", k, expectedHostPort)
			foundMalformedKeys[k] = true
		}
	}
	t.Logf("--- End Pages Map ---")

	// 1. Check if all expected INTERNAL keys are present
	for expected := range expectedInternalKeys {
		if _, found := foundInternalKeys[expected]; !found {
			t.Errorf("Expected internal URL key '%s' (host:port/path format) not found in pages map", expected)
		}
	}

	// 2. Check if any unexpected INTERNAL keys were added
	// (Keys that start with expectedHostPort but aren't in expectedInternalKeys)
	for found := range foundInternalKeys {
		if _, expected := expectedInternalKeys[found]; !expected {
			t.Errorf("Found unexpected internal URL key '%s' in pages map. Check crawl logic or test's expected keys.", found)
		}
	}

	// 3. CRITICAL: Double-check if any malformed/external keys were found earlier
	// This reinforces the check that only correctly formatted internal keys should be present.
	if len(foundMalformedKeys) > 0 {
		// Error logged previously when key was found
		t.Errorf("Test failed because malformed or potentially external keys were found in the pages map (see previous errors).")
	}
}
