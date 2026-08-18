"use client"

import { useState, useTransition } from "react"
import { useForm } from "react-hook-form"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

import { submitFormAction } from "./actions"

type FormValues = {
  email: string
  mobile: string
  name: string
  image_url: string
  age: string
  gender: string
}

export function SubmitForm({ token }: { token: string }) {
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
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
    },
  })

  const onSubmit = handleSubmit((values) => {
    setError(null)
    startTransition(async () => {
      const result = await submitFormAction(token, {
        email: values.email,
        mobile: values.mobile,
        name: values.name,
        image_url: values.image_url,
        age: values.age.trim() === "" ? undefined : Number(values.age),
        gender: values.gender,
      })
      if ("error" in result) {
        setError(result.error)
      } else {
        setSuccess(true)
      }
    })
  })

  if (success) {
    return (
      <p className="text-sm">You&apos;re registered — thanks for signing up!</p>
    )
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          aria-invalid={!!errors.name}
          {...register("name", { required: "Enter your name" })}
        />
        {errors.name && (
          <p className="text-xs text-destructive">{errors.name.message}</p>
        )}
      </div>

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          type="email"
          aria-invalid={!!errors.email}
          {...register("email", { required: "Enter your email" })}
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
          {...register("mobile", { required: "Enter your mobile number" })}
        />
        {errors.mobile && (
          <p className="text-xs text-destructive">{errors.mobile.message}</p>
        )}
      </div>

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="image_url">Photo URL (optional)</Label>
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

      {error && <p className="text-sm text-destructive">{error}</p>}

      <Button type="submit" disabled={isPending} className="mt-2">
        {isPending ? "Submitting…" : "Register"}
      </Button>
    </form>
  )
}
