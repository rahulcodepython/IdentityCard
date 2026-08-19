"use client"

import { useState, useTransition } from "react"

import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useRouter } from "next/navigation"

import { Button } from "@/components/ui/button"
import { deleteEvent } from "@/lib/client-api/events"
import { queryKeys } from "@/react-query/query-keys"

export function DeleteButton({ eventId }: { eventId: string }) {
  const [error, setError] = useState<string | null>(null)
  const queryClient = useQueryClient()
  const router = useRouter()

  const deleteMutation = useMutation({
    mutationFn: () => deleteEvent(eventId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.events() })
      router.push("/dashboard/events")
    },
    onError: (err: any) => {
      setError(err.message || "Failed to delete event")
    }
  })

  return (
    <div className="flex flex-col items-end gap-1">
      <Button
        variant="destructive"
        disabled={deleteMutation.isPending}
        onClick={() => deleteMutation.mutate()}
      >
        {deleteMutation.isPending ? "Deleting…" : "Delete"}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
