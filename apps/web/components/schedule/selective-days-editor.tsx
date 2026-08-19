"use client"

import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

import { fromDateKey, toDateKey } from "./date-utils"
import type { DayEntry } from "./types"

const DEFAULT_ENTRY = "09:00"
const DEFAULT_EXIT = "17:00"

// Click-to-toggle day selection (a full calendar, per the spec) with a
// per-date entry/exit time editable inline, plus a bulk "copy first time
// to all" action — the practical form of "copy/paste a time value" onto
// many selected days at once. `singleDay` renders a single-select
// calendar instead (used for flash events, which are always exactly one
// day). `allowedDates`, when given, restricts selection to a sub-event's
// parent event days.
export function SelectiveDaysEditor({
  value,
  onChange,
  allowedDates,
  singleDay = false,
}: {
  value: DayEntry[]
  onChange: (days: DayEntry[]) => void
  allowedDates?: Set<string>
  singleDay?: boolean
}) {
  const byDate = new Map(value.map((d) => [d.date, d]))
  const disabled = allowedDates ? (date: Date) => !allowedDates.has(toDateKey(date)) : undefined

  function reconcile(dates: Date[]) {
    const next = dates.map((d) => {
      const key = toDateKey(d)
      return byDate.get(key) ?? { date: key, entry_time: DEFAULT_ENTRY, exit_time: DEFAULT_EXIT }
    })
    next.sort((a, b) => a.date.localeCompare(b.date))
    onChange(next)
  }

  function updateDay(date: string, field: "entry_time" | "exit_time", val: string) {
    onChange(value.map((d) => (d.date === date ? { ...d, [field]: val } : d)))
  }

  function removeDay(date: string) {
    onChange(value.filter((d) => d.date !== date))
  }

  function copyFirstTimeToAll() {
    if (value.length === 0) return
    const { entry_time, exit_time } = value[0]
    onChange(value.map((d) => ({ ...d, entry_time, exit_time })))
  }

  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-start">
      {singleDay ? (
        <Calendar
          mode="single"
          selected={value[0] ? fromDateKey(value[0].date) : undefined}
          onSelect={(date) => reconcile(date ? [date] : [])}
          disabled={disabled}
          className="rounded-lg border"
        />
      ) : (
        <Calendar
          mode="multiple"
          selected={value.map((d) => fromDateKey(d.date))}
          onSelect={(dates) => reconcile(dates ?? [])}
          disabled={disabled}
          className="rounded-lg border"
        />
      )}

      <div className="flex min-w-0 flex-1 flex-col gap-2">
        <div className="flex items-center justify-between">
          <Label>Selected dates ({value.length})</Label>
          {!singleDay && value.length > 1 && (
            <Button type="button" variant="ghost" size="sm" onClick={copyFirstTimeToAll}>
              Copy first time to all
            </Button>
          )}
        </div>
        {value.length === 0 && (
          <p className="text-sm text-muted-foreground">
            Click a date on the calendar to add it.
          </p>
        )}
        <div className="flex max-h-72 flex-col gap-2 overflow-y-auto">
          {value.map((d) => (
            <div key={d.date} className="flex items-center gap-2 rounded-lg border p-2">
              <span className="w-24 shrink-0 text-sm">{d.date}</span>
              <Input
                type="time"
                className="w-28"
                value={d.entry_time}
                onChange={(e) => updateDay(d.date, "entry_time", e.target.value)}
              />
              <span className="text-sm text-muted-foreground">–</span>
              <Input
                type="time"
                className="w-28"
                value={d.exit_time}
                onChange={(e) => updateDay(d.date, "exit_time", e.target.value)}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => removeDay(d.date)}
                aria-label="Remove date"
              >
                ×
              </Button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
