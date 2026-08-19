import Link from "next/link"

import { RiArrowRightLine } from "@remixicon/react"

import Features from "@/components/features"
import Pricing from "@/components/pricing"
import { Button } from "@/components/ui/button"
import { me } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"

// Public — visited by both authenticated and unauthenticated users. The
// single "Get started" action resolves to the right destination
// server-side: an authenticated visitor skips straight to the dashboard,
// everyone else goes to /login (which also handles sign-up).
async function getStartedHref() {
    try {
        await me()
        return "/dashboard"
    } catch (err) {
        if (err instanceof ApiError && err.status === 401) return "/login"
        throw err
    }
}

export default async function LandingPage() {
    const startHref = await getStartedHref()

    return (
        <div className="flex min-h-svh flex-col">
            <header className="sticky top-0 z-20 border-b bg-background/80 backdrop-blur-md">
                <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
                    <span className="font-heading text-lg font-semibold tracking-tight">IdentityCard</span>
                    <Button render={<Link href={startHref} />}>Get Started</Button>
                </div>
            </header>

            <main className="flex-1">
                <section className="relative overflow-hidden px-6 py-28 md:py-36 flex min-h-screen flex-col items-center justify-center text-center">
                    <div className="pointer-events-none absolute -top-40 left-1/2 -z-10 h-125 w-200 -translate-x-1/2 rounded-full bg-linear-to-tr from-primary/15 via-primary/5 to-transparent blur-3xl opacity-60" />

                    <div className="mx-auto flex max-w-4xl flex-col items-center gap-6">
                        <h1 className="font-heading text-4xl font-semibold tracking-tight text-balance sm:text-5xl md:text-6xl leading-[1.15]">
                            Digital ID cards and attendance, without the printer
                        </h1>

                        <p className="max-w-2xl text-lg sm:text-xl text-muted-foreground text-balance leading-relaxed">
                            Run conferences, campuses, and one-day meetups with QR-coded ID
                            cards, scanner-based check-in, and live attendance analytics —
                            all from one intuitive dashboard.
                        </p>

                        <div className="mt-2 flex flex-wrap items-center justify-center gap-4">
                            <Button size="lg" render={<Link href={startHref} />}>
                                Get Started
                                <RiArrowRightLine className="ml-1.5 h-4 w-4" data-icon="inline-end" />
                            </Button>
                        </div>
                    </div>
                </section>

                <Features />
                <Pricing />
            </main>

            <footer className="border-t py-8">
                <div className="mx-auto flex max-w-6xl flex-col items-center gap-4 px-6 text-center text-xs text-muted-foreground sm:flex-row sm:justify-between">
                    <span>© {new Date().getFullYear()} IdentityCard</span>
                    <div className="flex gap-4">
                        <a href="#pricing" className="hover:text-foreground transition-colors">
                            Pricing
                        </a>
                        <Link href="/login" className="hover:text-foreground transition-colors">
                            Log in
                        </Link>
                    </div>
                </div>
            </footer>
        </div>
    )
}
