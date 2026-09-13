import type {
    PrepareTotpParams,
    TotpPrepareResponse,
    TotpVerifyParams,
    TotpVerifyResponse,
    OrganizationSetupParams,
} from "./auth.types"

export async function prepareTotpRegistration(params: PrepareTotpParams): Promise<TotpPrepareResponse> {
    const res = await fetch("/api/auth/totp/prepare", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(params),
    })
    return res.json() as Promise<TotpPrepareResponse>
}

export async function completeTotpRegistration(params: TotpVerifyParams): Promise<TotpVerifyResponse> {
    const res = await fetch("/api/auth/totp/complete", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(params),
    })
    return res.json() as Promise<TotpVerifyResponse>
}

export async function signInWithTotp(params: TotpVerifyParams): Promise<TotpVerifyResponse> {
    const res = await fetch("/api/auth/totp/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(params),
    })
    return res.json() as Promise<TotpVerifyResponse>
}

export async function organizationSetup(params: OrganizationSetupParams): Promise<void> {
    await fetch("/api/organization/setup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(params),
    })
}