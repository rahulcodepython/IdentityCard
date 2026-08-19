"use client"

import { useState, useTransition } from "react"
import { useRouter } from "next/navigation"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

import { addExcludedDateAction } from "./actions"

// The one schedule edit allowed on a recurring event regardless of
// draft/published status — see events.Service.AddExcludedDate.
export function AddExclusionForm({ eventId }: { eventId: string }) {
  const router = useRouter()
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
            const result = await addExcludedDateAction(eventId, date)
            if (result?.error) {
              setError(result.error)
              return
            }
            setError(null)
            setDate("")
            router.refresh()
          })
        }
      >
        {isPending ? "Adding…" : "Add exclusion"}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
