"use client";

import * as React from "react";
import { Check, Copy, Download, QrCode } from "lucide-react";
import QRCode from "qrcode";
import { toast } from "sonner";

import { Button } from "../ui/button";
import {
    Dialog,
    DialogBody,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "../ui/dialog";
import type { ApplicantItem } from "../../schema/applicants.types";

interface ApplicantQrDialogProps {
    applicant: ApplicantItem;
    eventId: string;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function ApplicantQrDialog({
    applicant,
    eventId,
    open,
    onOpenChange,
}: ApplicantQrDialogProps) {
    const [qrDataUrl, setQrDataUrl] = React.useState<string>("");
    const [copiedEncoded, setCopiedEncoded] = React.useState(false);

    // 1. Prepare raw JSON object
    const rawPayload = React.useMemo(() => {
        return {
            eventID: eventId,
            applicantID: applicant.user_id,
            registeredAt: applicant.created_at,
        };
    }, [eventId, applicant.user_id, applicant.created_at]);

    // 2. Generate encoded string from the JSON data
    const encodedPayload = React.useMemo(() => {
        const jsonStr = JSON.stringify(rawPayload);
        try {
            return typeof window !== "undefined"
                ? btoa(unescape(encodeURIComponent(jsonStr)))
                : Buffer.from(jsonStr).toString("base64");
        } catch {
            return btoa(jsonStr);
        }
    }, [rawPayload]);

    // 3. Generate the QR code over the encoded string value
    React.useEffect(() => {
        if (!open) return;

        QRCode.toDataURL(encodedPayload, {
            width: 320,
            margin: 2,
            color: {
                dark: "#000000",
                light: "#ffffff",
            },
            errorCorrectionLevel: "H",
        })
            .then((url) => {
                setQrDataUrl(url);
            })
            .catch(() => {
                toast.error("Failed to generate QR Code");
            });
    }, [open, encodedPayload]);

    const handleCopyEncoded = () => {
        navigator.clipboard.writeText(encodedPayload);
        setCopiedEncoded(true);
        setTimeout(() => setCopiedEncoded(false), 2000);
        toast.success("Encoded QR string copied to clipboard");
    };

    const handleDownload = () => {
        if (!qrDataUrl) return;
        const link = document.createElement("a");
        link.download = `qr-${applicant.user_id}.png`;
        link.href = qrDataUrl;
        link.click();
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-lg">
                <DialogHeader>
                    <div className="flex items-center gap-2">
                        <div className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                            <QrCode className="size-4" />
                        </div>
                        <div>
                            <DialogTitle className="text-base font-semibold">
                                Applicant QR Code
                            </DialogTitle>
                            <DialogDescription className="text-xs">
                                Digital ticket badge generated from encoded JSON payload.
                            </DialogDescription>
                        </div>
                    </div>
                </DialogHeader>

                <DialogBody className="space-y-4">
                    <div className="flex flex-col items-center justify-center">
                        <div className="flex size-64 items-center justify-center rounded-xl border border-border bg-white p-3 shadow-xs">
                            {qrDataUrl ? (
                                <img
                                    src={qrDataUrl}
                                    alt={`QR Code for ${applicant.name}`}
                                    className="size-full object-contain"
                                />
                            ) : (
                                <div className="text-xs text-muted-foreground">
                                    Generating QR Code...
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Encoded String display */}
                    <div className="w-full rounded-lg border bg-muted/30 p-3 text-xs space-y-1">
                        <div className="flex items-center justify-between">
                            <span className="text-[10px] text-muted-foreground font-semibold uppercase tracking-wider">
                                Encoded Payload
                            </span>
                            <button
                                type="button"
                                onClick={handleCopyEncoded}
                                className="flex items-center gap-1 text-[10px] text-primary hover:underline font-medium cursor-pointer"
                            >
                                {copiedEncoded ? (
                                    <Check className="size-3" />
                                ) : (
                                    <Copy className="size-3" />
                                )}
                                <span>
                                    {copiedEncoded ? "Copied" : "Copy Encoded String"}
                                </span>
                            </button>
                        </div>
                        <p className="font-mono text-[10px] text-foreground/80 break-all line-clamp-2 select-all">
                            {encodedPayload}
                        </p>
                    </div>

                    <div className="w-full rounded-lg border bg-muted/20 p-3 text-xs space-y-1.5">
                        <div className="flex justify-between text-muted-foreground">
                            <span>Attendee</span>
                            <span className="font-semibold text-foreground">
                                {applicant.name}
                            </span>
                        </div>
                        <div className="flex justify-between text-muted-foreground">
                            <span>Applicant ID</span>
                            <span className="font-mono text-foreground">
                                {applicant.user_id}
                            </span>
                        </div>
                        <div className="flex justify-between text-muted-foreground">
                            <span>Email</span>
                            <span className="text-foreground">
                                {applicant.email}
                            </span>
                        </div>
                        {applicant.phone && (
                            <div className="flex justify-between text-muted-foreground">
                                <span>Mobile</span>
                                <span className="font-mono text-foreground">
                                    {applicant.phone}
                                </span>
                            </div>
                        )}
                    </div>
                </DialogBody>

                <DialogFooter>
                    <Button
                        type="button"
                        variant="outline"
                        onClick={handleCopyEncoded}
                    >
                        <Copy className="size-4" />
                        <span>Copy Encoded</span>
                    </Button>
                    <Button
                        type="button"
                        variant="default"
                        onClick={handleDownload}
                        disabled={!qrDataUrl}
                    >
                        <Download className="size-4" />
                        <span>Download Badge</span>
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
