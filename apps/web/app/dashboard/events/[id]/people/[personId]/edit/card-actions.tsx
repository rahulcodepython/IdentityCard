"use client"

import { useState, useTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"

import { resendCard } from "@/lib/client-api/cards"
import { queryKeys } from "@/react-query/query-keys"

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
  const queryClient = useQueryClient()
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
              try {
                await resendCard(eventId, personId)
              } catch (err: any) {
                setError(err.message ?? "Something went wrong.")
                return
              }
              setSent(true)
              queryClient.invalidateQueries({
                queryKey: queryKeys.person(eventId, personId),
              })
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