"use client"

import { useState, useTransition } from "react"
import { useRouter } from "next/navigation"

import { Button } from "@/components/ui/button"

import { publishEventAction } from "./actions"

export function PublishButton({ eventId }: { eventId: string }) {
  const router = useRouter()
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  return (
    <div className="flex flex-col items-end gap-1">
      <Button
        disabled={isPending}
        onClick={() =>
          startTransition(async () => {
            const result = await publishEventAction(eventId)
            if (result?.error) {
              setError(result.error)
              return
            }
            router.refresh()
          })
        }
      >
        {isPending ? "Publishing…" : "Publish"}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
