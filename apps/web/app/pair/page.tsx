"use client"

import { useState, useTransition } from "react"
import { useRouter } from "next/navigation"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { DeviceApiError, pairDevice, setDeviceKey } from "@/lib/device-client"

export default function PairPage() {
  const router = useRouter()
  const [otp, setOtp] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    startTransition(async () => {
      try {
        const result = await pairDevice(otp)
        setDeviceKey(result.key)
        router.push("/scanner")
      } catch (err) {
        setError(
          err instanceof DeviceApiError
            ? err.message
            : "Something went wrong. Please try again."
        )
      }
    })
  }

  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Pair this device</CardTitle>
          <CardDescription>
            Enter the 6-digit code shown to the admin who added this device.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="otp">Code</Label>
              <Input
                id="otp"
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={6}
                value={otp}
                onChange={(e) => setOtp(e.target.value.replace(/\D/g, ""))}
                className="text-center text-2xl tracking-[0.5em]"
                autoFocus
              />
            </div>
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" disabled={isPending || otp.length !== 6}>
              {isPending ? "Pairing…" : "Pair device"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
