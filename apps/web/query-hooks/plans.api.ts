"use client"

import { apiRequest } from "@/react-query/client"
import { useAppQuery } from "@/react-query/query"
import { queryKeys } from "@/react-query/query-keys"
import { type OrgSubscriptions, orgSubscriptionsResponseSchema } from "@/schema/plans.types"

// Minimal for now — just what dashboard/layout.tsx needs for the
// "subscription expired" banner. Phase 4 extends this file for the full
// billing page (listPlans/subscribe/renew).
export function useListSubscriptionsQuery(enabled = true) {
  return useAppQuery<OrgSubscriptions>(
    queryKeys.orgSubscriptions(),
    () =>
      apiRequest(
        { url: "/plans/subscriptions", method: "GET" },
        orgSubscriptionsResponseSchema
      ),
    { enabled }
  )
}
