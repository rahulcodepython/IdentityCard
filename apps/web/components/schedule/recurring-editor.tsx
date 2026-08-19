"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

import { ExclusionCalendar } from "./exclusion-calendar"
import { WEEKDAY_LABELS, type RecurrenceValue } from "./types"

const DEFAULT_ENTRY = "09:00"
const DEFAULT_EXIT = "17:00"

// Weekday selector + per-weekday time (a Friday can close earlier than
// Mon-Thu — deliberately no single shared time for the week), a start
// date, an open-ended toggle that nulls the end date, and exclusion
// dates honored even on an otherwise-active weekday. Only used at the
// top-level event — sub-events don't get their own recurrence rule.
export function RecurringEditor({
  value,
  onChange,
  excludedDates,
  onExcludedDatesChange,
}: {
  value: RecurrenceValue
  onChange: (next: RecurrenceValue) => void
  excludedDates: string[]
  onExcludedDatesChange: (dates: string[]) => void
}) {
  const weekdayMap = new Map(value.weekdays.map((w) => [w.weekday, w]))
  const openEnded = value.ends_on == null

  function toggleWeekday(wd: number) {
    if (weekdayMap.has(wd)) {
      onChange({ ...value, weekdays: value.weekdays.filter((w) => w.weekday !== wd) })
    } else {
      const next = [...value.weekdays, { weekday: wd, entry_time: DEFAULT_ENTRY, exit_time: DEFAULT_EXIT }]
      next.sort((a, b) => a.weekday - b.weekday)
      onChange({ ...value, weekdays: next })
    }
  }

  function updateWeekday(wd: number, field: "entry_time" | "exit_time", val: string) {
    onChange({
      ...value,
      weekdays: value.weekdays.map((w) => (w.weekday === wd ? { ...w, [field]: val } : w)),
    })
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="starts_on">Start date</Label>
        <Input
          id="starts_on"
          type="date"
          value={value.starts_on}
          onChange={(e) => onChange({ ...value, starts_on: e.target.value })}
        />
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="open_ended"
          className="size-4"
          checked={openEnded}
          onChange={(e) => onChange({ ...value, ends_on: e.target.checked ? null : value.starts_on })}
        />
        <Label htmlFor="open_ended" className="font-normal">
          Open-ended — runs indefinitely (offices, ongoing programs)
        </Label>
      </div>
      {!openEnded && (
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="ends_on">End date</Label>
          <Input
            id="ends_on"
            type="date"
            value={value.ends_on ?? ""}
            onChange={(e) => onChange({ ...value, ends_on: e.target.value })}
          />
        </div>
      )}

      <div className="flex flex-col gap-2">
        <Label>Active weekdays</Label>
        <div className="flex flex-wrap gap-2">
          {WEEKDAY_LABELS.map((label, wd) => (
            <button
              key={wd}
              type="button"
              onClick={() => toggleWeekday(wd)}
              className={
                "rounded-lg border px-3 py-1.5 text-sm transition-colors " +
                (weekdayMap.has(wd) ? "border-primary bg-primary/10" : "hover:bg-muted")
              }
            >
              {label}
            </button>
          ))}
        </div>
        {value.weekdays.length === 0 && (
          <p className="text-xs text-muted-foreground">Pick at least one active weekday.</p>
        )}
        <div className="flex flex-col gap-2">
          {value.weekdays.map((w) => (
            <div key={w.weekday} className="flex items-center gap-2 rounded-lg border p-2">
              <span className="w-10 shrink-0 text-sm">{WEEKDAY_LABELS[w.weekday]}</span>
              <Input
                type="time"
                className="w-28"
                value={w.entry_time}
                onChange={(e) => updateWeekday(w.weekday, "entry_time", e.target.value)}
              />
              <span className="text-sm text-muted-foreground">–</span>
              <Input
                type="time"
                className="w-28"
                value={w.exit_time}
                onChange={(e) => updateWeekday(w.weekday, "exit_time", e.target.value)}
              />
            </div>
          ))}
        </div>
      </div>

      <ExclusionCalendar dates={excludedDates} onChange={onExcludedDatesChange} />
    </div>
  )
}
