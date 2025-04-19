import { Chat } from "./components/chat"
import { ThemeSwitcher } from "@/components/ThemeSwitcher"

function App() {
  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-50 border-b border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container mx-auto flex items-center justify-between h-16 px-4">
          <span className="font-bold text-lg">Minirag</span>
          <ThemeSwitcher />
        </div>
      </header>
      <main className="container mx-auto max-w-4xl p-4 pt-8 md:p-6 md:pt-10">
        <Chat />
      </main>
    </div>
  )
}

export default App
