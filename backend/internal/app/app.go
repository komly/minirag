package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"bytes"

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
	Files    map[string]FileInfo `json:"files"`
	DataPath string              `json:"data_path"`
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

	return app, nil
}

func (a *App) Run(mux *http.ServeMux, addr string) error {
	// Ensure Ollama and models are available
	if err := ensureOllamaAndModels(a.cfg); err != nil {
		return fmt.Errorf("ollama model check failed: %w", err)
	}

	// Load metadata first
	_ = a.loadMetadata() // ignore error, may not exist

	// Invalidate metadata if data dir changed
	absDataDir, err := filepath.Abs(a.cfg.DataDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute data dir: %w", err)
	}
	needInvalidate := a.metadata.DataPath != "" && a.metadata.DataPath != absDataDir
	if needInvalidate {
		log.Printf("Data directory changed from %s to %s, invalidating metadata and index...", a.metadata.DataPath, absDataDir)
		a.metadata.Files = make(map[string]FileInfo)
		a.metadata.DataPath = absDataDir
		_ = os.Remove(a.cfg.MetadataFile)
		_ = os.Remove(a.cfg.DBFile)
		a.db.DeleteCollection("docs")
		log.Printf("Deleted metadata and DB files")
		// Save the new (empty) metadata with updated DataPath
		if err := a.saveMetadata(); err != nil {
			return fmt.Errorf("failed to save new metadata: %w", err)
		}
	}

	// Load existing DB if it exists
	if _, err := os.Stat(a.cfg.DBFile); err == nil {
		log.Printf("Found existing DB file, loading...")
		if err := a.loadDB(); err != nil {
			return fmt.Errorf("failed to load vector database: %w", err)
		}

		log.Printf("Successfully restored collection with %d documents", len(a.metadata.Files))
	} else {
		log.Printf("No existing DB file found, starting fresh")
		// Create initial collection if no DB exists
		_, err = a.db.CreateCollection("docs", map[string]string{}, a.embeddingFunc)
		if err != nil {
			return fmt.Errorf("failed to create initial collection: %w", err)
		}
	}

	// Index documents
	if err := a.indexDocuments(); err != nil {
		return fmt.Errorf("failed to index documents: %w", err)
	}

	// Start HTTP server
	mux.HandleFunc("/query", a.handleQuery)
	mux.HandleFunc("/chat", a.handleChat)
	mux.HandleFunc("/debug/db", a.handleDebugDB)

	log.Printf("Documents indexed: %d", len(a.metadata.Files))
	log.Printf("Server is running on http://%s", trimHostPrefix(addr))
	return http.ListenAndServe(addr, mux)
}

// Helper to print address nicely in logs
func trimHostPrefix(addr string) string {
	if addr == "" {
		return "localhost"
	}
	if addr[0] == ':' {
		return "127.0.0.1" + addr
	}
	return addr
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

func ensureOllamaAndModels(cfg *config.Config) error {
	type ollamaPullRequest struct {
		Name   string `json:"name"`
		Stream bool   `json:"stream"`
	}

	// 1. Check if Ollama is running
	resp, err := http.Get(cfg.OllamaURL + "/api/tags")
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("ollama is not running or not reachable at %s", cfg.OllamaURL)
	}
	defer resp.Body.Close()

	// 2. Check if chat model exists
	models := []string{cfg.OllamaModel, cfg.OllamaEmbedModel}
	for _, model := range models {
		found := false
		resp, err := http.Get(cfg.OllamaURL + "/api/tags")
		if err == nil && resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			if bytes.Contains(body, []byte(model)) {
				found = true
			}
		}
		if !found {
			log.Printf("Model %s not found, pulling...", model)
			pullReq := ollamaPullRequest{Name: model, Stream: false}
			b, _ := json.Marshal(pullReq)
			pullResp, err := http.Post(cfg.OllamaURL+"/api/pull", "application/json", bytes.NewBuffer(b))
			if err != nil {
				return fmt.Errorf("failed to pull model %s: %v", model, err)
			}
			defer pullResp.Body.Close()
			if pullResp.StatusCode != 200 {
				return fmt.Errorf("failed to pull model %s: status %d", model, pullResp.StatusCode)
			}
			log.Printf("Model %s pulled successfully", model)
		} else {
			log.Printf("Model %s is available", model)
		}
	}
	return nil
}
