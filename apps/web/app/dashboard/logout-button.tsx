"use client"

import { useTransition } from "react"

import { Button } from "@/components/ui/button"

import { logoutAction } from "./actions"

export function LogoutButton() {
  const [isPending, startTransition] = useTransition()

  return (
    <Button
      variant="outline"
      disabled={isPending}
      onClick={() => startTransition(() => logoutAction())}
    >
      {isPending ? "Signing out…" : "Sign out"}
    </Button>
  )
}
