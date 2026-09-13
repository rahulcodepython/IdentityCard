import type { UseFormReturn } from "react-hook-form"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { RegisterInput } from "@/schema/auth.types"
import { ErrorAlert, SubmitButton } from "./auth-ui-elements"

interface RegisterFormProps {
    form: UseFormReturn<RegisterInput>
    pending: boolean
    errorMessage: string | null
    onSubmit: () => void
}

export function RegisterForm({ form, pending, errorMessage, onSubmit }: RegisterFormProps) {
    const {
        register,
        formState: { errors },
    } = form

    return (
        <form onSubmit={onSubmit} noValidate>
            <FieldGroup>
                <ErrorAlert message={errorMessage} />

                <Field data-invalid={!!errors.name}>
                    <FieldLabel htmlFor="register-name">Your name</FieldLabel>
                    <Input
                        id="register-name"
                        autoComplete="name"
                        autoFocus
                        placeholder="e.g. Alex Smith"
                        aria-invalid={!!errors.name}
                        {...register("name")}
                    />
                    <FieldError errors={[errors.name]} />
                </Field>

                <Field data-invalid={!!errors.organizationName}>
                    <FieldLabel htmlFor="register-org">Organization name</FieldLabel>
                    <Input
                        id="register-org"
                        placeholder="e.g. Acme Corp"
                        aria-invalid={!!errors.organizationName}
                        {...register("organizationName")}
                    />
                    <FieldError errors={[errors.organizationName]} />
                </Field>

                <Field data-invalid={!!errors.email}>
                    <FieldLabel htmlFor="register-email">Email</FieldLabel>
                    <Input
                        id="register-email"
                        type="email"
                        autoComplete="email"
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