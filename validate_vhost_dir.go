package validatehostdir

import (
	"net/http"
	"os"
	"path/filepath"
	//  "strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

func init() {
	caddy.RegisterModule(ValidateVhostDir{})
	httpcaddyfile.RegisterHandlerDirective("validate_vhost_dir", parseCaddyfile)
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {

	var m ValidateVhostDir

	// Move to the next argument in the Caddyfile, which should be the base path
	if !h.NextArg() {
			return nil, h.ArgErr() // No argument found for the directive
	}
	m.BasePath = h.Val() // Assign the argument to BasePath

	// Check for any extraneous arguments
	if h.NextArg() {
			return nil, h.ArgErr() // Too many arguments provided
	}

	return m, nil
}

type ValidateVhostDir struct {
	BasePath string `json:"base_path"`
}

func (ValidateVhostDir) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.validate_vhost_dir",
		New: func() caddy.Module { return new(ValidateVhostDir) },
	}
}

func (m ValidateVhostDir) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	dirPath := filepath.Join(m.BasePath, domain)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// If the directory does not exist, respond with StatusNotFound
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	// If the directory exists, respond with StatusOK
	w.WriteHeader(http.StatusOK)
	return nil
}

var (
	_ caddyhttp.MiddlewareHandler = (*ValidateVhostDir)(nil)
)
