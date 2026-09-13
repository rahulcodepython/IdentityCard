"use client";

import {
    type MutationFunction,
    type QueryKey,
    useMutation,
    useQueryClient,
    type UseMutationResult,
} from "@tanstack/react-query";
import { toast } from "sonner";

export interface MutationOptions<TData, TVariables = void, TCache = TData> {
    mutationFn: MutationFunction<TData, TVariables>;
    queryKey?: QueryKey;
    // 1. Reuse response data instead of invalidating existing cache data
    updateCache?: (data: TData, old: TCache | undefined) => TCache;
    // 2. Optimistic update before request completes
    optimisticUpdate?: (vars: TVariables, old: TCache | undefined) => TCache;
    onSuccess?: (data: TData, vars: TVariables) => void;
    onError?: (error: Error, vars: TVariables) => void;
    successMessage?: string;
    errorMessage?: string;
}

export function useCustomMutation<TData, TVariables = void, TCache = TData>(
    options: MutationOptions<TData, TVariables, TCache>,
): UseMutationResult<TData, Error, TVariables, { previous?: TCache }> {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: options.mutationFn,

        onMutate: async (vars) => {
            if (!options.queryKey || !options.optimisticUpdate) {
                return {};
            }
            await queryClient.cancelQueries({ queryKey: options.queryKey });
            const previous = queryClient.getQueryData<TCache>(options.queryKey);
            queryClient.setQueryData<TCache>(options.queryKey, (old) =>
                options.optimisticUpdate!(vars, old),
            );
            return { previous };
        },

        onError: (error, vars, context) => {
            if (options.queryKey && context?.previous !== undefined) {
                queryClient.setQueryData(options.queryKey, context.previous);
            }
            toast.error(options.errorMessage || error.message || "Something went wrong");
            options.onError?.(error, vars);
        },

        onSuccess: (data, vars) => {
            if (options.queryKey && options.updateCache) {
                queryClient.setQueryData<TCache>(options.queryKey, (old) =>
                    options.updateCache!(data, old),
                );
            }
            if (options.successMessage) {
                toast.success(options.successMessage);
            }
            options.onSuccess?.(data, vars);
        },
    });
}
