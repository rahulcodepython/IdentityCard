import { notFound } from "next/navigation"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { ApiError } from "@/lib/api/client"
import { getPublicForm } from "@/lib/api/forms"

import { SubmitForm } from "./submit-form"

export default async function PublicFormPage({
  params,
}: {
  params: Promise<{ token: string }>
}) {
  const { token } = await params

  let form
  try {
    form = await getPublicForm(token)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
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
