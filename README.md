# MiniRAG - Minimalist, Self-Hosted, Single-Binary RAG

> **Killer Feature:**  
> **Start chatting with your data in seconds:**  
> ```sh
> minirag -docs=mydata
> ```
> That's it. No setup, no config, just point to your folder and go!

![Demo Screenshot](docs/screenshot.png)

**MiniRAG** is a minimalist, self-hosted Retrieval-Augmented Generation (RAG) solution.  
It runs as a **single static binary** on your machine—no cloud, no dependencies, no external services.

- **Self-hosted:** All data and LLM inference stay on your device.
- **Single binary:** Download, run, and you're ready—no Docker, no Python, no Node.js required.
- **Private & secure:** Your documents and chat never leave your computer.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?style=for-the-badge&logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-5+-3178C6?style=for-the-badge&logo=typescript)
![Ollama](https://img.shields.io/badge/Ollama-0.1+-000000?style=for-the-badge&logo=ollama)

MiniRAG is a lightweight, production-ready implementation of Retrieval-Augmented Generation (RAG) that combines the power of large language models with efficient document retrieval. Built with Go and React, it provides a modern, responsive interface for interacting with your documents.

## 🍺 Install MiniRAG via Homebrew

```sh
brew tap komly/minirag
brew install minirag
```

Then run:

```sh
minirag -docs=mydata
```

- By default, MiniRAG will store its data in `~/.minirag`.
- You can specify a different data directory with the `-data` flag if needed.
- The server will listen on `:7492` by default. You can change this with the `-http` flag.

---

## 🚀 Features

- **Document Indexing**: Automatically indexes text documents with vector embeddings
- **Smart Search**: Uses semantic search to find relevant document chunks
- **Chat Interface**: Modern React-based chat UI with source attribution
- **Ollama Integration**: Works with any Ollama-compatible model
- **Production Ready**: Single binary deployment with embedded frontend

## 📋 Prerequisites

- Go 1.21 or later
- Node.js 18+ and pnpm
- Ollama running locally
- Text documents to index

## 🛡️ 100% Local LLM & Data Safety

**minirag** runs **entirely locally** on your machine:
- All LLM inference and document indexing happen on your computer.
- **No data ever leaves your device.**
- Your documents and chat history are never sent to any external server or cloud.

> **Your data stays private and secure.**

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
- `--data`: Directory for storing index and metadata (default: "~/.minirag")
- `--ollama-url`: Ollama API URL (default: "http://127.0.0.1:11434")
- `--ollama-model`: Chat model name (default: "gemma3:12b")
- `--ollama-embed-model`: Embedding model name (default: "nomic-embed-text:latest")
- `--http`: HTTP listen address (default: ":7492", e.g. ":7492" or "0.0.0.0:7492")
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

## ⚙️ Command-line Flags

The `minirag` binary supports the following flags:

| Flag           | Default         | Description                                                                 |
|----------------|----------------|-----------------------------------------------------------------------------|
| `-docs`        | `docs/`        | Path to the directory with documents to index and search.                   |
| `-data`        | `~/.minirag`   | Path to the directory for storing vector DB and metadata.                   |
| `-http`        | `:7492`        | HTTP listen address (e.g. ':7492' or '0.0.0.0:7492').                       |
| `-dev`         | (off)          | Enable development mode (useful for hot-reload and debugging).              |
| `-force-reindex` | (off)        | Force reindexing of all documents on startup, ignoring previous metadata.   |

### Example usage

```sh
minirag -docs=app/ -data=~/.minirag -http=:8080
```

- `-docs=app/` — use the `app/` directory for your documents.
- `-data=~/.minirag` — store all vector DB and metadata in the `~/.minirag` directory.
- `-http=:8080` — run the web server on port 8080.

> **Tip:** You can combine flags as needed. If you omit a flag, the default value will be used.

--- 