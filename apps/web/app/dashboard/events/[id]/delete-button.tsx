"use client"

import { useState, useTransition } from "react"

import { Button } from "@/components/ui/button"

import { deleteEventAction } from "./actions"

export function DeleteButton({ eventId }: { eventId: string }) {
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  return (
    <div className="flex flex-col items-end gap-1">
      <Button
        variant="destructive"
        disabled={isPending}
        onClick={() =>
          startTransition(async () => {
            const result = await deleteEventAction(eventId)
            if (result?.error) setError(result.error)
          })
        }
      >
        {isPending ? "Deleting…" : "Delete"}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
