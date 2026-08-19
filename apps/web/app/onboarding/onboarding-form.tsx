"use client"

import { useState, useTransition } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { useRouter } from "next/navigation"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  type CreateOrganizationInput,
  createOrganizationSchema,
} from "@/lib/validation/organizations"

import { createOrganization } from "@/lib/client-api/organizations"

export function OnboardingForm() {
  const router = useRouter()
  const [serverError, setServerError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreateOrganizationInput>({
    resolver: zodResolver(createOrganizationSchema),
  })

  const onSubmit = handleSubmit((values) => {
    setServerError(null)
    startTransition(async () => {
      try {
        await createOrganization(values)
      } catch (err: any) {
        setServerError(err.message ?? "Something went wrong.")
        return
      }
      router.push("/dashboard")
    })
  })

  return (
    <form onSubmit={onSubmit} className="flex w-full flex-col gap-4" noValidate>
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="organization_name">Organization name</Label>
        <Input
          id="organization_name"
          autoComplete="organization"
          aria-invalid={!!errors.organization_name}
          {...register("organization_name")}
        />
        {errors.organization_name && (
          <p className="text-xs text-destructive">
            {errors.organization_name.message}
          </p>
        )}
      </div>

      {serverError && <p className="text-sm text-destructive">{serverError}</p>}

      <Button type="submit" disabled={isPending} className="mt-2">
        {isPending ? "Setting up…" : "Continue"}
      </Button>
    </form>
  )
}