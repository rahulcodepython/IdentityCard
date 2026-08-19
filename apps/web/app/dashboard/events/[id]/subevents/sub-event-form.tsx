"use client"

import { type FormEvent, useMemo, useState, useTransition } from "react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { FixedRangeEditor } from "@/components/schedule/fixed-range-editor"
import { SelectiveDaysEditor } from "@/components/schedule/selective-days-editor"
import type { DayEntry } from "@/components/schedule/types"
import type { EventDay } from "@/lib/validation/events"
import type { SubEventScheduleMode } from "@/lib/validation/subevents"

const MODES: { value: SubEventScheduleMode; label: string; description: string }[] = [
  { value: "selective", label: "Selective dates", description: "Pick specific days from the parent event" },
  { value: "fixed_range", label: "Fixed range", description: "A date range within the parent event, one time window" },
]

// Shared by the create and edit sub-event pages. A sub-event always runs
// on a subset of the parent event's own materialized days — both modes
// here constrain their calendars to `eventDays`, matching the server-side
// validation in subevents.Service.
export function SubEventForm({
  eventDays,
  defaultValues,
  action,
  submitLabel,
}: {
  eventDays: EventDay[]
  defaultValues?: { name: string; scheduleMode: SubEventScheduleMode; days: EventDay[] }
  // Untyped input: create needs schedule_mode, update doesn't — this form
  // constructs whichever shape applies (see onSubmit) and the server
  // action validates it with the matching zod schema either way.
  action: (input: Record<string, unknown>) => Promise<{ error: string } | undefined>
  submitLabel: string
}) {
  const [serverError, setServerError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  const [name, setName] = useState(defaultValues?.name ?? "")
  const [nameError, setNameError] = useState<string | null>(null)
  const [scheduleMode, setScheduleMode] = useState<SubEventScheduleMode>(
    defaultValues?.scheduleMode ?? "selective"
  )
  const fixedMode = !!defaultValues // schedule_mode can't change after creation

  const allowedDates = useMemo(() => new Set(eventDays.map((d) => d.date)), [eventDays])

  const [days, setDays] = useState<DayEntry[]>(defaultValues?.days ?? [])
  const parentRange = eventDays.length > 0 ? [eventDays[0].date, eventDays[eventDays.length - 1].date] : ["", ""]
  const [rangeStart, setRangeStart] = useState(
    defaultValues?.scheduleMode === "fixed_range" && defaultValues.days[0]
      ? defaultValues.days[0].date
      : parentRange[0]
  )
  const [rangeEnd, setRangeEnd] = useState(
    defaultValues?.scheduleMode === "fixed_range" && defaultValues.days.length > 0
      ? defaultValues.days[defaultValues.days.length - 1].date
      : parentRange[1]
  )
  const [rangeEntryTime, setRangeEntryTime] = useState(
    defaultValues?.days[0]?.entry_time ?? eventDays[0]?.entry_time ?? "09:00"
  )
  const [rangeExitTime, setRangeExitTime] = useState(
    defaultValues?.days[0]?.exit_time ?? eventDays[0]?.exit_time ?? "17:00"
  )

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (name.trim().length < 2) {
      setNameError("Enter a sub-event name")
      return
    }
    setNameError(null)
    setServerError(null)

    const base = {
      name,
      days: scheduleMode === "selective" ? days : [],
      range_start: scheduleMode === "fixed_range" ? rangeStart : "",
      range_end: scheduleMode === "fixed_range" ? rangeEnd : "",
      range_entry_time: scheduleMode === "fixed_range" ? rangeEntryTime : "",
      range_exit_time: scheduleMode === "fixed_range" ? rangeExitTime : "",
    }
    const input = fixedMode ? base : { ...base, schedule_mode: scheduleMode }

    startTransition(async () => {
      const result = await action(input)
      if (result?.error) setServerError(result.error)
    })
  }

  return (
    <form onSubmit={onSubmit} className="flex max-w-2xl flex-col gap-4" noValidate>
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          aria-invalid={!!nameError}
        />
        {nameError && <p className="text-xs text-destructive">{nameError}</p>}
      </div>

      <div className="flex flex-col gap-2">
        <Label>Schedule</Label>
        {fixedMode ? (
          <p className="text-sm text-muted-foreground">
            {MODES.find((m) => m.value === scheduleMode)?.label} — can&apos;t be changed after creation
          </p>
        ) : (
          <div className="grid grid-cols-2 gap-2">
            {MODES.map((m) => (
              <button
                key={m.value}
                type="button"
                onClick={() => setScheduleMode(m.value)}
                className={
                  "flex flex-col items-start gap-0.5 rounded-lg border p-3 text-left text-sm transition-colors " +
                  (scheduleMode === m.value ? "border-primary ring-1 ring-primary" : "hover:bg-muted")
                }
              >
                <span className="font-medium">{m.label}</span>
                <span className="text-xs text-muted-foreground">{m.description}</span>
              </button>
            ))}
          </div>
        )}
      </div>

      {scheduleMode === "selective" && (
        <SelectiveDaysEditor value={days} onChange={setDays} allowedDates={allowedDates} />
      )}
      {scheduleMode === "fixed_range" && (
        <FixedRangeEditor
          rangeStart={rangeStart}
          rangeEnd={rangeEnd}
          entryTime={rangeEntryTime}
          exitTime={rangeExitTime}
          onRangeChange={(start, end) => {
            setRangeStart(start)
            setRangeEnd(end)
          }}
          onTimeChange={(entry, exit) => {
            setRangeEntryTime(entry)
            setRangeExitTime(exit)
          }}
          excludedDates={[]}
          onExcludedDatesChange={() => {}}
          allowedDates={allowedDates}
          supportsExclusions={false}
        />
      )}

      {serverError && <p className="text-sm text-destructive">{serverError}</p>}

      <Button type="submit" disabled={isPending} className="mt-2 self-start">
        {isPending ? "Saving…" : submitLabel}
      </Button>
    </form>
  )
}
