"use client";

import { Button } from "@/components/ui/button";
import { usePublishEventMutation } from "@/query-hooks/events.api";

export function PublishButton({ eventId }: { eventId: string }) {
    const publishMutation = usePublishEventMutation(eventId);

    return (
        <div className="flex flex-col items-end gap-1">
            <Button
                disabled={publishMutation.isPending}
                onClick={async () => {
                    await publishMutation.execute();
                }}
            >
                {publishMutation.isPending ? "Publishing…" : "Publish"}
            </Button>
            {publishMutation.error && (
                <p className="text-xs text-destructive">{publishMutation.error.message}</p>
            )}
        </div>
    );
}