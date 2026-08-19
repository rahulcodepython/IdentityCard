"use client"

import { useState, useTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"

import { publishEvent } from "@/lib/client-api/events"
import { queryKeys } from "@/react-query/query-keys"

export function PublishButton({ eventId }: { eventId: string }) {
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  return (
    <div className="flex flex-col items-end gap-1">
      <Button
        disabled={isPending}
        onClick={() =>
          startTransition(async () => {
            try {
              await publishEvent(eventId)
            } catch (err: any) {
              setError(err.message ?? "Something went wrong.")
              return
            }
            setError(null)
            queryClient.invalidateQueries({ queryKey: queryKeys.event(eventId) })
          })
        }
      >
        {isPending ? "Publishing…" : "Publish"}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}