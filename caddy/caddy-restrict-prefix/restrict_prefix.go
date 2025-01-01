package restrictprefix

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	/*
		Your module needs to register itself with Caddy upon initialization.
		The caddy.RegisterModule function accepts any object that implements the
		caddy.Module interface.
	*/
	caddy.RegisterModule(RestrictPrefix{})
}

// RestrictPrefix is middleware that restricts requests where any portion
// of the URI matches a given prefix.
type RestrictPrefix struct {
	Prefix string      `json:"prefix,omitempty"` // the prefix to restrict
	logger *zap.Logger // the logger
}

// CaddyModule returns the Caddy module information.
func (RestrictPrefix) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		/*
			Caddy requires an ID for each module. Since you’re creating an
			HTTP middleware handler, you’ll use the
			ID http.handler.restrict_prefix, where restrict_prefix is the unique name
			of your module.
		*/
		ID: "http.handlers.restrict_prefix",
		/*
			Caddy also expects a function that can create a new
			instance of your module.
		*/
		New: func() caddy.Module { return new(RestrictPrefix) },
	}
}

// Provision a Zap logger to RestrictPrefix.
/*
Caddy will recognize
that your module implements the caddy.Provisioner interface and call this
method. You can then retrieve the logger from the given caddy.Context
*/
func (p *RestrictPrefix) Provision(ctx caddy.Context) error {
	p.logger = ctx.Logger(p)
	return nil
}

// Validate the prefix from the module's configuration, setting the
// default prefix "." if necessary.
/*
Likewise, Caddy will call your module’s Validate method since it implements the caddy.
Validator interface. You can use this method to make sure
all required settings have been unmarshaled from the configuration into
your module. If anything goes wrong, you can return an error and Caddy
will complain on your behalf. In this example, you’re using this method to
set the default prefix if one was not provided in the configuration.
*/
func (p *RestrictPrefix) Validate() error {
	if p.Prefix == "" {
		p.Prefix = "."
	}
	return nil
}

// ServeHTTP implements the caddyhttp.MiddlewareHandler interface.
func (p RestrictPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request,
	next caddyhttp.Handler) error {
	for _, part := range strings.Split(r.URL.Path, "/") {
		if strings.HasPrefix(part, p.Prefix) {
			http.Error(w, "Not Found", http.StatusNotFound)
			if p.logger != nil {
				// log the restricted prefix
				p.logger.Debug(fmt.Sprintf(
					"restricted prefix: %q in %s", part, r.URL.Path))
			}
			return nil
		}
	}
	return next.ServeHTTP(w, r)
}

/*
It’s a good practice to guard against interface changes by explicitly
making sure your module implements the expected interfaces . If one of
these interfaces happens to change in the future (for example, if you add a
new method), these interface guards will cause compilation to fail, giving
you an early warning that you need to adapt your code.
*/
var (
	_ caddy.Provisioner           = (*RestrictPrefix)(nil)
	_ caddy.Validator             = (*RestrictPrefix)(nil)
	_ caddyhttp.MiddlewareHandler = (*RestrictPrefix)(nil)
)
