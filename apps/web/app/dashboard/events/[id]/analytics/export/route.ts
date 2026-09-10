import { headers } from "next/headers"
import { type NextRequest } from "next/server"

import { auth } from "@/lib/auth"
import { API_V1_PREFIX } from "@/lib/constants"

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8000"

export async function GET(
    request: NextRequest,
    { params }: { params: Promise<{ id: string }> }
) {
    const { id } = await params
    const { token } = await auth.api.getToken({ headers: await headers() })

    const res = await fetch(
        `${API_BASE_URL}${API_V1_PREFIX}/events/${id}/attendance/export${request.nextUrl.search}`,
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
