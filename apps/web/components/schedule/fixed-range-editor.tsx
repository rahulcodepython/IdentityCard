"use client"

import { Calendar } from "@/components/ui/calendar"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

import { fromDateKey, toDateKey } from "./date-utils"
import { ExclusionCalendar } from "./exclusion-calendar"

// A start/end range plus one shared entry/exit time window for every day
// in it — the fast-authoring path for a regular multi-day event.
// `supportsExclusions` is off for sub-events (they have no exclusion
// concept of their own — a range that reaches outside the parent's own
// days is simply rejected server-side). `allowedDates`, when given,
// restricts the range to a sub-event's parent event days.
export function FixedRangeEditor({
  rangeStart,
  rangeEnd,
  entryTime,
  exitTime,
  onRangeChange,
  onTimeChange,
  excludedDates,
  onExcludedDatesChange,
  allowedDates,
  supportsExclusions = true,
}: {
  rangeStart: string
  rangeEnd: string
  entryTime: string
  exitTime: string
  onRangeChange: (start: string, end: string) => void
  onTimeChange: (entry: string, exit: string) => void
  excludedDates: string[]
  onExcludedDatesChange: (dates: string[]) => void
  allowedDates?: Set<string>
  supportsExclusions?: boolean
}) {
  const selected = rangeStart
    ? { from: fromDateKey(rangeStart), to: rangeEnd ? fromDateKey(rangeEnd) : undefined }
    : undefined
  const disabled = allowedDates ? (date: Date) => !allowedDates.has(toDateKey(date)) : undefined

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-2">
        <Label>Date range</Label>
        <Calendar
          mode="range"
          selected={selected}
          onSelect={(range) =>
            onRangeChange(range?.from ? toDateKey(range.from) : "", range?.to ? toDateKey(range.to) : "")
          }
          disabled={disabled}
          className="rounded-lg border"
        />
      </div>

      <div className="flex gap-3">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="range_entry_time">Entry time</Label>
          <Input
            id="range_entry_time"
            type="time"
            value={entryTime}
            onChange={(e) => onTimeChange(e.target.value, exitTime)}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="range_exit_time">Exit time</Label>
          <Input
            id="range_exit_time"
            type="time"
            value={exitTime}
            onChange={(e) => onTimeChange(entryTime, e.target.value)}
          />
        </div>
      </div>

      {supportsExclusions && (
        <ExclusionCalendar dates={excludedDates} onChange={onExcludedDatesChange} />
      )}
    </div>
  )
}
