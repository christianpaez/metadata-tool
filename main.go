package main

import (
	"fmt"
	"github.com/barasher/go-exiftool"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8080"
	}
	return port
}

func handleEdit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing 'file' input", http.StatusBadRequest)
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	tempFile, err := os.CreateTemp("", "upload-*"+ext)
	if err != nil {
		http.Error(w, "failed to create temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "failed to save upload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tempFile.Close()
	meta := &Metadata{
		Artist: r.FormValue("artist"),
		Title:  r.FormValue("title"),
	}
	et, err := exiftool.NewExiftool()
	if err != nil {
		http.Error(w, "failed to init exiftool: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer et.Close()
	if err := WriteImageMetadata(et, tempPath, meta); err != nil {
		http.Error(w, "failed to update image: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", header.Filename))
	http.ServeFile(w, r, tempPath)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Could not load template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", homeHandler)

	wrappedRouter := LoggingMiddleware(mux)
	fmt.Printf("Server started on port: %v\n", getPort())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", getPort()), wrappedRouter))
}
