"use client";

import * as React from "react";
import QRCode from "qrcode";
import { Check, Copy, Download, ExternalLink, QrCode } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import type { EventFormDetails } from "@/schema/eventform.types";

interface QrSharingDialogProps {
    eventForm: EventFormDetails;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function QrSharingDialog({
    eventForm,
    open,
    onOpenChange,
}: QrSharingDialogProps) {
    const [qrCodeDataUrl, setQrCodeDataUrl] = React.useState<string>("");
    const [copied, setCopied] = React.useState(false);

    const publicUrl = React.useMemo(() => {
        if (!eventForm.id) return "";
        if (typeof window !== "undefined") {
            return `${window.location.origin}/apply/${eventForm.id}`;
        }
        return `/apply/${eventForm.id}`;
    }, [eventForm.id]);

    React.useEffect(() => {
        if (!publicUrl) {
            setQrCodeDataUrl("");
            return;
        }

        let isCancelled = false;
        QRCode.toDataURL(publicUrl, {
            width: 320,
            margin: 2,
            color: {
                dark: "#000000",
                light: "#ffffff",
            },
        })
            .then((url) => {
                if (!isCancelled) {
                    setQrCodeDataUrl(url);
                }
            })
            .catch(() => {
                // Ignore error
            });

        return () => {
            isCancelled = true;
        };
    }, [publicUrl]);

    const handleCopy = async () => {
        if (!publicUrl) return;
        try {
            await navigator.clipboard.writeText(publicUrl);
            setCopied(true);
            toast.success("Registration link copied to clipboard");
            setTimeout(() => setCopied(false), 2000);
        } catch {
            toast.error("Failed to copy link");
        }
    };

    const handleDownloadPng = () => {
        if (!qrCodeDataUrl) return;
        const link = document.createElement("a");
        link.download = `event-registration-qr-${eventForm.id}.png`;
        link.href = qrCodeDataUrl;
        link.click();
        toast.success("QR code downloaded as PNG");
    };

    const handleDownloadSvg = async () => {
        if (!publicUrl) return;
        try {
            const svgString = await QRCode.toString(publicUrl, {
                type: "svg",
                margin: 2,
            });
            const blob = new Blob([svgString], { type: "image/svg+xml" });
            const url = URL.createObjectURL(blob);
            const link = document.createElement("a");
            link.download = `event-registration-qr-${eventForm.id}.svg`;
            link.href = url;
            link.click();
            URL.revokeObjectURL(url);
            toast.success("QR code downloaded as SVG");
        } catch {
            toast.error("Failed to generate SVG QR code");
        }
    };

    const isFull =
        eventForm.max_applicants !== -1 &&
        eventForm.total_applicants >= eventForm.max_applicants;
    const isExpired = new Date(eventForm.expires_at) < new Date();

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2 text-base font-semibold">
                        <QrCode className="size-4 text-primary" />
                        <span>Share Registration & QR Code</span>
                    </DialogTitle>
                    <DialogDescription className="text-xs">
                        Distribute this link or print the QR code on badges, banners, or presentation slides.
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-4 pt-2">
                    {/* QR Code Presentation */}
                    <div className="flex flex-col items-center justify-center p-4 rounded-xl border bg-white shadow-xs">
                        {qrCodeDataUrl ? (
                            // eslint-disable-next-line @next/next/no-img-element
                            <img
                                src={qrCodeDataUrl}
                                alt="Registration QR Code"
                                className="size-48 object-contain rounded-md"
                            />
                        ) : (
                            <div className="size-48 flex items-center justify-center bg-muted/30 rounded-md text-xs text-muted-foreground">
                                Generating QR Code...
                            </div>
                        )}
                        <span className="text-[11px] text-muted-foreground mt-2 font-medium">
                            Scan with any smartphone camera to register
                        </span>
                    </div>

                    {/* Status badges */}
                    <div className="flex items-center justify-between px-1 text-xs text-muted-foreground">
                        <span>
                            Applicants:{" "}
                            <strong className="text-foreground font-semibold">
                                {eventForm.total_applicants}
                            </strong>{" "}
                            /{" "}
                            {eventForm.max_applicants === -1
                                ? "Unlimited"
                                : eventForm.max_applicants}
                        </span>
                        <span>
                            Deadline:{" "}
                            <strong className="text-foreground font-semibold">
                                {new Date(eventForm.expires_at).toLocaleDateString()}
                            </strong>
                        </span>
                    </div>

                    {isFull && (
                        <div className="rounded-md border border-amber-500/20 bg-amber-500/10 p-2.5 text-xs text-amber-800 dark:text-amber-300">
                            Registration capacity has been reached. Additional submissions will be blocked.
                        </div>
                    )}

                    {isExpired && (
                        <div className="rounded-md border border-destructive/20 bg-destructive/10 p-2.5 text-xs text-destructive">
                            Registration deadline has expired. Submissions will be blocked.
                        </div>
                    )}

                    {/* Direct Link Input with Copy */}
                    <div className="space-y-1.5">
                        <label className="text-xs font-medium text-foreground">
                            Public Registration URL
                        </label>
                        <div className="flex items-center gap-2">
                            <Input
                                readOnly
                                value={publicUrl}
                                className="text-xs font-mono bg-muted/40"
                            />
                            <Button
                                type="button"
                                variant="outline"
                                size="icon"
                                onClick={handleCopy}
                                className="shrink-0 size-8"
                                title="Copy Link"
                            >
                                {copied ? (
                                    <Check className="size-3.5 text-emerald-500" />
                                ) : (
                                    <Copy className="size-3.5" />
                                )}
                            </Button>
                        </div>
                    </div>

                    {/* Action Buttons */}
                    <div className="grid grid-cols-2 gap-2 pt-1">
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleDownloadPng}
                            className="gap-1.5 text-xs"
                        >
                            <Download className="size-3" />
                            <span>Download PNG</span>
                        </Button>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleDownloadSvg}
                            className="gap-1.5 text-xs"
                        >
                            <Download className="size-3" />
                            <span>Download SVG</span>
                        </Button>
                    </div>

                    <div className="pt-1">
                        <Button
                            variant="outline"
                            className="w-full gap-2 text-xs font-medium"
                            onClick={() => window.open(publicUrl, "_blank")}
                        >
                            <ExternalLink className="size-4" />
                            <span>Open Registration Page</span>
                        </Button>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
}
