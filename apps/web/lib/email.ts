
export const EmailOTPTemplate: Record<string, { subject: string; body: (otp: string) => string }> = {
    "sign-in": {
        subject: "Your IdentityCard sign-in code",
        body: (otp) => `Your sign-in code is ${otp}. It expires in 5 minutes.`,
    },
    "email-verification": {
        subject: "Verify your email",
        body: (otp) => `Your verification code is ${otp}. It expires in 5 minutes.`,
    },
    "forget-password": {
        subject: "Reset your password",
        body: (otp) => `Your password reset code is ${otp}. It expires in 5 minutes.`,
    },
    "change-email": {
        subject: "Confirm your new email",
        body: (otp) => `Your email change code is ${otp}. It expires in 5 minutes.`,
    },
}