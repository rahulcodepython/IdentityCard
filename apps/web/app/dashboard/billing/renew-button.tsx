"use client"

import { useState, useTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"
import { renewSubscription } from "@/lib/client-api/plans"
import { queryKeys } from "@/react-query/query-keys"

export function RenewButton({ subscriptionId }: { subscriptionId: string }) {
  const queryClient = useQueryClient()
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
            try {
              await renewSubscription(subscriptionId)
              queryClient.invalidateQueries({ queryKey: queryKeys.orgSubscriptions() })
              queryClient.invalidateQueries({ queryKey: queryKeys.plans() })
            } catch (err: any) {
              setError(err.message ?? "Something went wrong.")
            }
          })
        }}
      >
        {isPending ? "Renewing…" : "Renew now"}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}