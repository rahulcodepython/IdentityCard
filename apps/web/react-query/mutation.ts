"use client";

import {
    type MutationFunction,
    type QueryKey,
    useMutation,
    useQueryClient,
    type UseMutationResult,
} from "@tanstack/react-query";
import { toast } from "sonner";

export interface Identifiable {
    id: string;
}

export interface PaginatedResponse<T> {
    data: T[];
    total: number;
    page: number;
    limit: number;
}

export const appendToArray =
    <T>(item: T) =>
    (old: T[] | undefined): T[] =>
        [...(old ?? []), item];

export const prependToArray =
    <T>(item: T) =>
    (old: T[] | undefined): T[] =>
        [item, ...(old ?? [])];

export const replaceInArray =
    <T extends Identifiable>(item: T) =>
    (old: T[] | undefined): T[] =>
        (old ?? []).map((i) => (i.id === item.id ? item : i));

export const removeFromArray =
    <T extends Identifiable>(id: string) =>
    (old: T[] | undefined): T[] =>
        (old ?? []).filter((i) => i.id !== id);

export const appendToPaginated =
    <T>(item: T) =>
    (old: PaginatedResponse<T> | undefined): PaginatedResponse<T> =>
        old
            ? { ...old, data: [...old.data, item], total: old.total + 1 }
            : { data: [item], total: 1, page: 1, limit: 20 };

export const prependToPaginated =
    <T>(item: T) =>
    (old: PaginatedResponse<T> | undefined): PaginatedResponse<T> =>
        old
            ? { ...old, data: [item, ...old.data], total: old.total + 1 }
            : { data: [item], total: 1, page: 1, limit: 20 };

export const replaceInPaginated =
    <T extends Identifiable>(item: T) =>
    (old: PaginatedResponse<T> | undefined): PaginatedResponse<T> | undefined =>
        old ? { ...old, data: old.data.map((i) => (i.id === item.id ? item : i)) } : old;

export const removeFromPaginated =
    <T extends Identifiable>(id: string) =>
    (old: PaginatedResponse<T> | undefined): PaginatedResponse<T> | undefined =>
        old
            ? { ...old, data: old.data.filter((i) => i.id !== id), total: Math.max(0, old.total - 1) }
            : old;

type ToastConfig = boolean | { success?: string; error?: string };

export interface MutationConfig<TData, TVariables, TCacheData> {
    mutationFn: MutationFunction<TData, TVariables>;
    queryKey?: QueryKey;
    invalidateKeys?: QueryKey[];
    updater?: (data: TData, vars: TVariables) => (old: TCacheData | undefined) => TCacheData;
    optimistic?: (vars: TVariables) => (old: TCacheData | undefined) => TCacheData;
    showToast?: ToastConfig;
    onSuccessExtra?: (data: TData, vars: TVariables) => void;
}

interface Snapshot<TCacheData> {
    queryKey: QueryKey;
    data: TCacheData | undefined;
}

function useMutationCore<TData, TVariables, TCacheData>(
    config: MutationConfig<TData, TVariables, TCacheData>
): UseMutationResult<TData, Error, TVariables, Snapshot<TCacheData> | undefined> {
    const queryClient = useQueryClient();

    return useMutation<TData, Error, TVariables, Snapshot<TCacheData> | undefined>({
        mutationFn: config.mutationFn,

        onMutate: async (vars) => {
            if (!config.optimistic || !config.queryKey) return undefined;
            await queryClient.cancelQueries({ queryKey: config.queryKey });
            const snapshot: Snapshot<TCacheData> = {
                queryKey: config.queryKey,
                data: queryClient.getQueryData<TCacheData>(config.queryKey),
            };
            queryClient.setQueryData<TCacheData>(config.queryKey, config.optimistic(vars));
            return snapshot;
        },

        onError: (error, _vars, snapshot) => {
            if (snapshot) {
                queryClient.setQueryData(snapshot.queryKey, snapshot.data);
            }
            if (config.showToast) {
                const message =
                    typeof config.showToast === "object" && config.showToast.error
                        ? config.showToast.error
                        : error.message || "Something went wrong";
                toast.error(message);
            }
        },

        onSuccess: (data, vars) => {
            if (config.updater && config.queryKey) {
                queryClient.setQueryData<TCacheData>(config.queryKey, config.updater(data, vars));
            }
            if (config.invalidateKeys) {
                for (const key of config.invalidateKeys) {
                    queryClient.invalidateQueries({ queryKey: key });
                }
            }
            if (config.showToast) {
                const message =
                    typeof config.showToast === "object" && config.showToast.success
                        ? config.showToast.success
                        : "Done";
                toast.success(message);
            }
            config.onSuccessExtra?.(data, vars);
        },
    });
}

export function useSimpleMutation<TData, TVariables = void>(
    config: Omit<MutationConfig<TData, TVariables, never>, "updater" | "optimistic" | "queryKey"> & {
        queryKey?: QueryKey;
    }
) {
    return useMutationCore<TData, TVariables, never>(config);
}

export function useObjectMutation<TData, TVariables = void>(
    config: MutationConfig<TData, TVariables, TData>
) {
    return useMutationCore<TData, TVariables, TData>(config);
}

export function useArrayMutation<TItem, TData, TVariables = void>(
    config: MutationConfig<TData, TVariables, TItem[]>
) {
    return useMutationCore<TData, TVariables, TItem[]>(config);
}

export function usePaginatedMutation<TItem, TData, TVariables = void>(
    config: MutationConfig<TData, TVariables, PaginatedResponse<TItem>>
) {
    return useMutationCore<TData, TVariables, PaginatedResponse<TItem>>(config);
}

export function useWithExecute<TData, TError, TVariables, TContext>(
    mutation: UseMutationResult<TData, TError, TVariables, TContext>
) {
    const execute = async (vars: TVariables): Promise<TData | null> => {
        try {
            return await mutation.mutateAsync(vars);
        } catch {
            return null;
        }
    };
    return { ...mutation, execute };
}
