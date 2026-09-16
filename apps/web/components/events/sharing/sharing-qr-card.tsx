"use client";

import * as React from "react";
import QRCode from "qrcode";
import { Check, Copy, Download, ExternalLink, QrCode } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";

interface SharingQrCardProps {
    eventFormId?: string;
    isLive: boolean;
}

export function SharingQrCard({ eventFormId, isLive }: SharingQrCardProps) {
    const [qrCodeDataUrl, setQrCodeDataUrl] = React.useState<string>("");
    const [copied, setCopied] = React.useState(false);

    const publicUrl = React.useMemo(() => {
        if (!eventFormId) return "";
        if (typeof window !== "undefined") {
            return `${window.location.origin}/apply/${eventFormId}`;
        }
        return `/apply/${eventFormId}`;
    }, [eventFormId]);

    // Generate high-resolution QR code whenever publicUrl changes
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
            toast.success("Registration URL copied to clipboard");
            setTimeout(() => setCopied(false), 2000);
        } catch {
            toast.error("Failed to copy URL");
        }
    };

    const handleDownloadPng = () => {
        if (!qrCodeDataUrl) return;
        const link = document.createElement("a");
        link.download = `event-registration-qr-${eventFormId || "code"}.png`;
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
            link.download = `event-registration-qr-${eventFormId || "code"}.svg`;
            link.href = url;
            link.click();
            URL.revokeObjectURL(url);
            toast.success("QR code downloaded as SVG");
        } catch {
            toast.error("Failed to generate SVG QR code");
        }
    };

    if (!eventFormId) {
        return (
            <Card className="shadow-xs border-dashed">
                <CardHeader>
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                        <QrCode className="size-4 text-muted-foreground" />
                        <span>Public Link & QR Code</span>
                    </CardTitle>
                    <CardDescription className="text-xs">
                        Configure and assign a form on the left to generate the public registration QR code.
                    </CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col items-center justify-center p-8 text-center text-xs text-muted-foreground">
                    <span>No active registration link yet.</span>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className="shadow-xs">
            <CardHeader className="pb-3">
                <CardTitle className="text-sm font-semibold flex items-center gap-2">
                    <QrCode className="size-4 text-primary" />
                    <span>Public Link & QR Code</span>
                </CardTitle>
                <CardDescription className="text-xs">
                    Share this link or print the QR code on badges, flyers, or slides.
                </CardDescription>
            </CardHeader>

            <CardContent className="space-y-4">
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
                        Scan with any smartphone camera
                    </span>
                </div>

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
                            title="Copy URL"
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
                        size="sm"
                        onClick={handleDownloadPng}
                        className="gap-1.5 text-xs"
                    >
                        <Download className="size-3" />
                        <span>PNG</span>
                    </Button>
                    <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={handleDownloadSvg}
                        className="gap-1.5 text-xs"
                    >
                        <Download className="size-3" />
                        <span>SVG</span>
                    </Button>
                </div>

                <div className="pt-1">
                    <Button
                        type="button"
                        variant="secondary"
                        size="sm"
                        className="w-full gap-2 text-xs font-medium"
                        onClick={() => window.open(publicUrl, "_blank")}
                    >
                        <ExternalLink className="size-3.5" />
                        <span>Open Registration Page</span>
                    </Button>
                </div>
            </CardContent>
        </Card>
    );
}
