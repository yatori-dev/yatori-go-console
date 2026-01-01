import { redirect } from "next/navigation"
import { AccountList } from "@/components/account-list"

export default function HomePage() {
  redirect("/accounts")

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-card">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <span className="font-semibold text-lg">Y</span>
              </div>
              <div>
                <h1 className="text-xl font-semibold text-foreground">Yatori 课程管理系统</h1>
                <p className="text-sm text-muted-foreground">在线课程管理控制台</p>
              </div>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-6 py-8">
        <AccountList />
      </main>
    </div>
  )
}
