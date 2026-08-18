"use client"

import { useState, useTransition } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useFieldArray, useForm } from "react-hook-form"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  type EventKind,
  type UpdateEventInput,
  updateEventSchema,
} from "@/lib/validation/events"

import { updateEventAction } from "./actions"

export function EditEventForm({
  eventId,
  kind,
  defaultValues,
}: {
  eventId: string
  kind: EventKind
  defaultValues: UpdateEventInput
}) {
  const [serverError, setServerError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<UpdateEventInput>({
    resolver: zodResolver(updateEventSchema),
    defaultValues,
  })
  const { fields, append, remove } = useFieldArray({ control, name: "days" })
  const canAddDay = kind !== "flash" || fields.length === 0

  const onSubmit = handleSubmit((values) => {
    setServerError(null)
    startTransition(async () => {
      const result = await updateEventAction(eventId, values)
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
        <Input id="name" aria-invalid={!!errors.name} {...register("name")} />
        {errors.name && (
          <p className="text-xs text-destructive">{errors.name.message}</p>
        )}
      </div>

      <div className="flex flex-col gap-1.5">
        <Label>Type</Label>
        <p className="text-sm text-muted-foreground">
          {kind === "flash" ? "Flash (single day)" : "Established"} — can&apos;t
          be changed after creation
        </p>
      </div>

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="venue">Venue (optional)</Label>
        <Input id="venue" {...register("venue")} />
      </div>

      <div className="flex flex-col gap-3">
        <Label>Days</Label>
        {fields.map((field, index) => (
          <div
            key={field.id}
            className="grid grid-cols-[1fr_1fr_1fr_auto] items-end gap-2"
          >
            <div className="flex flex-col gap-1">
              <Label className="text-xs text-muted-foreground">Date</Label>
              <Input type="date" {...register(`days.${index}.date`)} />
            </div>
            <div className="flex flex-col gap-1">
              <Label className="text-xs text-muted-foreground">Entry</Label>
              <Input type="time" {...register(`days.${index}.entry_time`)} />
            </div>
            <div className="flex flex-col gap-1">
              <Label className="text-xs text-muted-foreground">Exit</Label>
              <Input type="time" {...register(`days.${index}.exit_time`)} />
            </div>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => remove(index)}
              aria-label="Remove day"
            >
              ×
            </Button>
          </div>
        ))}
        {errors.days?.message && (
          <p className="text-xs text-destructive">{errors.days.message}</p>
        )}
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="self-start"
          disabled={!canAddDay}
          onClick={() => append({ date: "", entry_time: "", exit_time: "" })}
        >
          Add day
        </Button>
      </div>

      {serverError && <p className="text-sm text-destructive">{serverError}</p>}

      <Button type="submit" disabled={isPending} className="mt-2 self-start">
        {isPending ? "Saving…" : "Save changes"}
      </Button>
    </form>
  )
}
