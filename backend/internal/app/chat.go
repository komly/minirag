package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ChatRequest struct {
	Query string `json:"query"`
}

type ChatResponse struct {
	Answer  string     `json:"answer"`
	Sources []Document `json:"sources"`
}

type Document struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}

type ollamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message Message `json:"message"`
}

func (a *App) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get relevant documents
	coll := a.db.GetCollection("docs", a.embeddingFunc)

	results, err := coll.Query(r.Context(), req.Query, 3, nil, nil)
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}

	// Prepare context from relevant documents
	var context string
	var sources []Document
	for _, doc := range results {
		context += fmt.Sprintf("\nDocument %s:\n%s\n", doc.ID, doc.Content)
		sources = append(sources, Document{
			ID:         doc.ID,
			Content:    doc.Content,
			Similarity: float64(doc.Similarity),
		})
	}

	// Prepare prompt
	prompt := fmt.Sprintf(`Use the following documents to answer the question. If you cannot find the answer in the documents, say so.

Context:
%s

Question: %s

Answer based on the above context:`, context, req.Query)

	// Call Ollama
	ollamaReq := ollamaRequest{
		Model: a.cfg.OllamaModel,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	ollamaBody, err := json.Marshal(ollamaReq)
	if err != nil {
		http.Error(w, "Failed to create Ollama request", http.StatusInternalServerError)
		return
	}

	ollamaResp, err := http.Post(a.cfg.OllamaURL+"/api/chat", "application/json", bytes.NewBuffer(ollamaBody))
	if err != nil {
		http.Error(w, "Failed to call Ollama API", http.StatusInternalServerError)
		return
	}
	defer ollamaResp.Body.Close()

	var ollama ollamaResponse
	if err := json.NewDecoder(ollamaResp.Body).Decode(&ollama); err != nil {
		http.Error(w, "Failed to decode Ollama response", http.StatusInternalServerError)
		return
	}

	// Send response
	response := ChatResponse{
		Answer:  ollama.Message.Content,
		Sources: sources,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
