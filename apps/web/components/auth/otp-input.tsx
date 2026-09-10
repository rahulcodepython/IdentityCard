"use client"

import { useEffect, useRef } from "react"
import {
    InputOTP,
    InputOTPGroup,
    InputOTPSlot,
} from "@/components/ui/input-otp"

const DEFAULT_SLOTS = [0, 1, 2, 3, 4, 5]

interface OtpInputProps {
    id?: string
    value: string
    onChange: (value: string) => void
    onComplete: (code: string) => void | Promise<void>
    disabled?: boolean
    autoFocus?: boolean
    maxLength?: number
    ariaInvalid?: boolean
}

export function OtpInput({
    id,
    value,
    onChange,
    onComplete,
    disabled = false,
    autoFocus = true,
    maxLength = 6,
    ariaInvalid = false,
}: OtpInputProps) {
    const lastSubmittedCode = useRef<string | null>(null)
    const isSubmitting = useRef<boolean>(false)

    useEffect(() => {
        if (value.length < maxLength) {
            lastSubmittedCode.current = null
            isSubmitting.current = false
            return
        }

        if (
            value.length === maxLength &&
            !disabled &&
            !isSubmitting.current &&
            lastSubmittedCode.current !== value
        ) {
            isSubmitting.current = true
            lastSubmittedCode.current = value
            try {
                const result = onComplete(value)
                if (result instanceof Promise) {
                    result.finally(() => {
                        isSubmitting.current = false
                    })
                } else {
                    isSubmitting.current = false
                }
            } catch {
                isSubmitting.current = false
            }
        }
    }, [value, maxLength, disabled, onComplete])

    return (
        <div className="flex flex-col items-center justify-center py-2">
            <InputOTP
                id={id}
                maxLength={maxLength}
                value={value}
                autoFocus={autoFocus}
                disabled={disabled}
                aria-invalid={ariaInvalid}
                onChange={(val) => {
                    onChange(val)
                }}
            >
                <InputOTPGroup>
                    {DEFAULT_SLOTS.slice(0, maxLength).map((i) => (
                        <InputOTPSlot key={i} index={i} />
                    ))}
                </InputOTPGroup>
            </InputOTP>
        </div>
    )
}
