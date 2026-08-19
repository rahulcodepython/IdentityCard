"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { getMe } from "@/lib/client-api/auth"
import { queryKeys } from "@/react-query/query-keys"

import { OnboardingForm } from "./onboarding-form"

// Only reachable by an authenticated user with no organization yet — a
// Google signup, since email registration already collects the
// organization name up front. Unauthenticated visitors bounce to /login;
// a user who already has an org (including one who just finished this
// step) bounces to /dashboard — this page can't be revisited.
export default function OnboardingPage() {
  const router = useRouter()

  const { data: user, isFetched, error } = useQuery({
    queryKey: queryKeys.me(),
    queryFn: getMe,
    retry: false,
  })

  useEffect(() => {
    if (!isFetched) return
    if (error) {
      router.replace("/login")
      return
    }
    if (user?.has_organization) {
      router.replace("/dashboard")
    }
  }, [user, isFetched, error, router])

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