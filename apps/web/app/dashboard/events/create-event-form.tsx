"use client"

import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useRouter } from "next/navigation"
import { createEvent } from "@/lib/client-api/events"
import { queryKeys } from "@/react-query/query-keys"

import Link from "next/link"
import { type FormEvent, useState, useTransition } from "react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { FixedRangeEditor } from "@/components/schedule/fixed-range-editor"
import { ModeSelect } from "@/components/schedule/mode-select"
import { RecurringEditor } from "@/components/schedule/recurring-editor"
import { SelectiveDaysEditor } from "@/components/schedule/selective-days-editor"
import type { DayEntry, RecurrenceValue } from "@/components/schedule/types"
import type { CreateEventInput, ScheduleMode } from "@/lib/validation/events"

const PLAN_ERROR_CODES = new Set(["plan_required", "event_limit_reached"])

export function CreateEventForm({ allowFlash }: { allowFlash: boolean }) {
  const [serverError, setServerError] = useState<string | null>(null)
  const [planErrorCode, setPlanErrorCode] = useState<string | null>(null)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string> | null>(null)
  const [isPending, startTransition] = useTransition()

  const [name, setName] = useState("")
  const [nameError, setNameError] = useState<string | null>(null)
  const [venue, setVenue] = useState("")
  const [scheduleMode, setScheduleMode] = useState<ScheduleMode>(allowFlash ? "flash" : "selective")

  const [days, setDays] = useState<DayEntry[]>([])
  const [rangeStart, setRangeStart] = useState("")
  const [rangeEnd, setRangeEnd] = useState("")
  const [rangeEntryTime, setRangeEntryTime] = useState("09:00")
  const [rangeExitTime, setRangeExitTime] = useState("17:00")
  const [excludedDates, setExcludedDates] = useState<string[]>([])
  const [recurrence, setRecurrence] = useState<RecurrenceValue>({
    starts_on: "",
    ends_on: null,
    weekdays: [],
  })

  const queryClient = useQueryClient()
  const router = useRouter()

  const createMutation = useMutation({
    mutationFn: createEvent,
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.events() })
      router.push(`/dashboard/events/${result.id}`)
    },
    onError: (err: any) => {
      setServerError(err.message || "Failed to create event")
      if (err.code && PLAN_ERROR_CODES.has(err.code)) {
        setPlanErrorCode(err.code)
      }
      if (err.fields) {
        setFieldErrors(err.fields)
      }
    }
  })

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (name.trim().length < 2) {
      setNameError("Enter an event name")
      return
    }
    setNameError(null)
    setServerError(null)
    setPlanErrorCode(null)
    setFieldErrors(null)

    const input: CreateEventInput = {
      name,
      schedule_mode: scheduleMode,
      venue,
      days,
      range_start: rangeStart,
      range_end: rangeEnd,
      range_entry_time: rangeEntryTime,
      range_exit_time: rangeExitTime,
      excluded_dates: excludedDates,
      recurrence: scheduleMode === "recurring" ? recurrence : null,
    }

    createMutation.mutate(input)
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
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

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="venue">Venue (optional)</Label>
        <Input id="venue" value={venue} onChange={(e) => setVenue(e.target.value)} />
      </div>

      <div className="flex flex-col gap-2">
        <Label>Schedule</Label>
        <ModeSelect value={scheduleMode} onChange={setScheduleMode} allowFlash={allowFlash} />
      </div>

      {(scheduleMode === "flash" || scheduleMode === "selective") && (
        <SelectiveDaysEditor value={days} onChange={setDays} singleDay={scheduleMode === "flash"} />
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
          excludedDates={excludedDates}
          onExcludedDatesChange={setExcludedDates}
        />
      )}
      {scheduleMode === "recurring" && (
        <RecurringEditor
          value={recurrence}
          onChange={setRecurrence}
          excludedDates={excludedDates}
          onExcludedDatesChange={setExcludedDates}
        />
      )}

      {fieldErrors && Object.keys(fieldErrors).length > 0 && (
        <ul className="list-disc pl-4 text-xs text-destructive">
          {Object.entries(fieldErrors).map(([field, message]) => (
            <li key={field}>
              {field}: {message}
            </li>
          ))}
        </ul>
      )}

      {serverError && (
        <p className="text-sm text-destructive">
          {serverError}
          {planErrorCode && (
            <>
              {" "}
              <Link href="/dashboard/billing" className="underline underline-offset-4">
                {planErrorCode === "plan_required" ? "Choose a plan" : "Upgrade your plan"}
              </Link>
            </>
          )}
        </p>
      )}

      <Button type="submit" disabled={isPending} className="mt-2 self-start">
        {isPending ? "Creating…" : "Create event"}
      </Button>
    </form>
  )
}
