"use client"

import { AuthCard } from "@/components/auth/auth-card"
import { EmailAuthFlow } from "@/components/auth/email-auth-flow"
import { TotpAuthFlow } from "@/components/auth/totp-auth-flow"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

export default function RegisterPage() {
    const handleSuccess = () => {
        window.location.assign("/dashboard")
    }

    return (
        <AuthCard
            title="Create your account"
            subtitle="Sign up to manage your organization's events."
            footerText="Already have an account?"
            footerLinkText="Sign in"
            footerLinkHref="/login"
        >
            <Tabs defaultValue="email" className="w-full">
                <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="email">Email OTP</TabsTrigger>
                    <TabsTrigger value="totp">Authenticator (TOTP)</TabsTrigger>
                </TabsList>

                <TabsContent value="email" className="mt-4">
                    <EmailAuthFlow mode="register" onSuccess={handleSuccess} />
                </TabsContent>

                <TabsContent value="totp" className="mt-4">
                    <TotpAuthFlow mode="register" onSuccess={handleSuccess} />
                </TabsContent>
            </Tabs>
        </AuthCard>
    )
}
