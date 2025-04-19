package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/philippgille/chromem-go"
)

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

	a.metadata.DataPath = a.cfg.DocsDir
	// Walk through docs directory
	err := filepath.Walk(a.cfg.DocsDir, func(path string, info os.FileInfo, err error) error {
		log.Printf("Walking path: %s", path)
		if err != nil {
			return err
		}

		// Skip directories and non-text files
		if info.IsDir() || !(isTextFile(path) || isPDFFile(path)) {
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

		var content string
		if isPDFFile(path) {
			// Extract text from PDF
			f, rErr := os.Open(path)
			if rErr != nil {
				return fmt.Errorf("failed to open PDF file %s: %w", path, rErr)
			}
			defer f.Close()
			pdfReader, rErr := pdf.NewReader(f, info.Size())
			if rErr != nil {
				return fmt.Errorf("failed to read PDF file %s: %w", path, rErr)
			}
			var sb strings.Builder
			for i := 1; i <= pdfReader.NumPage(); i++ {
				page := pdfReader.Page(i)
				if page.V.IsNull() {
					continue
				}
				text, _ := page.GetPlainText(nil)
				sb.WriteString(text)
			}
			content = sb.String()
		} else {
			// Read and index text file
			b, rErr := os.ReadFile(path)
			if rErr != nil {
				return fmt.Errorf("failed to read file %s: %w", path, rErr)
			}
			content = string(b)
		}

		// Split content into chunks
		chunks := splitIntoChunks(content, 2000) // ~2000 chars per chunk
		for i, chunk := range chunks {
			doc := chromem.Document{
				ID:      fmt.Sprintf("%s#chunk-%d", relPath, i),
				Content: chunk,
			}
			if err := coll.AddDocuments(ctx, []chromem.Document{doc}, runtime.NumCPU()); err != nil {
				return fmt.Errorf("failed to add document chunk %s: %w", path, err)
			}
		}

		// Update metadata (store only for the file, not per chunk)
		a.metadata.Files[relPath] = FileInfo{
			Path:         relPath,
			LastModified: info.ModTime(),
			Size:         info.Size(),
		}

		log.Printf("Indexed file: %s (%d chunks)", relPath, len(chunks))
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

// Add helper for PDF file detection
func isPDFFile(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".pdf"
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

// Helper to split text into chunks of maxChars length
func splitIntoChunks(text string, maxChars int) []string {
	var chunks []string
	for start := 0; start < len(text); start += maxChars {
		end := start + maxChars
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
	}
	return chunks
}
