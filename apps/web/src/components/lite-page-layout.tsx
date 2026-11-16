import type { ReactNode } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export function LitePageLayout({
  title,
  description,
  action,
  children,
}: {
  title: string
  description?: string
  action?: ReactNode
  children: ReactNode
}) {
  return (
    <div className="h-full min-h-0 overflow-y-auto bg-background">
      <main className="mx-auto w-full max-w-5xl space-y-5 px-4 py-5 sm:px-7 sm:py-7">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="space-y-1">
            {/* The unified app toolbar already displays the page title. */}
            <h1 className="sr-only">{title}</h1>
            {description && <p className="text-sm leading-relaxed text-muted-foreground">{description}</p>}
          </div>
          {action}
        </div>
        {children}
      </main>
    </div>
  )
}

export function LiteSection({
  title,
  description,
  action,
  children,
}: {
  title: string
  description?: string
  action?: ReactNode
  children: ReactNode
}) {
  return (
    <Card className="border-border shadow-sm">
      <CardHeader className="flex flex-row flex-wrap items-center justify-between gap-3 border-b px-4 py-4 sm:px-5">
        <div className="min-w-0 space-y-1">
          <CardTitle className="text-base font-semibold">{title}</CardTitle>
          {description && <p className="text-xs leading-relaxed text-muted-foreground">{description}</p>}
        </div>
        {action}
      </CardHeader>
      <CardContent className="px-4 py-4 sm:px-5">{children}</CardContent>
    </Card>
  )
}
