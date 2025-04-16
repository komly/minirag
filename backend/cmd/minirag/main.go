package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	"minirag/internal/app"
	"minirag/internal/config"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func main() {
	cfg := &config.Config{}

	flag.StringVar(&cfg.DocsDir, "docs", "./docs", "Directory containing documents to index")
	flag.StringVar(&cfg.DataDir, "data", "./data", "Directory for storing index and metadata")
	flag.StringVar(&cfg.OllamaURL, "ollama-url", "http://127.0.0.1:11434", "Ollama API URL")
	flag.StringVar(&cfg.OllamaModel, "ollama-model", "gemma3:4b", "Ollama model name for chat")
	flag.StringVar(&cfg.OllamaEmbedModel, "ollama-embed-model", "nomic-embed-text", "Ollama model name for embeddings")
	flag.IntVar(&cfg.Port, "port", 8080, "Server port")
	flag.BoolVar(&cfg.DevMode, "dev", false, "Run in development mode")
	flag.BoolVar(&cfg.ForceReindex, "force-reindex", false, "Force reindexing of all documents, ignoring saved state")
	flag.Parse()

	// Create a new mux
	mux := http.NewServeMux()

	if cfg.DevMode {
		// Development mode - proxy to Vite dev server
		viteURL, err := url.Parse("http://localhost:5173")
		if err != nil {
			log.Fatal(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(viteURL)

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})

		log.Printf("Development server starting on :%d...", cfg.Port)
		log.Printf("Proxying requests to Vite dev server at http://localhost:5173")
	} else {
		// Production mode - serve embedded frontend files
		frontend, err := fs.Sub(frontendFS, "frontend/dist")
		if err != nil {
			log.Fatal(err)
		}

		frontendHandler := http.FileServer(http.FS(frontend))
		mux.Handle("/", frontendHandler)

		log.Printf("Production server starting on :%d...", cfg.Port)
	}

	// Ensure directories exist
	for _, dir := range []string{cfg.DocsDir, cfg.DataDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Fatalf("Directory does not exist: %s", dir)
		}
	}

	// Initialize metadata file path
	cfg.MetadataFile = filepath.Join(cfg.DataDir, "metadata.json")
	cfg.DBFile = filepath.Join(cfg.DataDir, "vectordb.gob")

	// Create and run application
	app, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := app.Run(mux); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
