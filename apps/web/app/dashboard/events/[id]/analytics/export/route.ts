import { headers } from "next/headers"
import { type NextRequest } from "next/server"

import { auth } from "@/lib/auth"

// Same pattern as the people CSV export: a browser-navigated download
// link can't attach a bearer token itself, so this proxies the request
// server-side (minting a fresh JWT from the current session — see
// lib/api/client.ts's apiFetch) and streams the Go API's CSV response
// through.
const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const { token } = await auth.api.getToken({ headers: await headers() })

  const res = await fetch(
    `${API_BASE_URL}/events/${id}/attendance/export${request.nextUrl.search}`,
    {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    }
  )

  if (!res.ok || !res.body) {
    return new Response("Failed to export attendance", {
      status: res.status || 500,
    })
  }

  return new Response(res.body, {
    status: 200,
    headers: {
      "Content-Type": "text/csv; charset=utf-8",
      "Content-Disposition":
        res.headers.get("Content-Disposition") ??
        'attachment; filename="attendance.csv"',
    },
  })
}
