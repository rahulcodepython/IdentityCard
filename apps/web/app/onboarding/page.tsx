import { redirect } from "next/navigation"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { me } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"

import { OnboardingForm } from "./onboarding-form"

// Only reachable by an authenticated user with no organization yet — a
// Google signup, since email registration already collects the
// organization name up front. Unauthenticated visitors bounce to /login;
// a user who already has an org (including one who just finished this
// step) bounces to /dashboard — this page can't be revisited.
export default async function OnboardingPage() {
  let user
  try {
    user = await me()
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/login")
    throw err
  }

  if (user.has_organization) {
    redirect("/dashboard")
  }

  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>What&apos;s your organization called?</CardTitle>
          <CardDescription>
            One last step before your dashboard is ready.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <OnboardingForm />
        </CardContent>
      </Card>
    </div>
  )
}
