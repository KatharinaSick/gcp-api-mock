package handler

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/ksick/gcp-api-mock/internal/config"
)

// UI handles web UI endpoints with HTMX templates.
type UI struct {
	cfg       *config.Config
	templates *template.Template
}

// NewUI creates a new UI handler.
func NewUI(cfg *config.Config) *UI {
	// Parse all templates from the templates directory
	tmpl := template.Must(template.ParseGlob(filepath.Join("web", "templates", "*.html")))

	return &UI{
		cfg:       cfg,
		templates: tmpl,
	}
}

// PageData holds common data passed to templates.
type PageData struct {
	Title       string
	Environment string
}

// Index renders the main dashboard page.
func (u *UI) Index(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "GCP API Mock",
		Environment: u.cfg.Environment,
	}

	if err := u.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}
