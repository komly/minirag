import { Bot, User } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
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
  const chatContainerRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const abortControllerRef = useRef<AbortController | null>(null);

  // Scroll to bottom only if autoScroll is enabled
  useEffect(() => {
    if (autoScroll) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, autoScroll]);

  // Disable autoscroll as soon as user scrolls up (not just when not at bottom)
  useEffect(() => {
    const chatDiv = chatContainerRef.current;
    if (!chatDiv) return;

    let lastScrollTop = chatDiv.scrollTop;

    const handleScroll = () => {
      const threshold = 40; // px
      const atBottom =
        chatDiv.scrollHeight - chatDiv.scrollTop - chatDiv.clientHeight < threshold;

      // If user scrolls up (not just not at bottom, but any upward movement), disable autoscroll
      if (chatDiv.scrollTop < lastScrollTop) {
        setAutoScroll(false);
      } else if (atBottom) {
        setAutoScroll(true);
      }
      lastScrollTop = chatDiv.scrollTop;
    };

    chatDiv.addEventListener('scroll', handleScroll);
    return () => {
      chatDiv.removeEventListener('scroll', handleScroll);
    };
  }, []);

  const sendMessage = async () => {
    if (!input.trim() || loading) return;

    const userMessage = input.trim();
    setInput('');
    setError(null);
    setLoading(true);

    // Add user message immediately
    setMessages(prev => [...prev, { role: 'user', content: userMessage }]);
    // Add empty assistant message for streaming
    setMessages(prev => [...prev, { role: 'assistant', content: '' }]);

    // Create and store AbortController
    const abortController = new AbortController();
    abortControllerRef.current = abortController;

    try {
      const response = await fetch('/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: userMessage,
          temperature: 0.7,
          max_tokens: 1000,
        }),
        signal: abortController.signal,
      });

      if (!response.body) throw new Error('No response body');

      const reader = response.body.getReader();
      const decoder = new TextDecoder('utf-8');
      let done = false;
      let assistantContent = '';
      let buffer = '';

      while (!done) {
        const { value, done: doneReading } = await reader.read();
        done = doneReading;
        if (value) {
          buffer += decoder.decode(value, { stream: true });
          const parts = buffer.split('\n\n');
          buffer = parts.pop() || '';
          for (const part of parts) {
            if (part.startsWith('data: ')) {
              const data = part.slice(6).trim();
              if (data === '[DONE]') {
                setLoading(false);
                abortControllerRef.current = null;
                break;
              }
              try {
                const parsed = JSON.parse(data);
                if (parsed.role === 'assistant') {
                  assistantContent += parsed.content;
                  setMessages(prev => {
                    const updated = [...prev];
                    updated[updated.length - 1] = {
                      ...updated[updated.length - 1],
                      content: assistantContent,
                    };
                    return updated;
                  });
                } else if (parsed.type === 'meta' && parsed.meta) {
                  setMessages(prev => {
                    const updated = [...prev];
                    updated[updated.length - 1] = {
                      ...updated[updated.length - 1],
                      ...parsed.meta,
                      content: updated[updated.length - 1].content,
                    };
                    return updated;
                  });
                }
              } catch (err) {
                // ignore JSON parse errors for non-data lines
              }
            }
          }
        }
      }
      setLoading(false);
      abortControllerRef.current = null;
    } catch (err) {
      if ((err as any).name !== 'AbortError') {
        setError(err instanceof Error ? err.message : 'An error occurred');
      }
      setLoading(false);
      abortControllerRef.current = null;
    }
  };

  // SSE streaming chat handler
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    await sendMessage();
  };

  // Handler for stop button
  const handleStop = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
      setLoading(false);
      // Remove the last assistant message if it's still empty
      setMessages(prev => {
        if (
          prev.length > 0 &&
          prev[prev.length - 1].role === 'assistant' &&
          prev[prev.length - 1].content === ''
        ) {
          return prev.slice(0, -1);
        }
        return prev;
      });
    }
  };

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)] max-w-4xl mx-auto">
      <div
        className="flex-1 overflow-y-auto p-4 space-y-4 scrollbar-hide"
        ref={chatContainerRef}
      >
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
                      <SourceCollapse
                        key={idx}
                        source={source}
                        similarity={source.similarity}
                        id={`source-${index}-${idx}`}
                      />
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
            type="button"
            onClick={() => {
              if (loading) {
                handleStop();
              } else {
                sendMessage();
              }
            }}
            disabled={!input.trim() && !loading}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
          >
            {loading ? 'Stop' : 'Send'}
          </button>

        </div>
      </form>
    </div>
  );
}

// Collapsible source component
function SourceCollapse({
  source,
  similarity,
  id,
}: {
  source: { id: string; content: string };
  similarity: number;
  id: string;
}) {
  const [open, setOpen] = useState(false);

  return (
    <div className="text-xs bg-muted rounded-lg mb-2">
      <button
        className="flex justify-between items-center w-full px-2 py-1 font-medium truncate text-left hover:bg-muted/70 rounded-t-lg"
        onClick={() => setOpen((v) => !v)}
        type="button"
      >
        <span className="truncate">{source.id}</span>
        <span className="ml-2 text-muted-foreground">{(similarity * 100).toFixed(1)}%</span>
        <span className="ml-2">{open ? '▲' : '▼'}</span>
      </button>
      {open && (
        <div className="p-2 text-muted-foreground">
          <MemoizedMarkdown content={source.content} id={id} />
        </div>
      )}
    </div>
  );
} 