"use client"

import { useState, useTransition } from "react"

import { Button } from "@/components/ui/button"

import { resendCardAction } from "./actions"

export function CardActions({
  eventId,
  personId,
  downloadHref,
  eventPublished,
  cardSentAt,
}: {
  eventId: string
  personId: string
  downloadHref: string
  eventPublished: boolean
  cardSentAt: string | null
}) {
  const [error, setError] = useState<string | null>(null)
  const [sent, setSent] = useState(false)
  const [isPending, startTransition] = useTransition()

  return (
    <div className="flex flex-col gap-1 rounded-xl border p-3">
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          render={<a href={downloadHref} target="_blank" rel="noreferrer" />}
        >
          Download card
        </Button>
        <Button
          variant="outline"
          size="sm"
          disabled={isPending || !eventPublished}
          onClick={() =>
            startTransition(async () => {
              setError(null)
              const result = await resendCardAction(eventId, personId)
              if (result?.error) setError(result.error)
              else setSent(true)
            })
          }
        >
          {isPending ? "Sending…" : "Resend by email"}
        </Button>
      </div>
      <p className="text-xs text-muted-foreground">
        {!eventPublished
          ? "Emailing is available once the event is published — download works now for a preview."
          : cardSentAt
            ? `Last emailed ${new Date(cardSentAt).toLocaleString()}`
            : "Not yet emailed."}
        {sent && " · Resent just now."}
      </p>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
