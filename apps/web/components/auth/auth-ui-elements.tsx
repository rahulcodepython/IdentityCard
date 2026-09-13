import { RiLoader4Line } from "@remixicon/react"
import { Button } from "@/components/ui/button"

export function ErrorAlert({ message }: { message: string | null }) {
    if (!message) return null
    return (
        <p className="w-full rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-center text-xs text-destructive">
            {message}
        </p>
    )
}

export function LoadingSpinner({ message }: { message: string }) {
    return (
        <div className="flex animate-pulse items-center gap-2 text-xs font-medium text-primary">
            <RiLoader4Line className="size-4 animate-spin" />
            <span>{message}</span>
        </div>
    )
}

interface SubmitButtonProps {
    pending: boolean
    text: string
    loadingText: string
}

export function SubmitButton({ pending, text, loadingText }: SubmitButtonProps) {
    return (
        <Button type="submit" className="w-full" disabled={pending}>
            {
                pending ? <div className="flex items-center gap-2">
                    <RiLoader4Line className="size-4 animate-spin" />
                    <span>{loadingText}</span>
                </div> : text
            }
        </Button>
    )
}

interface BackToEmailButtonProps {
    onClick: () => void
    disabled: boolean
}

export function BackToEmailButton({ onClick, disabled }: BackToEmailButtonProps) {
    return (
        <button
            type="button"
            onClick={onClick}
            disabled={disabled}
            className="text-xs text-muted-foreground transition-colors hover:text-foreground"
        >
            ← Change email
        </button>
    )
}