"use client"

import { useState, useTransition } from "react"

import { renewAction } from "@/app/dashboard/billing/actions"
import { Button } from "@/components/ui/button"

export function RenewButton({ subscriptionId }: { subscriptionId: string }) {
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  return (
    <div className="flex flex-col items-end gap-1">
      <Button
        size="sm"
        disabled={isPending}
        onClick={() => {
          setError(null)
          startTransition(async () => {
            const result = await renewAction(subscriptionId)
            if ("error" in result) setError(result.error)
          })
        }}
      >
        {isPending ? "Renewing…" : "Renew now"}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
