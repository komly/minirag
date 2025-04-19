package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"
)

type ChatRequest struct {
	Query       string  `json:"query"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

type ChatResponse struct {
	Answer           string     `json:"answer"`
	Sources          []Document `json:"sources"`
	Model            string     `json:"model"`
	ProcessingTimeMs int64      `json:"processing_time_ms"`
}

type Document struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}

type ollamaRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"num_predict,omitempty"`
	Stream      bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (a *App) handleChat(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set defaults if not provided
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 10000
	}

	// Get relevant documents
	coll := a.db.GetCollection("docs", a.embeddingFunc)

	var context string
	var sources []Document

	maxResults := math.Min(10, float64(coll.Count()))
	results, err := coll.Query(r.Context(), req.Query, int(maxResults), nil, nil)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for _, doc := range results {
			context += fmt.Sprintf("\nDocument %s:\n%s\n", doc.ID, doc.Content)
			sources = append(sources, Document{
				ID:         doc.ID,
				Content:    doc.Content,
				Similarity: float64(doc.Similarity),
			})
		}
	}

	// Prepare prompt
	prompt := fmt.Sprintf(`You are a helpful AI assistant. Your task is to provide detailed and informative answers based on the given context.

Instructions:
1. Read the context carefully
2. Provide a complete, well-structured answer
3. If the context doesn't contain enough information, acknowledge that and provide general guidance
4. Always write full sentences and complete thoughts
5. Use markdown formatting for better readability

Context:
%s

User Question: %s

Important: Provide a complete, detailed response. Never stop at single words or incomplete sentences.
IMPORTANT: ANSWER IN LANGUAGE OF THE USER QUESTION.
Response:`, context, req.Query)
	// Call Ollama
	ollamaReq := ollamaRequest{
		Model:       a.cfg.OllamaModel,
		Messages:    []Message{{Role: "user", Content: prompt}},
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	}

	ollamaBody, err := json.Marshal(ollamaReq)
	if err != nil {
		log.Printf("Failed to create Ollama request: %v", err)
		http.Error(w, "Failed to create Ollama request", http.StatusInternalServerError)
		return
	}

	ollamaResp, err := http.Post(a.cfg.OllamaURL+"/api/chat", "application/json", bytes.NewBuffer(ollamaBody))
	if err != nil {
		log.Printf("Failed to call Ollama API: %v", err)
		http.Error(w, "Failed to call Ollama API", http.StatusInternalServerError)
		return
	}
	defer ollamaResp.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	type ollamaResponseChunk struct {
		CreatedAt time.Time `json:"created_at"`
		Done      bool      `json:"done"`
		Message   Message   `json:"message"`
		Model     string    `json:"model"`
	}
	// Begin streaming tokens
	decoder := json.NewDecoder(ollamaResp.Body)
	for {
		var chunk ollamaResponseChunk
		if err := decoder.Decode(&chunk); err != nil {
			break // done streaming
		}
		// send chunk (you can wrap in SSE-style JSON delta if needed)
		fmt.Fprintf(w, "data: %s\n\n", encodeJSON(map[string]string{
			"role":    chunk.Message.Role,
			"content": chunk.Message.Content,
		}))
		flusher.Flush()
	}

	// Calculate processing time
	processingTimeMs := time.Since(startTime).Milliseconds()

	// Send sources and metadata as a final event before [DONE]
	meta := map[string]interface{}{
		"sources":            sources,
		"model":              a.cfg.OllamaModel,
		"processing_time_ms": processingTimeMs,
	}
	fmt.Fprintf(w, "data: %s\n\n", encodeJSON(map[string]interface{}{
		"type": "meta",
		"meta": meta,
	}))
	flusher.Flush()

	// Send [DONE]
	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func encodeJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
