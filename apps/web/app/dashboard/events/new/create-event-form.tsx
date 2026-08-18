"use client"

import { useState, useTransition } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useFieldArray, useForm } from "react-hook-form"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  type CreateEventInput,
  createEventSchema,
} from "@/lib/validation/events"

import { createEventAction } from "./actions"

export function CreateEventForm() {
  const [serverError, setServerError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  const {
    register,
    handleSubmit,
    control,
    watch,
    formState: { errors },
  } = useForm<CreateEventInput>({
    resolver: zodResolver(createEventSchema),
    defaultValues: {
      name: "",
      kind: "established",
      venue: "",
      days: [{ date: "", entry_time: "", exit_time: "" }],
    },
  })
  const { fields, append, remove } = useFieldArray({ control, name: "days" })
  const kind = watch("kind")
  const canAddDay = kind !== "flash" || fields.length === 0

  const onSubmit = handleSubmit((values) => {
    setServerError(null)
    startTransition(async () => {
      const result = await createEventAction(values)
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
        <Label htmlFor="kind">Type</Label>
        <select
          id="kind"
          className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
          {...register("kind")}
        >
          <option value="established">Established (multi-day)</option>
          <option value="flash">Flash (single day)</option>
        </select>
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
        {kind === "flash" && (
          <p className="text-xs text-muted-foreground">
            A flash event runs for a single day.
          </p>
        )}
      </div>

      {serverError && <p className="text-sm text-destructive">{serverError}</p>}

      <Button type="submit" disabled={isPending} className="mt-2 self-start">
        {isPending ? "Creating…" : "Create event"}
      </Button>
    </form>
  )
}
