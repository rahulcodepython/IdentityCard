import type { UseFormReturn } from "react-hook-form"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { SendOtpInput } from "@/schema/auth.types"
import { ErrorAlert, SubmitButton } from "./auth-ui-elements"

interface LoginFormProps {
    form: UseFormReturn<SendOtpInput>
    pending: boolean
    errorMessage: string | null
    onSubmit: () => void
}

export function LoginForm({ form, pending, errorMessage, onSubmit }: LoginFormProps) {
    const {
        register,
        formState: { errors },
    } = form

    return (
        <form onSubmit={onSubmit} noValidate>
            <FieldGroup>
                <ErrorAlert message={errorMessage} />

                <Field data-invalid={!!errors.email}>
                    <FieldLabel htmlFor="login-email">Email</FieldLabel>
                    <Input
                        id="login-email"
                        type="email"
                        autoComplete="email"
                        autoFocus
                        placeholder="you@example.com"
                        aria-invalid={!!errors.email}
                        {...register("email")}
                    />
                    <FieldError errors={[errors.email]} />
                </Field>

                <Field>
                    <SubmitButton
                        pending={pending}
                        text="Proceed with this email"
                        loadingText="Sending verification code…"
                    />
                </Field>
            </FieldGroup>
        </form>
    )
}