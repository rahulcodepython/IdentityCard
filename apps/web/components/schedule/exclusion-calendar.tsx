"use client"

import { Calendar } from "@/components/ui/calendar"
import { Label } from "@/components/ui/label"

import { fromDateKey, toDateKey } from "./date-utils"

// Shared by fixed-range and recurring editors: a date that would
// otherwise be active but is explicitly skipped (a holiday in the middle
// of an office's term, etc).
export function ExclusionCalendar({
  dates,
  onChange,
  label = "Excluded dates (holidays)",
}: {
  dates: string[]
  onChange: (dates: string[]) => void
  label?: string
}) {
  return (
    <div className="flex flex-col gap-2">
      <Label>{label}</Label>
      <p className="text-xs text-muted-foreground">
        Click a date to exclude it even on a day that would otherwise be active.
      </p>
      <Calendar
        mode="multiple"
        selected={dates.map(fromDateKey)}
        onSelect={(selected) => onChange((selected ?? []).map(toDateKey).sort())}
        className="rounded-lg border"
      />
      {dates.length > 0 && (
        <p className="text-xs text-muted-foreground">{dates.join(", ")}</p>
      )}
    </div>
  )
}
