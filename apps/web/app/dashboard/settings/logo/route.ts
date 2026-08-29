import { headers } from "next/headers"

import { auth } from "@/lib/auth"

// Proxies the org logo image the same way the people CSV export does: a
// browser <img> tag can't attach a bearer token itself, so this mints one
// server-side and streams the Go API's response straight through.
const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

export async function GET() {
  const { token } = await auth.api.getToken({ headers: await headers() })

  const res = await fetch(`${API_BASE_URL}/organizations/logo`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  })

  if (!res.ok || !res.body) {
    return new Response("Not found", { status: res.status || 404 })
  }

  return new Response(res.body, {
    status: 200,
    headers: { "Content-Type": res.headers.get("Content-Type") ?? "image/png" },
  })
}
