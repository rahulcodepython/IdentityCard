"use client"

import { useState, useTransition } from "react"
import { useForm } from "react-hook-form"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { SubEvent } from "@/lib/validation/subevents"

type FormValues = {
  email: string
  mobile: string
  name: string
  image_url: string
  age: string
  gender: string
  sub_event_ids: string[]
}

// email/mobile are always present here (RHF initializes every field from
// defaultValues regardless of whether it's rendered) — updatePersonAction
// simply ignores them, since UpdatePersonInput has no such fields and zod
// silently drops unknown keys.
type PersonPayload = {
  email: string
  mobile: string
  name: string
  image_url: string
  age?: number
  gender: string
  sub_event_ids: string[]
}

// Shared by the create and edit person pages. Email/mobile key the
// upsert server-side (see people.Repository.Upsert) and can't be changed
// after creation — showIdentity controls whether they're editable here.
export function PersonForm({
  subEvents,
  defaultValues,
  showIdentity,
  action,
  submitLabel,
}: {
  subEvents: SubEvent[]
  defaultValues?: Partial<FormValues>
  showIdentity: boolean
  action: (values: PersonPayload) => Promise<{ error: string } | undefined>
  submitLabel: string
}) {
  const [serverError, setServerError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    defaultValues: {
      email: "",
      mobile: "",
      name: "",
      image_url: "",
      age: "",
      gender: "",
      sub_event_ids: [],
      ...defaultValues,
    },
  })

  const onSubmit = handleSubmit((values) => {
    setServerError(null)
    startTransition(async () => {
      const result = await action({
        email: values.email,
        mobile: values.mobile,
        name: values.name,
        image_url: values.image_url,
        age: values.age.trim() === "" ? undefined : Number(values.age),
        gender: values.gender,
        sub_event_ids: values.sub_event_ids,
      })
      if (result?.error) setServerError(result.error)
    })
  })

  return (
    <form
      onSubmit={onSubmit}
      className="flex max-w-lg flex-col gap-4"
      noValidate
    >
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          aria-invalid={!!errors.name}
          {...register("name", { required: "Enter a name" })}
        />
        {errors.name && (
          <p className="text-xs text-destructive">{errors.name.message}</p>
        )}
      </div>

      {showIdentity ? (
        <>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              aria-invalid={!!errors.email}
              {...register("email", { required: "Enter an email" })}
            />
            {errors.email && (
              <p className="text-xs text-destructive">{errors.email.message}</p>
            )}
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="mobile">Mobile</Label>
            <Input
              id="mobile"
              aria-invalid={!!errors.mobile}
              {...register("mobile", { required: "Enter a mobile number" })}
            />
            {errors.mobile && (
              <p className="text-xs text-destructive">
                {errors.mobile.message}
              </p>
            )}
          </div>
        </>
      ) : (
        defaultValues?.email && (
          <p className="text-sm text-muted-foreground">
            {defaultValues.email} · {defaultValues.mobile}
          </p>
        )
      )}

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="image_url">Photo URL</Label>
        <Input id="image_url" {...register("image_url")} />
      </div>

      <div className="flex gap-4">
        <div className="flex flex-1 flex-col gap-1.5">
          <Label htmlFor="age">Age</Label>
          <Input
            id="age"
            type="number"
            min={0}
            max={150}
            {...register("age")}
          />
        </div>
        <div className="flex flex-1 flex-col gap-1.5">
          <Label htmlFor="gender">Gender</Label>
          <Input id="gender" {...register("gender")} />
        </div>
      </div>

      {subEvents.length > 0 && (
        <div className="flex flex-col gap-2">
          <Label>Sub-events</Label>
          {subEvents.map((subEvent) => (
            <label
              key={subEvent.id}
              className="flex items-center gap-2 text-sm"
            >
              <input
                type="checkbox"
                value={subEvent.id}
                className="size-4"
                {...register("sub_event_ids")}
              />
              {subEvent.name}
            </label>
          ))}
        </div>
      )}

      {serverError && <p className="text-sm text-destructive">{serverError}</p>}

      <Button type="submit" disabled={isPending} className="mt-2 self-start">
        {isPending ? "Saving…" : submitLabel}
      </Button>
    </form>
  )
}
