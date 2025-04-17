package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"minirag/internal/config"

	"github.com/philippgille/chromem-go"
)

type App struct {
	cfg           *config.Config
	db            *chromem.DB
	metadata      *Metadata
	embeddingFunc chromem.EmbeddingFunc
}

type Metadata struct {
	Files map[string]FileInfo `json:"files"`
}

type FileInfo struct {
	Path         string    `json:"path"`
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
}

func NewApp(cfg *config.Config) (*App, error) {
	app := &App{
		cfg:      cfg,
		metadata: &Metadata{Files: make(map[string]FileInfo)},
	}

	// Initialize embedding function
	ollamaEmbeddingURL := cfg.OllamaURL + "/api"
	app.embeddingFunc = chromem.NewEmbeddingFuncOllama(cfg.OllamaEmbedModel, ollamaEmbeddingURL)

	// Initialize vector database
	app.db = chromem.NewDB()

	// Load metadata first
	if err := app.loadMetadata(); err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	// Load existing DB if it exists
	if _, err := os.Stat(cfg.DBFile); err == nil {
		log.Printf("Found existing DB file, loading...")
		if err := app.loadDB(); err != nil {
			return nil, fmt.Errorf("failed to load vector database: %w", err)
		}

		log.Printf("Successfully restored collection with %d documents", len(app.metadata.Files))
	} else {
		log.Printf("No existing DB file found, starting fresh")
		// Create initial collection if no DB exists
		_, err = app.db.CreateCollection("docs", map[string]string{}, app.embeddingFunc)
		if err != nil {
			return nil, fmt.Errorf("failed to create initial collection: %w", err)
		}
	}

	return app, nil
}

func (a *App) Run(mux *http.ServeMux) error {

	// Index documents
	if err := a.indexDocuments(); err != nil {
		return fmt.Errorf("failed to index documents: %w", err)
	}

	// Start HTTP server
	mux.HandleFunc("/query", a.handleQuery)
	mux.HandleFunc("/chat", a.handleChat)
	mux.HandleFunc("/debug/db", a.handleDebugDB)

	log.Printf("Server starting on port %d...", a.cfg.Port)
	log.Printf("Documents indexed: %d", len(a.metadata.Files))
	return http.ListenAndServe(":"+strconv.Itoa(a.cfg.Port), mux)
}

func (a *App) indexDocuments() error {
	ctx := context.Background()

	// Get existing collection or create new one
	coll := a.db.GetCollection("docs", a.embeddingFunc)
	if coll == nil {
		var err error
		coll, err = a.db.CreateCollection("docs", map[string]string{}, a.embeddingFunc)
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}

	// If force-reindex is set, clear everything
	if a.cfg.ForceReindex {
		log.Printf("Force reindexing enabled, clearing existing metadata and collection")
		a.metadata.Files = make(map[string]FileInfo)
		// Remove and recreate collection
		a.db.DeleteCollection("docs")
		coll, _ = a.db.CreateCollection("docs", map[string]string{}, a.embeddingFunc)
	}

	log.Printf("Current metadata contains %d files", len(a.metadata.Files))
	log.Printf("Indexing documents in: %s", a.cfg.DocsDir)

	// Walk through docs directory
	err := filepath.Walk(a.cfg.DocsDir, func(path string, info os.FileInfo, err error) error {
		log.Printf("Walking path: %s", path)
		if err != nil {
			return err
		}

		// Skip directories and non-text files
		if info.IsDir() || !isTextFile(path) {
			log.Printf("Skipping non-text file: %s", path)
			return nil
		}

		// Check if file needs indexing
		relPath, _ := filepath.Rel(a.cfg.DocsDir, path)
		fileInfo, exists := a.metadata.Files[relPath]
		if !a.cfg.ForceReindex && exists && fileInfo.LastModified.Equal(info.ModTime()) && fileInfo.Size == info.Size() {
			log.Printf("Skipping unchanged file: %s", relPath)
			return nil
		}
		log.Printf("Indexing file: %s", relPath)

		// Read and index file
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		doc := chromem.Document{
			ID:      relPath,
			Content: string(content),
		}

		// Add document with proper error handling
		if err := coll.AddDocuments(ctx, []chromem.Document{doc}, runtime.NumCPU()); err != nil {
			return fmt.Errorf("failed to add document %s: %w", path, err)
		}

		// Update metadata
		a.metadata.Files[relPath] = FileInfo{
			Path:         relPath,
			LastModified: info.ModTime(),
			Size:         info.Size(),
		}

		log.Printf("Indexed file: %s", relPath)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk docs directory: %w", err)
	}

	// Save metadata and DB
	if err := a.saveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	if err := a.saveDB(); err != nil {
		return fmt.Errorf("failed to save vector database: %w", err)
	}

	return nil
}

func (a *App) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Query: %s", req.Query)
	// Get relevant documents
	coll := a.db.GetCollection("docs", a.embeddingFunc)

	results, err := coll.Query(r.Context(), req.Query, 5, nil, nil)
	if err != nil {
		log.Printf("Query failed: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (a *App) loadMetadata() error {
	f, err := os.Open(a.cfg.MetadataFile)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(&a.metadata)
}

func (a *App) saveMetadata() error {
	f, err := os.Create(a.cfg.MetadataFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(a.metadata)
}

func (a *App) loadDB() error {
	log.Printf("Loading vector database from: %s", a.cfg.DBFile)
	err := a.db.ImportFromFile(a.cfg.DBFile, "", "docs")
	if err != nil {
		return fmt.Errorf("failed to import DB: %w", err)
	}

	// Проверяем состояние после загрузки
	coll := a.db.GetCollection("docs", a.embeddingFunc)
	if coll == nil {
		log.Printf("Warning: Collection 'docs' not found after DB load")
	} else {
		log.Printf("Successfully loaded vector database and found 'docs' collection")
	}

	return nil
}

func (a *App) saveDB() error {
	return a.db.ExportToFile(a.cfg.DBFile, true, "", "docs")
}

func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	textExtensions := map[string]bool{
		".txt": true,
		".md":  true,
		".rst": true,
		".csv": true,
		".log": true,
	}
	return textExtensions[ext]
}

func (a *App) handleDebugDB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type DebugInfo struct {
		CollectionName string              `json:"collection_name"`
		DocumentCount  int                 `json:"document_count"`
		Metadata       map[string]FileInfo `json:"metadata"`
		Config         struct {
			OllamaURL        string `json:"ollama_url"`
			OllamaModel      string `json:"ollama_model"`
			OllamaEmbedModel string `json:"ollama_embed_model"`
		} `json:"config"`
	}

	debugInfo := DebugInfo{
		CollectionName: "docs",
		DocumentCount:  len(a.metadata.Files),
		Metadata:       a.metadata.Files,
		Config: struct {
			OllamaURL        string `json:"ollama_url"`
			OllamaModel      string `json:"ollama_model"`
			OllamaEmbedModel string `json:"ollama_embed_model"`
		}{
			OllamaURL:        a.cfg.OllamaURL,
			OllamaModel:      a.cfg.OllamaModel,
			OllamaEmbedModel: a.cfg.OllamaEmbedModel,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(debugInfo)
}
