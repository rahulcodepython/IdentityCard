export const queryKeys = {
    events: {
        all: ["events"] as const,
        list: (filters?: { search?: string }) => ["events", "list", filters] as const,
        detail: (id: string) => ["events", "detail", id] as const,
    },
} as const;
