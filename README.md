# MiniRAG - A Minimalist RAG Implementation

> A single-binary RAG system that lets you chat with your documents. Just drop your files, run the binary, and start asking questions.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?style=for-the-badge&logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-5+-3178C6?style=for-the-badge&logo=typescript)
![Ollama](https://img.shields.io/badge/Ollama-0.1+-000000?style=for-the-badge&logo=ollama)

MiniRAG is a lightweight, production-ready implementation of Retrieval-Augmented Generation (RAG) that combines the power of large language models with efficient document retrieval. Built with Go and React, it provides a modern, responsive interface for interacting with your documents.

## 🚀 Features

- **Document Indexing**: Automatically indexes text documents with vector embeddings
- **Smart Search**: Uses semantic search to find relevant document chunks
- **Chat Interface**: Modern React-based chat UI with source attribution
- **Ollama Integration**: Works with any Ollama-compatible model
- **Development Mode**: Hot reloading for both frontend and backend
- **Production Ready**: Single binary deployment with embedded frontend

## 📋 Prerequisites

- Go 1.21 or later
- Node.js 18+ and pnpm
- Ollama running locally
- Text documents to index

## 🛠️ Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/minirag.git
   cd minirag
   ```

2. Install frontend dependencies:
   ```bash
   cd backend/cmd/minirag/frontend
   pnpm install
   ```

3. Create required directories:
   ```bash
   mkdir -p data docs
   ```

4. Add your documents to the `docs` directory

## 🚀 Usage

### Development Mode

1. Start the frontend development server:
   ```bash
   cd backend/cmd/minirag/frontend
   pnpm dev
   ```

2. In another terminal, start the backend:
   ```bash
   cd backend
   go run cmd/minirag/main.go --dev
   ```

### Production Mode

1. Build the frontend:
   ```bash
   cd backend/cmd/minirag/frontend
   pnpm build
   ```

2. Run the server:
   ```bash
   cd backend
   go run cmd/minirag/main.go
   ```

### Command Line Options

- `--docs`: Directory containing documents to index (default: "./docs")
- `--data`: Directory for storing index and metadata (default: "./data")
- `--ollama-url`: Ollama API URL (default: "http://127.0.0.1:11434")
- `--ollama-model`: Chat model name (default: "gemma3:4b")
- `--ollama-embed-model`: Embedding model name (default: "nomic-embed-text")
- `--port`: Server port (default: 8080)
- `--dev`: Run in development mode
- `--force-reindex`: Force reindexing of all documents

## 🔍 API Endpoints

- `POST /chat`: Chat with the system
  ```json
  {
    "query": "your question",
    "temperature": 0.7,
    "max_tokens": 1000
  }
  ```

- `POST /query`: Search documents
  ```json
  {
    "query": "search term"
  }
  ```

- `GET /debug/db`: View database state

## 🏗️ Architecture

### Frontend
- React Vite TypeScript SPA
- Modern UI with responsive design
- Real-time chat interface
- Source attribution display

### Backend
- Go application serving the frontend
- Vector database using chromem-go
- Integration with Ollama API
- Document processing and indexing

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [chromem-go](https://github.com/philippgille/chromem-go) for vector database
- [Ollama](https://ollama.ai/) for LLM integration
- [React](https://reactjs.org/) and [Vite](https://vitejs.dev/) for frontend 