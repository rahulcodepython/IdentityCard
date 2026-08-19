"use client"

import { notFound, useParams } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { ApiError } from "@/react-query/client"
import { getPublicForm } from "@/lib/client-api/forms"
import { queryKeys } from "@/react-query/query-keys"

import { SubmitForm } from "./submit-form"

export default function PublicFormPage() {
  const { token } = useParams<{ token: string }>()

  const { data: form, error } = useQuery({
    queryKey: queryKeys.publicForm(token),
    queryFn: () => getPublicForm(token),
    retry: false,
  })

  if (error instanceof ApiError && error.status === 404) notFound()

  if (!form) {
    return (
      <div className="flex min-h-svh items-center justify-center p-6">
        <Card className="w-full max-w-sm">
          <CardContent className="flex justify-center py-8 text-sm text-muted-foreground">
            Loading…
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>
            {form.event_name}
            {form.sub_event_name ? ` — ${form.sub_event_name}` : ""}
          </CardTitle>
          <CardDescription>
            {form.is_open
              ? "Register your details below."
              : "This sign-up is currently closed."}
          </CardDescription>
        </CardHeader>
        {form.is_open && (
          <CardContent>
            <SubmitForm token={token} />
          </CardContent>
        )}
      </Card>
    </div>
  )
}