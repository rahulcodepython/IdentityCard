"use client"

import { useState, useTransition } from "react"
import { useForm } from "react-hook-form"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { EventDay } from "@/lib/validation/events"
import type { CreateSubEventInput } from "@/lib/validation/subevents"

type DayFieldValue = {
  selected: boolean
  date: string
  entry_time: string
  exit_time: string
}
type FormValues = { name: string; days: DayFieldValue[] }

// Shared by the create and edit sub-event pages: a fixed checklist of the
// parent event's own days (see events.Service.GetForSubEvent — a
// sub-event's days must be a subset of these), each with its own
// entry/exit override, pre-filled from the parent's times.
export function SubEventForm({
  eventDays,
  defaultValues,
  action,
  submitLabel,
}: {
  eventDays: EventDay[]
  defaultValues?: { name: string; days: EventDay[] }
  action: (input: CreateSubEventInput) => Promise<{ error: string } | undefined>
  submitLabel: string
}) {
  const [serverError, setServerError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  const selectedByDate = new Map(
    (defaultValues?.days ?? []).map((day) => [day.date, day])
  )

  const { register, handleSubmit } = useForm<FormValues>({
    defaultValues: {
      name: defaultValues?.name ?? "",
      days: eventDays.map((day) => {
        const override = selectedByDate.get(day.date)
        return {
          selected: !!override,
          date: day.date,
          entry_time: override?.entry_time ?? day.entry_time,
          exit_time: override?.exit_time ?? day.exit_time,
        }
      }),
    },
  })

  const onSubmit = handleSubmit((values) => {
    setServerError(null)
    startTransition(async () => {
      const result = await action({
        name: values.name,
        days: values.days
          .filter((day) => day.selected)
          .map(({ date, entry_time, exit_time }) => ({
            date,
            entry_time,
            exit_time,
          })),
      })
      if (result?.error) setServerError(result.error)
    })
  })

  return (
    <form
      onSubmit={onSubmit}
      className="flex max-w-lg flex-col gap-4"
      noValidate
    >
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          {...register("name", { required: true, minLength: 2 })}
        />
      </div>

      <div className="flex flex-col gap-2">
        <Label>Days</Label>
        <p className="text-xs text-muted-foreground">
          Leave every day unchecked to run this sub-event on the whole
          event&apos;s schedule.
        </p>
        {eventDays.map((day, index) => (
          <div
            key={day.date}
            className="flex items-center gap-3 rounded-lg border p-2"
          >
            <input
              type="checkbox"
              className="size-4"
              {...register(`days.${index}.selected`)}
            />
            <span className="w-28 text-sm">{day.date}</span>
            <Input
              type="time"
              className="w-28"
              {...register(`days.${index}.entry_time`)}
            />
            <span className="text-sm text-muted-foreground">–</span>
            <Input
              type="time"
              className="w-28"
              {...register(`days.${index}.exit_time`)}
            />
          </div>
        ))}
      </div>

      {serverError && <p className="text-sm text-destructive">{serverError}</p>}

      <Button type="submit" disabled={isPending} className="mt-2 self-start">
        {isPending ? "Saving…" : submitLabel}
      </Button>
    </form>
  )
}
