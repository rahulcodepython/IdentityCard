"use client"

import { useActionState } from "react"

import { Button } from "@/components/ui/button"

import { type LogoActionResult, uploadLogoAction } from "./actions"

export function LogoUploadForm() {
  const [state, formAction, isPending] = useActionState<
    LogoActionResult,
    FormData
  >(uploadLogoAction, undefined)

  return (
    <form action={formAction} className="flex flex-col gap-3">
      <input
        type="file"
        name="logo"
        accept="image/png,image/jpeg"
        required
        className="text-sm"
      />
      <Button type="submit" disabled={isPending} className="self-start">
        {isPending ? "Uploading…" : "Upload logo"}
      </Button>
      {state?.error && (
        <p className="text-sm text-destructive">{state.error}</p>
      )}
    </form>
  )
}
