import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

export function formatMacInput(value: string | null | undefined): string {
    if (!value) return "";
    // Remove non-hex characters
    const clean = value.replace(/[^a-fA-F0-9]/g, "");
    // Add colons
    const parts = clean.match(/.{1,2}/g) || [];
    return parts.join(":").substring(0, 17); // Limit to XX:XX:XX:XX:XX:XX
}
