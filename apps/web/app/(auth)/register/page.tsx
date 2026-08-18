import Link from "next/link"

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { listPlans } from "@/lib/api/plans"

import { RegisterForm } from "./register-form"

export default async function RegisterPage({
  searchParams,
}: {
  searchParams: Promise<{ plan?: string }>
}) {
  const { plan: planCode } = await searchParams
  const plans = await listPlans()
  const defaultPlanCode = plans.find((plan) => plan.code === planCode)?.code

  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Create your organization</CardTitle>
          <CardDescription>
            No payment is taken yet — plan selection is stubbed for now.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <RegisterForm plans={plans} defaultPlanCode={defaultPlanCode} />
        </CardContent>
        <CardFooter className="justify-center border-t text-sm text-muted-foreground">
          Already have an account?{" "}
          <Link href="/login" className="ml-1 underline underline-offset-4">
            Sign in
          </Link>
        </CardFooter>
      </Card>
    </div>
  )
}
