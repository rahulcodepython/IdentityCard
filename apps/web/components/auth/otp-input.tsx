"use client"

import {
    InputOTP,
    InputOTPGroup,
    InputOTPSlot,
} from "@/components/ui/input-otp"

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

export function OtpInput(props: OtpInputProps) {
    return (
        <div className="flex flex-col items-center justify-center py-2">
            <InputOTP
                id={props.id}
                maxLength={props.maxLength ?? 6}
                value={props.value}
                autoFocus={props.autoFocus}
                disabled={props.disabled}
                aria-invalid={props.ariaInvalid}
                onChange={(val) => {
                    if (val.length === props.maxLength) {
                        props.onComplete(val)
                    }
                    props.onChange(val)
                }}
            >
                <InputOTPGroup>
                    {[...Array(props.maxLength).keys()].map((i) => <InputOTPSlot key={i} index={i} />)}
                </InputOTPGroup>
            </InputOTP>
        </div>
    )
}
