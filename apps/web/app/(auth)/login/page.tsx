"use client"

import { AuthCard } from "@/components/auth/auth-card"
import { EmailAuthFlow } from "@/components/auth/email-auth-flow"
import { TotpAuthFlow } from "@/components/auth/totp-auth-flow"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

export default function LoginPage() {
    const handleSuccess = () => {
        window.location.assign("/dashboard")
    }

    return (
        <AuthCard
            title="Welcome back"
            subtitle="Sign in to manage your organization's events."
            footerText="Don't have an account?"
            footerLinkText="Sign up"
            footerLinkHref="/register"
        >
            <Tabs defaultValue="email" className="w-full">
                <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="email">Email OTP</TabsTrigger>
                    <TabsTrigger value="totp">Authenticator (TOTP)</TabsTrigger>
                </TabsList>

                <TabsContent value="email" className="mt-4">
                    <EmailAuthFlow mode="login" onSuccess={handleSuccess} />
                </TabsContent>

                <TabsContent value="totp" className="mt-4">
                    <TotpAuthFlow mode="login" onSuccess={handleSuccess} />
                </TabsContent>
            </Tabs>
        </AuthCard>
    )
}
