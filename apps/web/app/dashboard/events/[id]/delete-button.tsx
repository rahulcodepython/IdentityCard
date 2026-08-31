"use client";

import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { useDeleteEventMutation } from "@/query-hooks/events.api";

export function DeleteButton({ eventId }: { eventId: string }) {
    const router = useRouter();
    const deleteMutation = useDeleteEventMutation(eventId);

    return (
        <div className="flex flex-col items-end gap-1">
            <Button
                variant="destructive"
                disabled={deleteMutation.isPending}
                onClick={async () => {
                    const res = await deleteMutation.execute();
                    if (res !== null) {
                        router.push("/dashboard/events");
                    }
                }}
            >
                {deleteMutation.isPending ? "Deleting…" : "Delete"}
            </Button>
            {deleteMutation.error && (
                <p className="text-xs text-destructive">{deleteMutation.error.message}</p>
            )}
        </div>
    );
}
