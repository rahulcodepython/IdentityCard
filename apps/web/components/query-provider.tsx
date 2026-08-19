"use client"

import { useState } from "react"
import { QueryCache, QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { toast } from "sonner"

import { ApiError } from "@/react-query/client"

// One QueryClient per app, constructed once (template §1.8). staleTime is
// the one app-wide default; override per-query only when a resource
// genuinely needs fresher/staler data. No MutationCache.onError here — the
// mutation factory (react-query/mutation.ts) already toasts per-mutation
// via its own showToast option, so this only owns query errors.
export function QueryProvider({ children }: { children: React.ReactNode }) {
  const [client] = useState(
    () =>
      new QueryClient({
        queryCache: new QueryCache({
          // A 401 here is normal (an auth gate deciding to redirect, e.g.
          // dashboard/layout.tsx's useMeQuery) — not something to surface
          // as an error toast.
          onError: (error) => {
            if (error instanceof ApiError && error.status === 401) return
            toast.error(error.message || "Something went wrong")
          },
        }),
        defaultOptions: {
          queries: {
            staleTime: 60_000,
            refetchOnWindowFocus: false,
            retry: 1,
          },
        },
      })
  )

  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}
