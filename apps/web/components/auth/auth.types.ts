export interface AuthFlowProps {
    mode: "login" | "register"
    onSuccess?: () => void
}

export type AuthStep = "details" | "verify"
export type AuthTab = "email" | "totp"

export interface PrepareTotpParams {
    email: string
    name: string
    organizationName: string
}

export interface TotpVerifyParams {
    email: string
    code: string
}

export interface TotpPrepareResponse {
    success: boolean
    totpURI?: string
    secret?: string
    error?: string
}

export interface TotpVerifyResponse {
    success: boolean
    error?: string
}

export interface OrganizationSetupParams {
    organizationName: string
    userEmail: string
}