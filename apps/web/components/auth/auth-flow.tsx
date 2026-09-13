"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { toast } from "sonner"
import { AuthCard } from "@/components/auth/auth-card"
import { authClient } from "@/lib/auth-client"
import { useSessionStore } from "@/store/session.store"
import {
    type SendOtpInput,
    type RegisterInput,
    sendOtpSchema,
    registerSchema,
} from "@/schema/auth.types"
import type { AuthFlowProps, AuthStep, AuthTab } from "@/components/auth/auth.types"
import {
    prepareTotpRegistration,
    completeTotpRegistration,
    signInWithTotp,
    organizationSetup,
} from "@/components/auth/auth.api"
import { VerificationPanel } from "@/components/auth/verification-panel"
import { LoginForm } from "@/components/auth/login-form"
import { RegisterForm } from "@/components/auth/register-form"
import {
    MSG_ACCOUNT_CREATED,
    MSG_SIGNED_IN,
    ROUTE_DASHBOARD,
    ROUTE_LOGIN,
    ROUTE_REGISTER,
} from "@/lib/constants"

export function AuthFlow({ mode, onSuccess }: AuthFlowProps) {
    const router = useRouter()
    const [step, setStep] = useState<AuthStep>("details")
    const [activeTab, setActiveTab] = useState<AuthTab>("email")

    const [email, setEmail] = useState("")
    const [name, setName] = useState("")
    const [organizationName, setOrganizationName] = useState("")

    const [code, setCode] = useState("")
    const [pending, setPending] = useState(false)
    const [errorMessage, setErrorMessage] = useState<string | null>(null)
    const [resendCooldown, setResendCooldown] = useState(0)

    const [totpUri, setTotpUri] = useState<string | null>(null)
    const [totpSecret, setTotpSecret] = useState<string | null>(null)
    const [totpLoading, setTotpLoading] = useState(false)

    const loginForm = useForm<SendOtpInput>({
        resolver: zodResolver(sendOtpSchema),
        defaultValues: { email: "" },
    })

    const registerForm = useForm<RegisterInput>({
        resolver: zodResolver(registerSchema),
        defaultValues: { name: "", organizationName: "", email: "" },
    })

    useEffect(() => {
        if (resendCooldown <= 0) return
        const timer = setInterval(() => {
            setResendCooldown((prev) => prev - 1)
        }, 1000)
        return () => clearInterval(timer)
    }, [resendCooldown])

    const handleInitialSubmit = async (targetEmail: string, targetName?: string, targetOrg?: string) => {
        setPending(true)
        setErrorMessage(null)

        const { error } = await authClient.emailOtp.sendVerificationOtp({
            email: targetEmail,
            type: "sign-in",
        })

        if (error) {
            setPending(false)
            const msg = error.message ?? "Failed to send verification code."
            setErrorMessage(msg)
            toast.error(msg)
            return
        }

        setEmail(targetEmail)
        if (targetName) setName(targetName)
        if (targetOrg) setOrganizationName(targetOrg)

        if (mode === "register" && targetName && targetOrg) {
            setTotpLoading(true)
            void prepareTotpRegistration({
                email: targetEmail,
                name: targetName,
                organizationName: targetOrg,
            })
                .then((totpRes) => {
                    setTotpLoading(false)
                    if (totpRes.success && totpRes.totpURI && totpRes.secret) {
                        setTotpUri(totpRes.totpURI)
                        setTotpSecret(totpRes.secret)
                    }
                })
                .catch(() => {
                    setTotpLoading(false)
                })
        }

        setPending(false)
        setStep("verify")
        setCode("")
        setResendCooldown(30)
        toast.success(`Verification code sent to ${targetEmail}`)
    }

    const handleResendEmailOtp = async () => {
        if (resendCooldown > 0 || pending) return
        setPending(true)
        setErrorMessage(null)

        const { error } = await authClient.emailOtp.sendVerificationOtp({
            email,
            type: "sign-in",
        })
        setPending(false)

        if (error) {
            const msg = error.message ?? "Failed to resend verification code."
            setErrorMessage(msg)
            toast.error(msg)
            return
        }

        setResendCooldown(30)
        toast.success(`New verification code sent to ${email}`)
    }

    const handleChangeEmail = () => {
        setStep("details")
        setCode("")
        setErrorMessage(null)
    }

    const handleTabChange = (tab: AuthTab) => {
        setActiveTab(tab)
        setCode("")
        setErrorMessage(null)
    }

    const handleVerify = async (verifyCode: string) => {
        if (pending) return
        setPending(true)
        setErrorMessage(null)

        try {
            if (activeTab === "email") {
                const { data, error } = await authClient.signIn.emailOtp({
                    email,
                    otp: verifyCode,
                    name: name || undefined,
                })

                if (error || !data) {
                    setPending(false)
                    const msg = error?.message ?? "Invalid or expired verification code."
                    setErrorMessage(msg)
                    toast.error(msg)
                    setCode("")
                    return
                }

                if (mode === "register" && organizationName) {
                    await organizationSetup({ organizationName, userEmail: email })
                }
            } else {
                const totpResult =
                    mode === "register"
                        ? await completeTotpRegistration({ email, code: verifyCode })
                        : await signInWithTotp({ email, code: verifyCode })

                if (!totpResult.success) {
                    setPending(false)
                    const msg = totpResult.error ?? "Invalid authenticator code."
                    setErrorMessage(msg)
                    toast.error(msg)
                    setCode("")
                    return
                }
            }

            try {
                const { data: tokenData } = await authClient.token()
                const { data: sessionData } = await authClient.getSession()
                if (sessionData?.user && tokenData?.token) {
                    useSessionStore.getState().setSession({
                        token: tokenData.token,
                        user: {
                            id: sessionData.user.id,
                            name: sessionData.user.name,
                            email: sessionData.user.email,
                            image: sessionData.user.image,
                        }
                    })
                }
            } catch (sessErr) {
                console.warn("Failed to cache session after sign in:", sessErr)
            }

            setPending(false)
            toast.success(mode === "register" ? MSG_ACCOUNT_CREATED : MSG_SIGNED_IN)

            if (onSuccess) {
                onSuccess()
            } else {
                router.push(ROUTE_DASHBOARD)
            }
        } catch (err) {
            setPending(false)
            const msg = err instanceof Error ? err.message : "Verification failed. Please try again."
            setErrorMessage(msg)
            toast.error(msg)
            setCode("")
        }
    }

    const cardTitle =
        step === "details"
            ? mode === "login"
                ? "Welcome back"
                : "Create your account"
            : "Verify your account"

    const cardSubtitle =
        step === "details"
            ? mode === "login"
                ? "Sign in to manage your organization's events."
                : "Sign up to manage your organization's events."
            : `Enter the code to verify ${email}`

    return (
        <AuthCard
            title={cardTitle}
            subtitle={cardSubtitle}
            footerText={mode === "login" ? "Don't have an account?" : "Already have an account?"}
            footerLinkText={mode === "login" ? "Sign up" : "Sign in"}
            footerLinkHref={mode === "login" ? ROUTE_REGISTER : ROUTE_LOGIN}
            showSocial={step === "details"}
        >
            {
                step === "details" ?
                    mode === "login" ? <LoginForm
                        form={loginForm}
                        pending={pending}
                        errorMessage={errorMessage}
                        onSubmit={loginForm.handleSubmit((values) => handleInitialSubmit(values.email))}
                    /> : <RegisterForm
                        form={registerForm}
                        pending={pending}
                        errorMessage={errorMessage}
                        onSubmit={registerForm.handleSubmit((values) =>
                            handleInitialSubmit(values.email, values.name, values.organizationName)
                        )}
                    />
                    : <VerificationPanel
                        mode={mode}
                        email={email}
                        activeTab={activeTab}
                        code={code}
                        pending={pending}
                        errorMessage={errorMessage}
                        resendCooldown={resendCooldown}
                        totpLoading={totpLoading}
                        totpUri={totpUri}
                        totpSecret={totpSecret}
                        onTabChange={handleTabChange}
                        onCodeChange={setCode}
                        onVerify={handleVerify}
                        onResendEmailOtp={() => void handleResendEmailOtp()}
                        onChangeEmail={handleChangeEmail}
                    />
            }
        </AuthCard>
    )
}