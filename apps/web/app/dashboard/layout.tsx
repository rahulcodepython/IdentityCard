import Link from "next/link"
import { redirect } from "next/navigation"

import { me } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"

import { LogoutButton } from "./logout-button"

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  let user
  try {
    user = await me()
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login")
    }
    throw err
  }

  return (
    <div className="flex min-h-svh flex-col">
      <header className="flex items-center justify-between border-b p-4">
        <div>
          <Link href="/dashboard" className="font-heading text-lg font-medium">
            {user.organization_name}
          </Link>
          <p className="text-xs text-muted-foreground">
            {user.name} · {user.roles.join(", ")}
          </p>
        </div>
        <nav className="flex items-center gap-4">
          <Link
            href="/dashboard/events"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            Events
          </Link>
          {user.roles.includes("super_admin") && (
            <Link
              href="/dashboard/devices"
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Devices
            </Link>
          )}
          <Link
            href="/dashboard/settings"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            Settings
          </Link>
          <LogoutButton />
        </nav>
      </header>
      <main className="flex-1 p-6">{children}</main>
    </div>
  )
}
