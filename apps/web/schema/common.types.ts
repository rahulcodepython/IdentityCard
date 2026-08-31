import { z } from "zod";

export const ApiResponseZod = <T extends z.ZodTypeAny>(dataSchema: T) =>
    z.object({
        success: z.boolean(),
        message: z.string(),
        data: dataSchema.nullable().optional(),
        error: z.any().optional().nullable(),
    });

export type ApiResponse<T> = {
    success: boolean;
    message: string;
    data: T | null;
    error?: string | Record<string, string> | null;
};

export const PaginatedResponseZod = <T extends z.ZodTypeAny>(itemSchema: T) =>
    z.object({
        items: z.array(itemSchema),
        total: z.number(),
        page: z.number(),
        page_size: z.number(),
        total_pages: z.number(),
    });

export type PaginatedResponse<T> = {
    items: T[];
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
};

export const DeleteResponseZod = z.object({
    id: z.string().uuid().optional(),
    success: z.boolean().optional(),
});

export const SuccessResponseZod = z.object({
    message: z.string().optional(),
});

