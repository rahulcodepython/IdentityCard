"use client"

import { useState, useTransition } from "react"

import { Button } from "@/components/ui/button"
import type { Device } from "@/lib/validation/devices"

import { revokeDeviceAction } from "./actions"

export function DeviceRow({ device }: { device: Device }) {
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  return (
    <div className="flex items-center justify-between rounded-xl border p-3">
      <div>
        <p className="font-medium">{device.name}</p>
        <p className="text-xs text-muted-foreground">
          {device.status}
          {device.verified_at &&
            ` · verified ${new Date(device.verified_at).toLocaleString()}`}
        </p>
        {error && <p className="text-xs text-destructive">{error}</p>}
      </div>
      {device.status !== "revoked" && (
        <Button
          variant="destructive"
          size="sm"
          disabled={isPending}
          onClick={() =>
            startTransition(async () => {
              const result = await revokeDeviceAction(device.id)
              if (result?.error) setError(result.error)
            })
          }
        >
          Revoke
        </Button>
      )}
    </div>
  )
}
