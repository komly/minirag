import { Message } from "../store/chat"
import { cn } from "../lib/utils"
import { UserCircle, Bot } from "lucide-react"

interface ChatMessageProps {
  message: Message
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === "user"
  return (
    <div
      className={cn(
        "group relative flex w-full items-start gap-3",
        isUser ? "flex-row-reverse" : ""
      )}
    >
      <div className="flex h-8 w-8 shrink-0 select-none items-center justify-center rounded-full border border-border/50 bg-background shadow-sm">
        {isUser ? (
          <UserCircle className="h-5 w-5 text-foreground/60" />
        ) : (
          <Bot className="h-5 w-5 text-foreground/60" />
        )}
      </div>
      <div className={cn("flex w-full max-w-[80%] flex-col gap-1", isUser ? "items-end" : "items-start")}>
        <div
          className={cn(
            "rounded-2xl px-4 py-2 text-sm",
            isUser
              ? "bg-primary text-primary-foreground"
              : "bg-muted text-foreground/90"
          )}
        >
          {message.content}
        </div>
        <span className="px-2 text-xs text-muted-foreground">
          {new Date(message.timestamp).toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
          })}
        </span>
      </div>
    </div>
  )
} 