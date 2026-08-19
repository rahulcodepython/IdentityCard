"use client"

import { useEffect, useRef, useState, useTransition } from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { RiLoader4Line } from "@remixicon/react"

import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp"
import { Label } from "@/components/ui/label"
import { DeviceApiError, pairDevice, setDeviceKey } from "@/lib/device-client"

const SLOTS = [0, 1, 2, 3, 4, 5]

export default function PairPage() {
    const router = useRouter()
    const [otp, setOtp] = useState("")
    const [isPending, startTransition] = useTransition()
    const lastAttemptedRef = useRef("")

    // Auto-submit API call when 6th digit is entered
    useEffect(() => {
        if (otp.length === 6) {
            if (otp !== lastAttemptedRef.current) {
                lastAttemptedRef.current = otp
                startTransition(async () => {
                    try {
                        const result = await pairDevice(otp)
                        setDeviceKey(result.key)
                        toast.success("Device paired successfully!")
                        router.push("/scanner")
                    } catch (err) {
                        const errorMsg =
                            err instanceof DeviceApiError
                                ? err.message
                                : "Invalid code or pairing failed. Please try again."
                        toast.error(errorMsg)
                    }
                })
            }
        } else {
            // Reset last attempted code so backspacing & re-typing 6th digit fires API call again
            lastAttemptedRef.current = ""
        }
    }, [otp, router])

    return (
        <div className="flex min-h-svh items-center justify-center p-6 bg-background">
            <Card className="w-full max-w-sm shadow-md">
                <CardHeader className="text-center">
                    <CardTitle className="text-xl font-bold">Pair scanner device</CardTitle>
                    <CardDescription>
                        Enter the 6-digit code shown to the admin who created this scanner device.
                    </CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col items-center gap-4 py-2">
                    <div className="flex flex-col items-center gap-2 w-full">
                        <Label htmlFor="otp" className="sr-only">
                            Pairing Code
                        </Label>
                        <InputOTP
                            maxLength={6}
                            value={otp}
                            onChange={setOtp}
                            disabled={isPending}
                        >
                            <InputOTPGroup>
                                {SLOTS.map((i) => (
                                    <InputOTPSlot key={i} index={i} />
                                ))}
                            </InputOTPGroup>
                        </InputOTP>
                    </div>

                    {isPending && (
                        <div className="flex items-center gap-2 text-xs font-medium text-primary animate-pulse py-1">
                            <RiLoader4Line className="size-4 animate-spin" />
                            <span>Pairing device…</span>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}
