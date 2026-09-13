import Link from "next/link"
import { Check, Sparkles } from "lucide-react"

import { Button } from "@/components/ui/button"

export default function Pricing() {
    return (
        <section id="pricing" className="px-6 py-24 border-t bg-muted/30">
            <div className="mx-auto max-w-6xl">
                <div className="text-center max-w-2xl mx-auto mb-16 space-y-4">
                    <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-primary/10 text-primary">
                        <Sparkles className="size-3.5" />
                        <span>Transparent & Simple</span>
                    </div>
                    <h2 className="font-heading text-3xl font-bold tracking-tight sm:text-4xl">
                        Simple, event-based pricing
                    </h2>
                    <p className="text-muted-foreground text-base">
                        No confusing tiers or seat licenses. Pay only when you create events, and never pay a rupee if you have no events.
                    </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-8 items-stretch">
                    {/* Free Starter */}
                    <div className="flex flex-col justify-between rounded-2xl border bg-card p-8 shadow-sm">
                        <div className="space-y-4">
                            <div>
                                <h3 className="text-xl font-bold">Free Welcome</h3>
                                <p className="text-xs text-muted-foreground mt-1">For every new organization</p>
                            </div>
                            <div className="flex items-baseline gap-1">
                                <span className="text-4xl font-extrabold tracking-tight">₹0</span>
                                <span className="text-xs text-muted-foreground">free forever</span>
                            </div>
                            <p className="text-sm text-muted-foreground">
                                Get started immediately with 1 free event credit on sign up.
                            </p>
                            <ul className="space-y-2.5 pt-4 text-sm text-foreground/90 border-t">
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>1 Event Credit automatically granted</span>
                                </li>
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>Full badge designer & QR generator</span>
                                </li>
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>Unlimited scanner devices</span>
                                </li>
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>Zero annual fee if 0 events stored</span>
                                </li>
                            </ul>
                        </div>
                        <div className="pt-8">
                            <Button variant="outline" className="w-full" render={<Link href="/auth/login" />}>
                                Get Started Free
                            </Button>
                        </div>
                    </div>

                    {/* Pay As You Go */}
                    <div className="relative flex flex-col justify-between rounded-2xl border-2 border-primary bg-card p-8 shadow-md">
                        <div className="absolute -top-3 left-1/2 -translate-x-1/2 px-3 py-0.5 rounded-full text-xs font-semibold bg-primary text-primary-foreground tracking-wide uppercase">
                            Most Popular
                        </div>
                        <div className="space-y-4">
                            <div>
                                <h3 className="text-xl font-bold">Event Credit</h3>
                                <p className="text-xs text-muted-foreground mt-1">On-demand pay-per-event</p>
                            </div>
                            <div className="flex items-baseline gap-1">
                                <span className="text-4xl font-extrabold tracking-tight">₹1,499</span>
                                <span className="text-xs text-muted-foreground">/ credit</span>
                            </div>
                            <p className="text-sm text-muted-foreground">
                                1 credit lets you launch 1 complete event of any duration.
                            </p>
                            <ul className="space-y-2.5 pt-4 text-sm text-foreground/90 border-t">
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>1 Credit = 1 Event Creation</span>
                                </li>
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>No duration limits (single or multi-day)</span>
                                </li>
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>Purchase anytime, credits never expire</span>
                                </li>
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>Real-time scan sync & live check-in stats</span>
                                </li>
                            </ul>
                        </div>
                        <div className="pt-8">
                            <Button className="w-full" render={<Link href="/auth/login" />}>
                                Buy Credits
                            </Button>
                        </div>
                    </div>

                    {/* Annual Maintenance */}
                    <div className="flex flex-col justify-between rounded-2xl border bg-card p-8 shadow-sm">
                        <div className="space-y-4">
                            <div>
                                <h3 className="text-xl font-bold">Annual Retention</h3>
                                <p className="text-xs text-muted-foreground mt-1">Data & event archiving</p>
                            </div>
                            <div className="flex items-baseline gap-1">
                                <span className="text-4xl font-extrabold tracking-tight">₹2,999</span>
                                <span className="text-xs text-muted-foreground">/ year</span>
                            </div>
                            <p className="text-sm text-muted-foreground">
                                Flat annual maintenance, billed only if you retain events.
                            </p>
                            <ul className="space-y-2.5 pt-4 text-sm text-foreground/90 border-t">
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>Preserves historical attendance data</span>
                                </li>
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>Exports & attendee lookup anytime</span>
                                </li>
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span>Flat fee regardless of event volume</span>
                                </li>
                                <li className="flex items-center gap-2">
                                    <Check className="size-4 text-primary shrink-0" />
                                    <span><strong>100% Free</strong> if you have 0 events</span>
                                </li>
                            </ul>
                        </div>
                        <div className="pt-8">
                            <Button variant="outline" className="w-full" render={<Link href="/auth/login" />}>
                                Explore Platform
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
        </section>
    )
}
