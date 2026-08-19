import { apiRequest } from "@/react-query/client"
import { meResponseSchema } from "@/schema/auth.types"

export async function getMe() {
  return apiRequest({ url: "/auth/me", method: "GET" }, meResponseSchema)
}