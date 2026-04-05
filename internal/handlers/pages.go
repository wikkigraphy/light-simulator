package handlers

import (
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"
)

// Pages serves HTML template pages.
type Pages struct {
	templateDir string
	logger      *slog.Logger
	cache       map[string]*template.Template
	mu          sync.RWMutex
	isDev       bool
}

// NewPages creates a page handler with template directory and dev-mode flag.
func NewPages(templateDir string, isDev bool, logger *slog.Logger) *Pages {
	return &Pages{
		templateDir: templateDir,
		logger:      logger,
		cache:       make(map[string]*template.Template),
		isDev:       isDev,
	}
}

// RegisterRoutes mounts page routes on the given mux.
func (p *Pages) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", p.Index)
	mux.HandleFunc("GET /simulator", p.Simulator)
	mux.HandleFunc("GET /cheatsheet", p.Cheatsheet)
}

// Index serves the landing page.
func (p *Pages) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	p.render(w, "index.html", nil)
}

// Simulator serves the interactive lighting simulator page.
func (p *Pages) Simulator(w http.ResponseWriter, _ *http.Request) {
	p.render(w, "simulator.html", nil)
}

// Cheatsheet serves the lighting cheatsheet page.
func (p *Pages) Cheatsheet(w http.ResponseWriter, _ *http.Request) {
	p.render(w, "cheatsheet.html", nil)
}

func (p *Pages) render(w http.ResponseWriter, name string, data any) {
	tmpl, err := p.loadTemplate(name)
	if err != nil {
		p.logger.Error("template load failed", "template", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		p.logger.Error("template execute failed", "template", name, "error", err)
	}
}

func (p *Pages) loadTemplate(name string) (*template.Template, error) {
	if !p.isDev {
		p.mu.RLock()
		if tmpl, ok := p.cache[name]; ok {
			p.mu.RUnlock()
			return tmpl, nil
		}
		p.mu.RUnlock()
	}

	layoutPath := filepath.Join(p.templateDir, "layout.html")
	pagePath := filepath.Join(p.templateDir, name)

	tmpl, err := template.ParseFiles(layoutPath, pagePath)
	if err != nil {
		return nil, err
	}

	if !p.isDev {
		p.mu.Lock()
		p.cache[name] = tmpl
		p.mu.Unlock()
	}

	return tmpl, nil
}
