import { useEffect, useRef } from "react"
import { useChatStore } from "../store/chat"
import { ChatMessage } from "./chat-message"
import { ChatInput } from "./chat-input"

export function Chat() {
  const messages = useChatStore((state) => state.messages)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  return (
    <div className="flex h-[calc(100vh-7rem)] flex-col rounded-lg border border-border/40 bg-background/95 shadow-sm">
      <div className="flex-1 overflow-y-auto px-4 py-6">
        <div className="flex flex-col space-y-6">
          {messages.length === 0 ? (
            <div className="flex h-full items-center justify-center">
              <p className="text-center text-sm text-muted-foreground">
                No messages yet. Start a conversation!
              </p>
            </div>
          ) : (
            messages.map((message) => (
              <ChatMessage key={message.id} message={message} />
            ))
          )}
          <div ref={messagesEndRef} />
        </div>
      </div>
      <div className="border-t border-border/40">
        <ChatInput />
      </div>
    </div>
  )
} 