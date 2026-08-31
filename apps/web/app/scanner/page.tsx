"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import QrScanner from "qr-scanner";

import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import {
    clearDeviceKey,
    DeviceApiError,
    getDeviceKey,
    getScannerMe,
    scanQr,
} from "@/lib/device-client";
import type { ScannerMe, ScanResponse } from "@/schema/scanner.types";

import { ScanResultCard } from "./scan-result-card";

export default function ScannerPage() {
    const router = useRouter();
    const videoRef = useRef<HTMLVideoElement>(null);
    const scannerRef = useRef<QrScanner | null>(null);

    const [me, setMe] = useState<ScannerMe | null>(null);
    const [scanning, setScanning] = useState(false);
    const [isProcessing, setIsProcessing] = useState(false);
    const [result, setResult] = useState<ScanResponse | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!getDeviceKey()) {
            router.replace("/pair");
            return;
        }
        getScannerMe()
            .then(setMe)
            .catch((err: unknown) => {
                if (err instanceof DeviceApiError && err.status === 401) {
                    clearDeviceKey();
                    router.replace("/pair");
                }
            });
    }, [router]);

    useEffect(() => {
        return () => {
            scannerRef.current?.destroy();
        };
    }, []);

    const stopScanning = () => {
        scannerRef.current?.stop();
        scannerRef.current?.destroy();
        scannerRef.current = null;
        setScanning(false);
    };

    const handleDecode = (qrToken: string) => {
        stopScanning();
        setIsProcessing(true);
        scanQr(qrToken)
            .then(setResult)
            .catch((err: unknown) => {
                setError(
                    err instanceof DeviceApiError
                        ? err.message
                        : "Could not verify this code."
                );
            })
            .finally(() => setIsProcessing(false));
    };

    const startScanning = async () => {
        setError(null);
        setResult(null);
        if (!videoRef.current) return;

        const scanner = new QrScanner(
            videoRef.current,
            (scanResult) => handleDecode(scanResult.data),
            {
                returnDetailedScanResult: true,
                highlightScanRegion: true,
            }
        );
        scannerRef.current = scanner;
        setScanning(true);
        await scanner.start();
    };

    return (
        <div className="flex min-h-svh flex-col items-center gap-6 p-6">
            <div className="w-full max-w-sm text-center">
                <h1 className="font-heading text-lg font-medium">
                    {me ? me.organization_name : "Scanner"}
                </h1>
                {me && (
                    <p className="text-xs text-muted-foreground">{me.device_name}</p>
                )}
            </div>

            {result ? (
                <ScanResultCard
                    result={result}
                    onDismiss={() => {
                        setResult(null);
                        startScanning();
                    }}
                />
            ) : (
                <Card className="w-full max-w-sm">
                    <CardHeader>
                        <CardTitle>Scan an ID card</CardTitle>
                        <CardDescription>Point the camera at the QR code.</CardDescription>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-4">
                        <video
                            ref={videoRef}
                            className="aspect-square w-full rounded-lg bg-black object-cover"
                            muted
                            playsInline
                        />
                        {error && <p className="text-sm text-destructive">{error}</p>}
                        {scanning ? (
                            <Button variant="outline" onClick={stopScanning}>
                                Cancel
                            </Button>
                        ) : (
                            <Button
                                onClick={startScanning}
                                disabled={isProcessing}
                                className="h-16 text-lg"
                            >
                                {isProcessing ? "Verifying…" : "Scan"}
                            </Button>
                        )}
                    </CardContent>
                </Card>
            )}
        </div>
    );
}
