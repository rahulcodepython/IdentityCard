"use client"

import {
  type QueryKey,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/react-query"

export function useAppQuery<TData, TError = Error>(
  queryKey: QueryKey,
  queryFn: () => Promise<TData>,
  options?: Omit<UseQueryOptions<TData, TError>, "queryKey" | "queryFn">
) {
  return useQuery<TData, TError>({ queryKey, queryFn, ...options })
}
