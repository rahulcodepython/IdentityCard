import Link from "next/link"

export default function DashboardPage() {
  return (
    <div className="flex flex-col gap-4">
      <h1 className="font-heading text-xl font-medium">Dashboard</h1>
      <Link
        href="/dashboard/events"
        className="text-sm underline underline-offset-4"
      >
        Manage events
      </Link>
    </div>
  )
}
