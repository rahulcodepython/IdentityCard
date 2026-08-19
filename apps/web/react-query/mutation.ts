"use client"

import {
  type MutationFunction,
  type QueryKey,
  useMutation,
  useQueryClient,
  type UseMutationResult,
} from "@tanstack/react-query"
import { toast } from "sonner"

export interface Identifiable {
  id: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

// ---------------------------------------------------------------------
// Cache-updater helpers — the generic building blocks every resource's
// `updater`/`optimistic` option is composed from (template §1.10).
// ---------------------------------------------------------------------

export const appendToArray =
  <T>(item: T) =>
  (old: T[] | undefined): T[] =>
    [...(old ?? []), item]

export const prependToArray =
  <T>(item: T) =>
  (old: T[] | undefined): T[] =>
    [item, ...(old ?? [])]

export const replaceInArray =
  <T extends Identifiable>(item: T) =>
  (old: T[] | undefined): T[] =>
    (old ?? []).map((i) => (i.id === item.id ? item : i))

export const removeFromArray =
  <T extends Identifiable>(id: string) =>
  (old: T[] | undefined): T[] =>
    (old ?? []).filter((i) => i.id !== id)

export const appendToPaginated =
  <T>(item: T) =>
  (old: PaginatedResponse<T> | undefined): PaginatedResponse<T> =>
    old
      ? { ...old, data: [...old.data, item], total: old.total + 1 }
      : { data: [item], total: 1, page: 1, limit: 20 }

export const prependToPaginated =
  <T>(item: T) =>
  (old: PaginatedResponse<T> | undefined): PaginatedResponse<T> =>
    old
      ? { ...old, data: [item, ...old.data], total: old.total + 1 }
      : { data: [item], total: 1, page: 1, limit: 20 }

export const replaceInPaginated =
  <T extends Identifiable>(item: T) =>
  (old: PaginatedResponse<T> | undefined): PaginatedResponse<T> | undefined =>
    old ? { ...old, data: old.data.map((i) => (i.id === item.id ? item : i)) } : old

export const removeFromPaginated =
  <T extends Identifiable>(id: string) =>
  (old: PaginatedResponse<T> | undefined): PaginatedResponse<T> | undefined =>
    old
      ? { ...old, data: old.data.filter((i) => i.id !== id), total: Math.max(0, old.total - 1) }
      : old

// ---------------------------------------------------------------------
// Mutation factory core
// ---------------------------------------------------------------------

type ToastConfig = boolean | { success?: string; error?: string }

export interface MutationConfig<TData, TVariables, TCacheData> {
  mutationFn: MutationFunction<TData, TVariables>
  /** The cache entry this mutation's `updater`/`optimistic` target. */
  queryKey?: QueryKey
  /** Extra keys to invalidate on success — combine with `updater` as a safety net, or use alone when the response can't be located deterministically. */
  invalidateKeys?: QueryKey[]
  /** Direct cache overwrite from the mutation response — the default; cheaper than invalidateKeys. */
  updater?: (data: TData, vars: TVariables) => (old: TCacheData | undefined) => TCacheData
  /** Optimistic cache write applied in onMutate, automatically rolled back on error. */
  optimistic?: (vars: TVariables) => (old: TCacheData | undefined) => TCacheData
  showToast?: ToastConfig
  onSuccessExtra?: (data: TData, vars: TVariables) => void
}

interface Snapshot<TCacheData> {
  queryKey: QueryKey
  data: TCacheData | undefined
}

function useMutationCore<TData, TVariables, TCacheData>(
  config: MutationConfig<TData, TVariables, TCacheData>
): UseMutationResult<TData, Error, TVariables, Snapshot<TCacheData> | undefined> {
  const queryClient = useQueryClient()

  return useMutation<TData, Error, TVariables, Snapshot<TCacheData> | undefined>({
    mutationFn: config.mutationFn,

    onMutate: async (vars) => {
      if (!config.optimistic || !config.queryKey) return undefined
      await queryClient.cancelQueries({ queryKey: config.queryKey })
      const snapshot: Snapshot<TCacheData> = {
        queryKey: config.queryKey,
        data: queryClient.getQueryData<TCacheData>(config.queryKey),
      }
      queryClient.setQueryData<TCacheData>(config.queryKey, config.optimistic(vars))
      return snapshot
    },

    onError: (error, _vars, snapshot) => {
      if (snapshot) {
        queryClient.setQueryData(snapshot.queryKey, snapshot.data)
      }
      if (config.showToast) {
        const message =
          typeof config.showToast === "object" && config.showToast.error
            ? config.showToast.error
            : error.message || "Something went wrong"
        toast.error(message)
      }
    },

    onSuccess: (data, vars) => {
      if (config.updater && config.queryKey) {
        queryClient.setQueryData<TCacheData>(config.queryKey, config.updater(data, vars))
      }
      if (config.invalidateKeys) {
        for (const key of config.invalidateKeys) {
          queryClient.invalidateQueries({ queryKey: key })
        }
      }
      if (config.showToast) {
        const message =
          typeof config.showToast === "object" && config.showToast.success
            ? config.showToast.success
            : "Done"
        toast.success(message)
      }
      config.onSuccessExtra?.(data, vars)
    },
  })
}

/** Actions with no convenient response→cache mapping (role assign, ban/unban, resend card...). */
export function useSimpleMutation<TData, TVariables = void>(
  config: Omit<MutationConfig<TData, TVariables, never>, "updater" | "optimistic" | "queryKey"> & {
    queryKey?: QueryKey
  }
) {
  return useMutationCore<TData, TVariables, never>(config)
}

/** Profile-style "there is exactly one of these" resources (org settings, scanner /me...). */
export function useObjectMutation<TData, TVariables = void>(
  config: MutationConfig<TData, TVariables, TData>
) {
  return useMutationCore<TData, TVariables, TData>(config)
}

/** Lists without server-side pagination (sub-events, forms, devices, roster rows...). */
export function useArrayMutation<TItem, TData, TVariables = void>(
  config: MutationConfig<TData, TVariables, TItem[]>
) {
  return useMutationCore<TData, TVariables, TItem[]>(config)
}

/** Paginated lists (events, people, members — the common admin-table case). */
export function usePaginatedMutation<TItem, TData, TVariables = void>(
  config: MutationConfig<TData, TVariables, PaginatedResponse<TItem>>
) {
  return useMutationCore<TData, TVariables, PaginatedResponse<TItem>>(config)
}

// ---------------------------------------------------------------------
// useWithExecute — every mutation hook also exposes `.execute(vars)`,
// which wraps mutateAsync, swallows the rejection, and returns null on
// failure, so call sites never repeat try/catch (template §1.7).
// ---------------------------------------------------------------------

export function useWithExecute<TData, TError, TVariables, TContext>(
  mutation: UseMutationResult<TData, TError, TVariables, TContext>
) {
  const execute = async (vars: TVariables): Promise<TData | null> => {
    try {
      return await mutation.mutateAsync(vars)
    } catch {
      return null
    }
  }
  return { ...mutation, execute }
}
