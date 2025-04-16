import { useState } from "react"
import { Button } from "./ui/button"
import { useChatStore } from "../store/chat"
import { SendIcon } from "lucide-react"

export function ChatInput() {
  const [input, setInput] = useState("")
  const addMessage = useChatStore((state) => state.addMessage)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim()) return

    addMessage({ content: input, role: "user" })
    setInput("")

    // TODO: Add API call to backend for RAG response
  }

  return (
    <form onSubmit={handleSubmit} className="flex w-full items-center gap-2 p-4">
      <input
        type="text"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        placeholder="Type your message..."
        className="flex h-10 w-full rounded-full border-0 bg-muted px-4 py-2 text-sm text-foreground/90 ring-offset-background placeholder:text-muted-foreground/50 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
      />
      <Button type="submit" size="icon" variant="ghost" className="h-10 w-10 rounded-full">
        <SendIcon className="h-5 w-5" />
      </Button>
    </form>
  )
} 