"use client"

import { useState, useTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import type { EventForm } from "@/lib/validation/forms"

import { deleteForm, updateForm } from "@/lib/client-api/forms"
import { queryKeys } from "@/react-query/query-keys"

// NEXT_PUBLIC_ vars are inlined at build time identically on server and
// client, so this is safe to compute during SSR — unlike `window.location`,
// it can't cause a hydration mismatch.
const APP_URL = process.env.NEXT_PUBLIC_APP_URL ?? ""

export function FormRow({
  eventId,
  form,
  subEventName,
}: {
  eventId: string
  form: EventForm
  subEventName?: string
}) {
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [isPending, startTransition] = useTransition()

  const link = `${APP_URL}/forms/${form.token}`

  const copyLink = async () => {
    await navigator.clipboard.writeText(link)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: queryKeys.forms(eventId) })

  const toggleActive = () =>
    startTransition(async () => {
      setError(null)
      try {
        await updateForm(eventId, form.id, {
          capacity: form.capacity ?? undefined,
          is_active: !form.is_active,
        })
      } catch (err: any) {
        setError(err.message ?? "Something went wrong.")
        return
      }
      refresh()
    })

  const remove = () =>
    startTransition(async () => {
      setError(null)
      try {
        await deleteForm(eventId, form.id)
      } catch (err: any) {
        setError(err.message ?? "Something went wrong.")
        return
      }
      refresh()
    })

  return (
    <Card>
      <CardHeader>
        <CardTitle>{subEventName ?? "Whole event"}</CardTitle>
        <CardDescription>
          {form.submissions_count}
          {form.capacity ? ` / ${form.capacity}` : ""} submissions ·{" "}
          {form.is_active ? "Active" : "Inactive"}
        </CardDescription>
        <CardAction className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={copyLink}>
            {copied ? "Copied" : "Copy link"}
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={isPending}
            onClick={toggleActive}
          >
            {form.is_active ? "Deactivate" : "Activate"}
          </Button>
          <Button
            variant="destructive"
            size="sm"
            disabled={isPending}
            onClick={remove}
          >
            Delete
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-col gap-1">
        <Input
          readOnly
          value={link}
          onFocus={(e) => e.currentTarget.select()}
          className="text-xs"
        />
        {error && <p className="text-xs text-destructive">{error}</p>}
      </CardContent>
    </Card>
  )
}