"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { useResendCardMutation } from "@/query-hooks/cards.api";

export function CardActions({
    eventId,
    personId,
    downloadHref,
    eventPublished,
    cardSentAt,
}: {
    eventId: string;
    personId: string;
    downloadHref: string;
    eventPublished: boolean;
    cardSentAt?: string | null;
}) {
    const resendMutation = useResendCardMutation(eventId);
    const [sent, setSent] = useState(false);

    return (
        <div className="flex flex-col gap-1 rounded-lg border p-3">
            <div className="flex items-center gap-2">
                <Button
                    variant="outline"
                    size="sm"
                    render={<a href={downloadHref} target="_blank" rel="noreferrer" />}
                >
                    Download card
                </Button>
                <Button
                    variant="outline"
                    size="sm"
                    disabled={resendMutation.isPending || !eventPublished}
                    onClick={async () => {
                        const res = await resendMutation.execute(personId);
                        if (res) {
                            setSent(true);
                        }
                    }}
                >
                    {resendMutation.isPending ? "Sending…" : "Resend by email"}
                </Button>
            </div>
            <p className="text-xs text-muted-foreground">
                {!eventPublished
                    ? "Emailing is available once the event is published — download works now for a preview."
                    : cardSentAt
                        ? `Last emailed ${new Date(cardSentAt).toLocaleString()}`
                        : "Not yet emailed."}
                {sent && " · Resent just now."}
            </p>
            {resendMutation.error && (
                <p className="text-xs text-destructive">{resendMutation.error.message}</p>
            )}
        </div>
    );
}