"use server"

import { register } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import { type RegisterInput, registerSchema } from "@/lib/validation/auth"

export type RegisterActionResult =
    | { error: string }
    | { email: string; totpQrImage: string; totpSecret: string }

export async function registerAction(
    input: RegisterInput
): Promise<RegisterActionResult> {
    const parsed = registerSchema.safeParse(input)
    if (!parsed.success) {
        return { error: "Please check the highlighted fields." }
    }

    try {
        const resp = await register(parsed.data)
        return {
            email: resp.email,
            totpQrImage: resp.totp_qr_image,
            totpSecret: resp.totp_secret,
        }
    } catch (err) {
        if (err instanceof ApiError) {
            return { error: err.message }
        }
        return { error: "Something went wrong. Please try again." }
    }
}
