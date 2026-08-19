"use client"

import { type FormEvent, useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useRouter } from "next/navigation"
import { updateEvent } from "@/lib/client-api/events"
import { queryKeys } from "@/react-query/query-keys"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { FixedRangeEditor } from "@/components/schedule/fixed-range-editor"
import { RecurringEditor } from "@/components/schedule/recurring-editor"
import { SelectiveDaysEditor } from "@/components/schedule/selective-days-editor"
import type { DayEntry, RecurrenceValue } from "@/components/schedule/types"
import type { EventDetail, UpdateEventInput } from "@/lib/validation/events"



const SCHEDULE_MODE_LABEL: Record<EventDetail["schedule_mode"], string> = {
  flash: "Flash (single day) — can't be changed after creation",
  fixed_range: "Fixed range — can't be changed after creation",
  selective: "Selective dates — can't be changed after creation",
  recurring: "Recurring — can't be changed after creation",
}

export function EditEventForm({
  eventId,
  event,
}: {
  eventId: string
  event: EventDetail
}) {
  const [serverError, setServerError] = useState<string | null>(null)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string> | null>(null)
  const [isPending, setIsPending] = useState(false)

  const [name, setName] = useState(event.name)
  const [nameError, setNameError] = useState<string | null>(null)
  const [venue, setVenue] = useState(event.venue ?? "")

  const [days, setDays] = useState<DayEntry[]>(event.days)
  const [rangeStart, setRangeStart] = useState(event.start_date)
  const [rangeEnd, setRangeEnd] = useState(event.end_date ?? event.start_date)
  const [rangeEntryTime, setRangeEntryTime] = useState(event.days[0]?.entry_time ?? "09:00")
  const [rangeExitTime, setRangeExitTime] = useState(event.days[0]?.exit_time ?? "17:00")
  const [excludedDates, setExcludedDates] = useState<string[]>(event.excluded_dates ?? [])
  const [recurrence, setRecurrence] = useState<RecurrenceValue>(
    event.recurrence ?? { starts_on: event.start_date, ends_on: event.end_date, weekdays: [] }
  )

  const queryClient = useQueryClient()
  const router = useRouter()

  const updateMutation = useMutation({
    mutationFn: (data: { id: string, input: UpdateEventInput }) => updateEvent(data.id, data.input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.events() })
      queryClient.invalidateQueries({ queryKey: queryKeys.event(eventId) })
      router.push(`/dashboard/events/${eventId}`)
    },
    onError: (err: any) => {
      setServerError(err.message || "Failed to update event")
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
    setFieldErrors(null)

    const input: UpdateEventInput = {
      name,
      venue,
      days,
      range_start: rangeStart,
      range_end: rangeEnd,
      range_entry_time: rangeEntryTime,
      range_exit_time: rangeExitTime,
      excluded_dates: excludedDates,
      recurrence: event.schedule_mode === "recurring" ? recurrence : null,
    }

    updateMutation.mutate({ id: eventId, input })
  }

  return (
    <form onSubmit={onSubmit} className="flex max-w-3xl flex-col gap-4" noValidate>
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
        <Label>Schedule type</Label>
        <p className="text-sm text-muted-foreground">
          {SCHEDULE_MODE_LABEL[event.schedule_mode]}
        </p>
      </div>

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="venue">Venue (optional)</Label>
        <Input id="venue" value={venue} onChange={(e) => setVenue(e.target.value)} />
      </div>

      {(event.schedule_mode === "flash" || event.schedule_mode === "selective") && (
        <SelectiveDaysEditor
          value={days}
          onChange={setDays}
          singleDay={event.schedule_mode === "flash"}
        />
      )}
      {event.schedule_mode === "fixed_range" && (
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
      {event.schedule_mode === "recurring" && (
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

      {serverError && <p className="text-sm text-destructive">{serverError}</p>}

      <Button type="submit" disabled={updateMutation.isPending} className="mt-2 self-start">
        {updateMutation.isPending ? "Saving…" : "Save changes"}
      </Button>
    </form>
  )
}
