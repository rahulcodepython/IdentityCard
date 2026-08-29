import "server-only"

import { PayKit } from "@paykit-sdk/core"

import { ManualProvider } from "@/lib/paykit-manual-provider"

// Swap ManualProvider for `new (await import("@paykit-sdk/razorpay")).RazorpayProvider(...)`
// once out of testing — nothing else in the checkout flow (lib/actions/checkout.ts)
// needs to change, since it only ever talks to this `paykit` instance.
export const paykit = new PayKit(new ManualProvider())
