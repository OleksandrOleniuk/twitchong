package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/OleksandrOleniuk/twitchong/internal/config"
)

// StaticHandler handles serving static files
type StaticHandler struct {
	cfg *config.Config
}

// NewStaticHandler creates a new StaticHandler
func NewStaticHandler(cfg *config.Config) http.HandlerFunc {
	handler := &StaticHandler{cfg: cfg}
	return handler.ServeStatic
}

// ServeStatic serves static files from the web/static directory
func (h *StaticHandler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	// Get the requested file path
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Construct the full file path
	fullPath := filepath.Join("web/static", path)

	// If it's the index page, we need to inject the client ID
	if path == "/index.html" {
		h.serveIndexWithClientID(w, fullPath)
		return
	}

	// For other static files, serve them normally
	serveStaticFile(w, fullPath)
}

func (h *StaticHandler) serveIndexWithClientID(w http.ResponseWriter, filePath string) {
	// Read the template file
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading template file: %v", err)
		http.Error(w, "Error reading template file", http.StatusInternalServerError)
		return
	}

	// Parse the template
	tmpl, err := template.New("index").Parse(string(content))
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	// Create a buffer to hold the rendered template
	var buf bytes.Buffer

	// Get config values
	cfg := h.cfg

	// Execute the template with the client ID
	data := struct {
		ClientId          string
		TwitchSecretState string
	}{
		ClientId:          cfg.ClientId,
		TwitchSecretState: cfg.TwitchSecretState,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}

	// Set the content type and write the rendered template
	w.Header().Set("Content-Type", "text/html")
	w.Write(buf.Bytes())
}

func serveStaticFile(w http.ResponseWriter, filePath string) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set the appropriate content type based on file extension
	switch filepath.Ext(filePath) {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	// Serve the file
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error serving file", http.StatusInternalServerError)
		return
	}
}

