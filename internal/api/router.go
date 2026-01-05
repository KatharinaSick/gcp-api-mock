package api

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// pathParamsKey is the context key for storing path parameters
const pathParamsKey contextKey = "pathParams"

// PathParams holds extracted path parameters from GCP API paths
type PathParams struct {
	Bucket string
	Object string
}

// Router handles HTTP routing for the GCP mock API
type Router struct {
	mux       *http.ServeMux
	gcpRoutes []gcpRoute
}

// gcpRoute represents a GCP-style route with parameter extraction
type gcpRoute struct {
	pattern   string
	regex     *regexp.Regexp
	handler   http.HandlerFunc
	hasBucket bool
	hasObject bool
}

// NewRouter creates a new Router instance
func NewRouter() *Router {
	return &Router{
		mux:       http.NewServeMux(),
		gcpRoutes: make([]gcpRoute, 0),
	}
}

// Handle registers a handler for a pattern
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

// HandleFunc registers a handler function for a pattern
func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.mux.HandleFunc(pattern, handler)
}

// HandleGCPRoute registers a handler for a GCP-style route pattern
// Supports patterns like:
//   - /storage/v1/b/{bucket}
//   - /storage/v1/b/{bucket}/o/{object...}
func (r *Router) HandleGCPRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	route := gcpRoute{
		pattern: pattern,
		handler: handler,
	}

	// Convert pattern to regex
	regexPattern := pattern

	// Check for bucket parameter
	if strings.Contains(pattern, "{bucket}") {
		route.hasBucket = true
		regexPattern = strings.Replace(regexPattern, "{bucket}", `([^/]+)`, 1)
	}

	// Check for object parameter (with ... for wildcard matching)
	if strings.Contains(pattern, "{object...}") {
		route.hasObject = true
		regexPattern = strings.Replace(regexPattern, "{object...}", `(.+)`, 1)
	} else if strings.Contains(pattern, "{object}") {
		route.hasObject = true
		regexPattern = strings.Replace(regexPattern, "{object}", `([^/]+)`, 1)
	}

	// Anchor the pattern
	regexPattern = "^" + regexPattern + "$"
	route.regex = regexp.MustCompile(regexPattern)

	r.gcpRoutes = append(r.gcpRoutes, route)
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// Try GCP routes first
	for _, route := range r.gcpRoutes {
		if matches := route.regex.FindStringSubmatch(path); matches != nil {
			params := &PathParams{}
			matchIdx := 1

			if route.hasBucket && matchIdx < len(matches) {
				params.Bucket = matches[matchIdx]
				matchIdx++
			}

			if route.hasObject && matchIdx < len(matches) {
				// URL decode the object path
				decoded, err := url.PathUnescape(matches[matchIdx])
				if err == nil {
					params.Object = decoded
				} else {
					params.Object = matches[matchIdx]
				}
			}

			// Add params to context
			ctx := context.WithValue(req.Context(), pathParamsKey, params)
			route.handler(w, req.WithContext(ctx))
			return
		}
	}

	// Fall back to standard mux
	r.mux.ServeHTTP(w, req)
}

// GetPathParams retrieves path parameters from the request context
func GetPathParams(ctx context.Context) *PathParams {
	if params, ok := ctx.Value(pathParamsKey).(*PathParams); ok {
		return params
	}
	return nil
}

// GetPathParamsFromRequest retrieves path parameters from either the context (for GCP routes)
// or from Go 1.22+ path values (for standard mux routes with {param} patterns)
func GetPathParamsFromRequest(r *http.Request) *PathParams {
	// First try context (GCP routes)
	if params := GetPathParams(r.Context()); params != nil {
		return params
	}

	// Try Go 1.22+ path values
	bucket := r.PathValue("bucket")
	object := r.PathValue("object")

	if bucket != "" || object != "" {
		return &PathParams{
			Bucket: bucket,
			Object: object,
		}
	}

	return nil
}

// SetPathParams sets path parameters in the context (useful for testing)
func SetPathParams(ctx context.Context, params *PathParams) context.Context {
	return context.WithValue(ctx, pathParamsKey, params)
}

// ExtractPathParams extracts bucket and object names from a GCP Storage API path
// Returns the extracted parameters and a boolean indicating success
func ExtractPathParams(path string) (*PathParams, bool) {
	if path == "" {
		return nil, false
	}

	// Pattern for bucket: /storage/v1/b/{bucket}
	bucketRegex := regexp.MustCompile(`^/storage/v1/b/([^/]+)/?$`)
	// Pattern for object: /storage/v1/b/{bucket}/o/{object}
	objectRegex := regexp.MustCompile(`^/storage/v1/b/([^/]+)/o/(.+)$`)

	// Try object pattern first (more specific)
	if matches := objectRegex.FindStringSubmatch(path); matches != nil {
		// URL decode the object path
		object, err := url.PathUnescape(matches[2])
		if err != nil {
			object = matches[2]
		}
		return &PathParams{
			Bucket: matches[1],
			Object: object,
		}, true
	}

	// Try bucket pattern
	if matches := bucketRegex.FindStringSubmatch(path); matches != nil {
		if matches[1] == "" {
			return nil, false
		}
		return &PathParams{
			Bucket: matches[1],
			Object: "",
		}, true
	}

	return nil, false
}
