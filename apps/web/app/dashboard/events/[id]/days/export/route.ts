import { headers } from "next/headers"
import { type NextRequest } from "next/server"

import { auth } from "@/lib/auth"

// Mirrors app/dashboard/events/[id]/people/export/route.ts — a CSV
// download can't go through the JSON-envelope apiFetch client, and (being
// a browser-navigated link, not a fetch call) can't attach a bearer token
// itself either, so this mints one server-side. Always available
// regardless of schedule_mode or draft/published status (see
// events.Service.ExportDays).
const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const { token } = await auth.api.getToken({ headers: await headers() })

  const res = await fetch(`${API_BASE_URL}/events/${id}/days/export`, {
    headers: { Authorization: `Bearer ${token}` },
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
