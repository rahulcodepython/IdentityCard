"use client";

import * as React from "react";
import {
    DEFAULT_COUNTRY_CODES,
    fetchCountryCodes,
    type CountryCode,
} from "@/lib/country-codes";
import { Input } from "@/components/ui/input";

interface PhoneInputProps {
    value?: string;
    onChange: (value: string) => void;
    placeholder?: string;
    disabled?: boolean;
    className?: string;
}

export function PhoneInput({
    value = "",
    onChange,
    placeholder = "10-digit mobile number",
    disabled = false,
    className = "",
}: PhoneInputProps) {
    const [countryCodes, setCountryCodes] = React.useState<CountryCode[]>(DEFAULT_COUNTRY_CODES);
    const [selectedDialCode, setSelectedDialCode] = React.useState<string>("+91");
    const [digits, setDigits] = React.useState<string>("");

    // Load dynamic country codes from restcountries.com in background
    React.useEffect(() => {
        let isMounted = true;
        fetchCountryCodes().then((codes) => {
            if (isMounted && codes.length > 0) {
                setCountryCodes(codes);
            }
        });
        return () => {
            isMounted = false;
        };
    }, []);

    // Sync from incoming controlled value e.g. "+919876543210"
    React.useEffect(() => {
        if (!value) {
            setDigits("");
            return;
        }

        // Check if value starts with a known dial code
        const matched = countryCodes.find((c) => value.startsWith(c.dial_code));
        if (matched) {
            setSelectedDialCode(matched.dial_code);
            const numPart = value.slice(matched.dial_code.length).replace(/\D/g, "").slice(0, 10);
            setDigits(numPart);
        } else if (value.startsWith("+")) {
            // Generic plus prefix fallback
            const match = value.match(/^(\+\d{1,4})(\d{0,10})/);
            if (match) {
                setSelectedDialCode(match[1]);
                setDigits(match[2]);
            }
        } else {
            setDigits(value.replace(/\D/g, "").slice(0, 10));
        }
    }, [value, countryCodes]);

    const handleDialCodeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        const newCode = e.target.value;
        setSelectedDialCode(newCode);
        if (digits) {
            onChange(`${newCode}${digits}`);
        } else {
            onChange("");
        }
    };

    const handleNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        // Enforce strictly digits, max length 10
        const clean = e.target.value.replace(/\D/g, "").slice(0, 10);
        setDigits(clean);
        if (clean.length > 0) {
            onChange(`${selectedDialCode}${clean}`);
        } else {
            onChange("");
        }
    };

    return (
        <div className={`flex items-center gap-1.5 ${className}`}>
            {/* Country code selector */}
            <div className="relative min-w-[100px] max-w-[110px] shrink-0">
                <select
                    value={selectedDialCode}
                    onChange={handleDialCodeChange}
                    disabled={disabled}
                    aria-label="Country calling code"
                    className="h-9 w-full rounded-md border border-input bg-transparent px-2 py-1 text-xs shadow-xs focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50 appearance-none pr-6 cursor-pointer font-mono"
                >
                    {countryCodes.map((c, i) => (
                        <option
                            key={`${c.code}-${c.dial_code}-${i}`}
                            value={c.dial_code}
                            className="bg-popover text-popover-foreground text-xs py-1"
                        >
                            {c.flag} {c.dial_code} ({c.code})
                        </option>
                    ))}
                </select>
                <div className="pointer-events-none absolute inset-y-0 right-1.5 flex items-center px-1 text-muted-foreground text-[10px]">
                    ▼
                </div>
            </div>

            {/* 10-digit number input */}
            <div className="relative flex-1">
                <Input
                    type="tel"
                    inputMode="numeric"
                    placeholder={placeholder}
                    value={digits}
                    onChange={handleNumberChange}
                    disabled={disabled}
                    maxLength={10}
                    minLength={10}
                    className="h-9 text-xs font-mono"
                />
                {digits.length > 0 && (
                    <div className="absolute right-2 top-2.5 text-[10px] text-muted-foreground font-mono">
                        {digits.length}/10
                    </div>
                )}
            </div>
        </div>
    );
}
