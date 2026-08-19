import { listOrgSubscriptions, listPlans } from "@/lib/api/plans"
import { getOrgSettings } from "@/lib/api/organizations"

import { BillingClient } from "./billing-client"

export default async function BillingPage() {
  const [plans, subs, orgSettings] = await Promise.all([
    listPlans(),
    listOrgSubscriptions(),
    getOrgSettings().catch(() => null),
  ])

  return (
    <BillingClient
      plans={plans}
      subs={subs}
      organizationName={orgSettings?.name || "IdentityCard Org"}
    />
  )
}
