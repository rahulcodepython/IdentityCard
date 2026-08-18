"use client"

import { useActionState } from "react"

import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import type { SubEvent } from "@/lib/validation/subevents"

import { type ImportActionResult, importCsvAction } from "./actions"

export function ImportForm({
  eventId,
  subEvents,
}: {
  eventId: string
  subEvents: SubEvent[]
}) {
  const [state, formAction, isPending] = useActionState<
    ImportActionResult | null,
    FormData
  >((_prevState, formData) => importCsvAction(eventId, formData), null)

  return (
    <div className="flex flex-col gap-6">
      <form action={formAction} className="flex max-w-lg flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="file">CSV file</Label>
          <input
            id="file"
            name="file"
            type="file"
            accept=".csv,text/csv"
            required
            className="text-sm"
          />
          <p className="text-xs text-muted-foreground">
            Required columns: email, mobile, name. Optional: image_url, age,
            gender. Uploading again for an existing email+mobile updates their
            details rather than duplicating them.
          </p>
        </div>

        {subEvents.length > 0 && (
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="sub_event_id">Assign to sub-event</Label>
            <select
              id="sub_event_id"
              name="sub_event_id"
              className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
            >
              <option value="">None — whole event</option>
              {subEvents.map((subEvent) => (
                <option key={subEvent.id} value={subEvent.id}>
                  {subEvent.name}
                </option>
              ))}
            </select>
          </div>
        )}

        <Button type="submit" disabled={isPending} className="self-start">
          {isPending ? "Importing…" : "Import"}
        </Button>
      </form>

      {state && "error" in state && (
        <p className="text-sm text-destructive">{state.error}</p>
      )}

      {state && "summary" in state && (
        <div className="flex flex-col gap-2 rounded-xl border p-4 text-sm">
          <p>
            Inserted {state.summary.inserted}, updated {state.summary.updated},
            skipped {state.summary.skipped}.
          </p>
          {state.summary.errors.length > 0 && (
            <ul className="list-disc pl-4 text-muted-foreground">
              {state.summary.errors.map((rowError) => (
                <li key={rowError.row}>
                  Row {rowError.row}: {rowError.message}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
