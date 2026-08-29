"use client"

import { createAuthClient } from "better-auth/react"
import { adminClient, jwtClient, organizationClient, emailOTPClient, twoFactorClient } from "better-auth/client/plugins"

import { ac, roles } from "@/lib/auth-access-control"

export const authClient = createAuthClient({
    baseURL: process.env.NEXT_PUBLIC_BETTER_AUTH_URL,
    plugins: [
        adminClient(),
        jwtClient(),
        organizationClient({ ac, roles }),
        emailOTPClient(),
        twoFactorClient(),
    ],
})

export const { signIn, signOut, useSession } = authClient
