"use client"

import { useState, useTransition } from "react"
import { useForm } from "react-hook-form"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

import { createDeviceAction } from "./actions"

type FormValues = { name: string }
type OtpReveal = { deviceName: string; code: string; expiresAt: string }

export function CreateDeviceForm() {
  const [error, setError] = useState<string | null>(null)
  const [otp, setOtp] = useState<OtpReveal | null>(null)
  const [isPending, startTransition] = useTransition()
  const { register, handleSubmit, reset } = useForm<FormValues>({
    defaultValues: { name: "" },
  })

  const onSubmit = handleSubmit((values) => {
    setError(null)
    setOtp(null)
    startTransition(async () => {
      const result = await createDeviceAction(values)
      if ("error" in result) {
        setError(result.error)
        return
      }
      setOtp({
        deviceName: result.deviceName,
        code: result.otpCode,
        expiresAt: result.otpExpiresAt,
      })
      reset()
    })
  })

  return (
    <div className="flex flex-col gap-3">
      <form onSubmit={onSubmit} className="flex items-end gap-2">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="name">Device name</Label>
          <Input
            id="name"
            placeholder="Entrance Tablet"
            {...register("name", { required: true })}
          />
        </div>
        <Button type="submit" disabled={isPending}>
          {isPending ? "Adding…" : "Add device"}
        </Button>
      </form>

      {error && <p className="text-sm text-destructive">{error}</p>}

      {otp && (
        <div className="rounded-xl border bg-muted/50 p-4 text-sm">
          <p className="font-medium">
            Pairing code for {otp.deviceName}:{" "}
            <span className="text-2xl tracking-widest">{otp.code}</span>
          </p>
          <p className="text-muted-foreground">
            On the scanner device, go to <code>/pair</code> and enter this code
            before it expires at {new Date(otp.expiresAt).toLocaleTimeString()}.
          </p>
        </div>
      )}
    </div>
  )
}
