import Link from "next/link"

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

import { LoginForm } from "./login-form"

export default function LoginPage() {
  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Sign in</CardTitle>
          <CardDescription>
            Sign in to manage your organization&apos;s events.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <LoginForm />
        </CardContent>
        <CardFooter className="justify-center border-t text-sm text-muted-foreground">
          Don&apos;t have an organization yet?{" "}
          <Link href="/plans" className="ml-1 underline underline-offset-4">
            View plans
          </Link>
        </CardFooter>
      </Card>
    </div>
  )
}
