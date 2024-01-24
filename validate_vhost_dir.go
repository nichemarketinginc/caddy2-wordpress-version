package validatehostdir

import (
	"net/http"
	"os"
	"path/filepath"
  "log"
	//  "strings"

	"github.com/caddyserver/caddy/v2"
  "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

func init() {
	caddy.RegisterModule(ValidateVhostDir{})
	httpcaddyfile.RegisterHandlerDirective("validate_vhost_dir", parseDirective)
}

type ValidateVhostDir struct {
  Name string
	BasePath string `json:"base_path"`
}

func parseDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {

	var m ValidateVhostDir
	if err := m.UnmarshalCaddyfile(h.Dispenser); err != nil {
			return nil, err
	}
	return m, nil
}

func (m *ValidateVhostDir) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
  
  m.Name = d.Val();
  log.Printf("ValidateVhostDir after directive name parsed: %+v\n", m) 

	if !d.Args(&m.BasePath) {
    log.Println("No args were found at all after directive.")
		// not enough args
		return d.ArgErr()
	}

  log.Printf("ValidateVhostDir after first directive argument parsed: %+v\n", m) 

	if d.NextArg() {
		// too many args
		return d.ArgErr()
	}

	return nil
}


// Interface guard
var _ caddyfile.Unmarshaler = (*ValidateVhostDir)(nil)


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
