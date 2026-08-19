import { cookies } from "next/headers"
import { type NextRequest } from "next/server"

// Mirrors app/dashboard/events/[id]/people/export/route.ts — a CSV
// download can't go through the JSON-envelope apiFetch client, and (being
// a browser-navigated link, not a fetch call) can't attach the httpOnly
// auth cookie itself either. Always available regardless of schedule_mode
// or draft/published status (see events.Service.ExportDays).
const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const cookieStore = await cookies()

  const res = await fetch(`${API_BASE_URL}/events/${id}/days/export`, {
    headers: { Cookie: cookieStore.toString() },
    cache: "no-store",
  })

  if (!res.ok || !res.body) {
    return new Response("Failed to export days", { status: res.status || 500 })
  }

  return new Response(res.body, {
    status: 200,
    headers: {
      "Content-Type": "text/csv; charset=utf-8",
      "Content-Disposition":
        res.headers.get("Content-Disposition") ?? 'attachment; filename="event-days.csv"',
    },
  })
}
