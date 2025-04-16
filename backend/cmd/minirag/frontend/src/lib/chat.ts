import { ChatRequest, ChatResponse } from './types';

const API_URL = '/chat';

export async function sendChatMessage(request: ChatRequest): Promise<ChatResponse> {
  const response = await fetch(API_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    throw new Error(`Chat API error: ${response.statusText}`);
  }

  return response.json();
} 