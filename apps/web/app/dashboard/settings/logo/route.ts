import { cookies } from "next/headers"

// Proxies the org logo image the same way the people CSV export does: a
// browser <img> tag can't attach the httpOnly auth cookie itself, so this
// forwards it server-side and streams the Go API's response straight through.
const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

export async function GET() {
  const cookieStore = await cookies()

  const res = await fetch(`${API_BASE_URL}/organizations/logo`, {
    headers: { Cookie: cookieStore.toString() },
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
