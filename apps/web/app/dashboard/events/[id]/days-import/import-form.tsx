"use client"

import { useState, useTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"

import { importDaysCsv, importExcludedDatesCsv } from "@/lib/client-api/events"
import { queryKeys } from "@/react-query/query-keys"
import type { DayImportSummary } from "@/schema/events.types"

type ImportState = { summary: DayImportSummary } | { error: string } | null

function ImportResult({ state }: { state: ImportState }) {
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
  const queryClient = useQueryClient()
  const [daysState, setDaysState] = useState<ImportState>(null)
  const [exclusionsState, setExclusionsState] = useState<ImportState>(null)
  const [daysPending, startDaysTransition] = useTransition()
  const [exclusionsPending, startExclusionsTransition] = useTransition()

  function submitDays(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    setDaysState(null)
    startDaysTransition(async () => {
      try {
        const summary = await importDaysCsv(eventId, formData)
        setDaysState({ summary })
        queryClient.invalidateQueries({ queryKey: queryKeys.event(eventId) })
      } catch (err: any) {
        setDaysState({ error: err.message ?? "Something went wrong." })
      }
    })
  }

  function submitExclusions(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    setExclusionsState(null)
    startExclusionsTransition(async () => {
      try {
        const summary = await importExcludedDatesCsv(eventId, formData)
        setExclusionsState({ summary })
        queryClient.invalidateQueries({ queryKey: queryKeys.event(eventId) })
      } catch (err: any) {
        setExclusionsState({ error: err.message ?? "Something went wrong." })
      }
    })
  }

  return (
    <div className="flex flex-col gap-8">
      {showDaysImport && (
        <div className="flex flex-col gap-4">
          <h2 className="font-heading text-lg font-medium">Import days</h2>
          <form onSubmit={submitDays} className="flex max-w-lg flex-col gap-4">
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
          <form onSubmit={submitExclusions} className="flex max-w-lg flex-col gap-4">
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