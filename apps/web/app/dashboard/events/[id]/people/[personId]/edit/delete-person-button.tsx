"use client"

import { useState, useTransition } from "react"

import { Button } from "@/components/ui/button"

export function DeletePersonButton({
  action,
}: {
  action: () => Promise<{ error: string } | undefined>
}) {
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  return (
    <div className="flex flex-col items-end gap-1">
      <Button
        variant="destructive"
        disabled={isPending}
        onClick={() =>
          startTransition(async () => {
            const result = await action()
            if (result?.error) setError(result.error)
          })
        }
      >
        {isPending ? "Removing…" : "Remove"}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
