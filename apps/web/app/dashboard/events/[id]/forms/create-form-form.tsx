"use client"

import { useState, useTransition } from "react"
import { useForm } from "react-hook-form"
import { useQueryClient } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { SubEvent } from "@/lib/validation/subevents"

import { createForm } from "@/lib/client-api/forms"
import { queryKeys } from "@/react-query/query-keys"

type FormValues = { sub_event_id: string; capacity: string }

export function CreateFormForm({
  eventId,
  subEvents,
}: {
  eventId: string
  subEvents: SubEvent[]
}) {
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()
  const { register, handleSubmit, reset } = useForm<FormValues>({
    defaultValues: { sub_event_id: "", capacity: "" },
  })

  const onSubmit = handleSubmit((values) => {
    setError(null)
    startTransition(async () => {
      try {
        await createForm(eventId, {
          sub_event_id: values.sub_event_id || undefined,
          capacity:
            values.capacity.trim() === "" ? undefined : Number(values.capacity),
        })
      } catch (err: any) {
        setError(err.message ?? "Something went wrong.")
        return
      }
      queryClient.invalidateQueries({ queryKey: queryKeys.forms(eventId) })
      reset()
    })
  })

  return (
    <form
      onSubmit={onSubmit}
      className="flex flex-wrap items-end gap-3 rounded-lg border p-4"
    >
      {subEvents.length > 0 && (
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="sub_event_id">Scope</Label>
          <select
            id="sub_event_id"
            className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
            {...register("sub_event_id")}
          >
            <option value="">Whole event</option>
            {subEvents.map((subEvent) => (
              <option key={subEvent.id} value={subEvent.id}>
                {subEvent.name}
              </option>
            ))}
          </select>
        </div>
      )}
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="capacity">Capacity (optional)</Label>
        <Input
          id="capacity"
          type="number"
          min={1}
          className="w-32"
          {...register("capacity")}
        />
      </div>
      <Button type="submit" disabled={isPending}>
        {isPending ? "Creating…" : "Create link"}
      </Button>
      {error && <p className="w-full text-sm text-destructive">{error}</p>}
    </form>
  )
}