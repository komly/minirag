import { Bot, User } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { sendChatMessage } from '../lib/chat';
import { ChatResponse } from '../lib/types';
import { MemoizedMarkdown } from './MemoizedMarkdown';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  sources?: ChatResponse['sources'];
  model?: string;
  processing_time_ms?: number;
}

export function Chat() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    const userMessage = input.trim();
    setInput('');
    setError(null);
    setLoading(true);

    // Add user message immediately
    setMessages(prev => [...prev, { role: 'user', content: userMessage }]);

    try {
      const result = await sendChatMessage({
        query: userMessage,
        temperature: 0.7,
        max_tokens: 1000,
      });

      // Add assistant response
      setMessages(prev => [...prev, {
        role: 'assistant',
        content: result.answer,
        sources: result.sources,
        model: result.model,
        processing_time_ms: result.processing_time_ms,
      }]);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)] max-w-4xl mx-auto">
      <div className="flex-1 overflow-y-auto p-4 space-y-4 scrollbar-hide">
        {messages.length === 0 ? (
          <div className="flex h-full items-center justify-center">
            <p className="text-gray-500">Start a conversation by typing a message below</p>
          </div>
        ) : (
          messages.map((message, index) => (
            <div
              key={index}
              className={`flex gap-3 ${
                message.role === 'user' ? 'flex-row-reverse' : ''
              }`}
            >
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-border/50 bg-background shadow-sm">
                {message.role === 'user' ? (
                  <User className="h-5 w-5 text-foreground/60" />
                ) : (
                  <Bot className="h-5 w-5 text-foreground/60" />
                )}
              </div>
              <div className={`flex flex-col gap-1 max-w-[80%] ${
                message.role === 'user' ? 'items-end' : 'items-start'
              }`}>
               
                  <MemoizedMarkdown 
                    content={message.content} 
                    id={`msg-${index}`} 
                  />
                {message.role === 'assistant' && message.sources && message.sources.length > 0 && (
                  <div className="mt-2 space-y-2 w-full">
                    <h4 className="text-xs font-medium text-muted-foreground">Sources:</h4>
                    {message.sources.map((source, idx) => (
                      <div key={idx} className="text-xs p-2 bg-muted rounded-lg">
                        <div className="flex justify-between items-start mb-1">
                          <span className="font-medium truncate">{source.id}</span>
                          <span className="text-muted-foreground">
                            {(source.similarity * 100).toFixed(1)}%
                          </span>
                        </div>
                        <div className="text-muted-foreground">
                          <MemoizedMarkdown 
                            content={source.content} 
                            id={`source-${index}-${idx}`} 
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                )}
                {message.role === 'assistant' && message.model && (
                  <div className="text-xs text-muted-foreground mt-1">
                    Model: {message.model} • {message.processing_time_ms}ms
                  </div>
                )}
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {error && (
        <div className="p-4 text-red-700 bg-red-100 dark:bg-red-950 rounded-lg mx-4">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="p-4 border-t dark:border-gray-800">
        <div className="flex gap-2 w-full">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Type your message..."
            className="flex-1 p-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-900 dark:text-white dark:border-gray-700"
            disabled={loading}
          />
          <button
            type="submit"
            disabled={loading || !input.trim()}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
          >
            {loading ? 'Sending...' : 'Send'}
          </button>
        </div>
      </form>
    </div>
  );
} 