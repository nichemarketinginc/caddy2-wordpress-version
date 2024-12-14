package wpversion

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(&WPVersion{})
	httpcaddyfile.RegisterHandlerDirective("wp_version", parseCaddyfile)
}

// WPVersion is the middleware struct.
type WPVersion struct {
	BasePath            string        `json:"base_path"`
	CacheExpiryDuration time.Duration `json:"cache_expiry_duration"`
	cache               map[string]cacheEntry
	mu                  sync.RWMutex
}

type cacheEntry struct {
	version   string
	timestamp time.Time
}

// CaddyModule returns the Caddy module information.
func (m *WPVersion) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.wp_version",
		New: func() caddy.Module { return &WPVersion{} },
	}
}

// ServeHTTP implements the caddyhttp.MiddlewareHandler interface.
func (m *WPVersion) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Ensure the host is not empty
	host := r.Host
	if host == "" {
		return next.ServeHTTP(w, r) // Skip middleware if no host is provided
	}

	// Retrieve the replacer from the request context
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	// Try to get the WordPress version from cache
	version := m.getCachedVersion(host)
	if version == "" {
		// Attempt to detect the version
		version = m.detectWPVersion(host)
		if version != "" {
			m.cacheVersion(host, version) // Cache the detected version
		}
	}

	// If version is still blank, respond with an error
	if version == "" {
		http.Error(w, "Unable to determine WordPress version for the requested host", http.StatusServiceUnavailable)
		return nil // Stop further request processing
	}

	// Set WordPress version in both header and request context
	r.Header.Set("X-WP-Core-Version", version)

	// Set the detected version in the replacer
	repl.Set("wp-version", version)

	// Continue processing the request
	return next.ServeHTTP(w, r)
}

// detectWPVersion scans the WordPress directory to find the version.
func (m *WPVersion) detectWPVersion(host string) string {
	dirPath := filepath.Join(m.BasePath, host, "httpdocs")
	versionFile := filepath.Join(dirPath, "wp-includes", "version.php")

	data, err := os.ReadFile(versionFile)
	if err != nil {
		return ""
	}

	version := extractVersion(string(data))
	return version
}

// extractVersion parses the WordPress version from the version.php file content.
func extractVersion(fileContent string) string {
	lines := strings.Split(fileContent, "\n")
	for _, line := range lines {
		if strings.Contains(line, "$wp_version") {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				version := strings.TrimSpace(parts[1])
				version = strings.Trim(version, "'\";")
				return version
			}
		}
	}
	return ""
}

// getCachedVersion retrieves the cached WordPress version for a host.
func (m *WPVersion) getCachedVersion(host string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, exists := m.cache[host]
	if !exists || time.Since(entry.timestamp) > m.CacheExpiryDuration {
		return ""
	}
	return entry.version
}

// cacheVersion stores the WordPress version in the cache.
func (m *WPVersion) cacheVersion(host, version string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cache == nil {
		m.cache = make(map[string]cacheEntry)
	}
	m.cache[host] = cacheEntry{
		version:   version,
		timestamp: time.Now(),
	}
}

// parseCaddyfile parses the Caddyfile configuration.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m WPVersion
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return &m, err
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *WPVersion) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "base_path":
				if !d.Args(&m.BasePath) {
					return d.ArgErr()
				}
			case "wp_version_cache_expiry":
				var expiryHoursStr string
				if !d.Args(&expiryHoursStr) {
					return d.ArgErr()
				}
				expiryHours, err := strconv.Atoi(expiryHoursStr)
				if err != nil {
					return d.Errf("invalid value for wp_version_cache_expiry: %s", expiryHoursStr)
				}
				m.CacheExpiryDuration = time.Duration(expiryHours) * time.Hour
			default:
				return d.Errf("Unknown directive: %s", d.Val())
			}
		}
	}
	return nil
}

func (m *WPVersion) Provision(ctx caddy.Context) error {

	// Resolve global placeholders (e.g., environment variables) in BasePath
	repl := caddy.NewReplacer()
	m.BasePath = repl.ReplaceAll(m.BasePath, "")

	// Ensure BasePath is set and valid
	if m.BasePath == "" {
		return fmt.Errorf("base_path cannot be empty")
	}

	// Initialize cache if not already done
	if m.cache == nil {
		m.cache = make(map[string]cacheEntry)
	}

	return nil
}

func (m *WPVersion) Cleanup() error {
	return nil
}

func (m *WPVersion) InterfaceGuard() caddyhttp.MiddlewareHandler {
	return (*WPVersion)(nil)
}
