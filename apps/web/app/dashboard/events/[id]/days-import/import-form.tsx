"use client"

import { useActionState } from "react"

import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"

import {
  type ImportActionResult,
  importDaysCsvAction,
  importExclusionsCsvAction,
} from "./actions"

function ImportResult({ state }: { state: ImportActionResult | null }) {
  if (!state) return null
  if ("error" in state) {
    return <p className="text-sm text-destructive">{state.error}</p>
  }
  return (
    <div className="flex flex-col gap-2 rounded-xl border p-4 text-sm">
      <p>
        Imported {state.summary.imported}, skipped {state.summary.skipped}.
      </p>
      {state.summary.errors.length > 0 && (
        <ul className="list-disc pl-4 text-muted-foreground">
          {state.summary.errors.map((rowError, i) => (
            <li key={i}>
              Row {rowError.row}: {rowError.message}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

// Days import (date,entry_time,exit_time columns) wholesale-replaces the
// day list — only for fixed_range/selective draft events. Exclusions
// import (date column) adds each row as an exclusion — for fixed_range
// (draft) or recurring (any status, including after publish).
export function ImportForm({
  eventId,
  showDaysImport,
  showExclusionsImport,
}: {
  eventId: string
  showDaysImport: boolean
  showExclusionsImport: boolean
}) {
  const [daysState, daysFormAction, daysPending] = useActionState<
    ImportActionResult | null,
    FormData
  >((_prev, formData) => importDaysCsvAction(eventId, formData), null)
  const [exclusionsState, exclusionsFormAction, exclusionsPending] = useActionState<
    ImportActionResult | null,
    FormData
  >((_prev, formData) => importExclusionsCsvAction(eventId, formData), null)

  return (
    <div className="flex flex-col gap-8">
      {showDaysImport && (
        <div className="flex flex-col gap-4">
          <h2 className="font-heading text-lg font-medium">Import days</h2>
          <form action={daysFormAction} className="flex max-w-lg flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="days-file">CSV file</Label>
              <input
                id="days-file"
                name="file"
                type="file"
                accept=".csv,text/csv"
                required
                className="text-sm"
              />
              <p className="text-xs text-muted-foreground">
                Required columns: date, entry_time, exit_time. Replaces the entire day list.
              </p>
            </div>
            <Button type="submit" disabled={daysPending} className="self-start">
              {daysPending ? "Importing…" : "Import days"}
            </Button>
          </form>
          <ImportResult state={daysState} />
        </div>
      )}

      {showExclusionsImport && (
        <div className="flex flex-col gap-4">
          <h2 className="font-heading text-lg font-medium">Import exclusion dates</h2>
          <form action={exclusionsFormAction} className="flex max-w-lg flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="exclusions-file">CSV file</Label>
              <input
                id="exclusions-file"
                name="file"
                type="file"
                accept=".csv,text/csv"
                required
                className="text-sm"
              />
              <p className="text-xs text-muted-foreground">
                Required column: date. Adds each date as an exclusion (holiday) —
                never removes existing ones.
              </p>
            </div>
            <Button type="submit" disabled={exclusionsPending} className="self-start">
              {exclusionsPending ? "Importing…" : "Import exclusions"}
            </Button>
          </form>
          <ImportResult state={exclusionsState} />
        </div>
      )}
    </div>
  )
}
