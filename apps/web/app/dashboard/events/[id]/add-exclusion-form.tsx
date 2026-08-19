"use client"

import { useState, useTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

import { addExcludedDate } from "@/lib/client-api/events"
import { queryKeys } from "@/react-query/query-keys"

// The one schedule edit allowed on a recurring event regardless of
// draft/published status — see events.Service.AddExcludedDate.
export function AddExclusionForm({ eventId }: { eventId: string }) {
  const queryClient = useQueryClient()
  const [date, setDate] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  return (
    <div className="flex items-end gap-2">
      <div className="flex flex-col gap-1">
        <Input
          type="date"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          className="h-8"
        />
      </div>
      <Button
        type="button"
        variant="outline"
        size="sm"
        disabled={!date || isPending}
        onClick={() =>
          startTransition(async () => {
            try {
              await addExcludedDate(eventId, { date })
            } catch (err: any) {
              setError(err.message ?? "Something went wrong.")
              return
            }
            setError(null)
            setDate("")
            queryClient.invalidateQueries({ queryKey: queryKeys.event(eventId) })
          })
        }
      >
        {isPending ? "Adding…" : "Add exclusion"}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}