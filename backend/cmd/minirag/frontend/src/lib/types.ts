export interface ChatRequest {
  query: string;
  temperature?: number;
  max_tokens?: number;
}

export interface Source {
  id: string;
  content: string;
  similarity: number;
}

export interface ChatResponse {
  answer: string;
  sources: Source[];
  model: string;
  processing_time_ms: number;
} 