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

import { cubicOut } from "svelte/easing";
import type { TransitionConfig } from "svelte/transition";

type FlyAndScaleParams = {
	y?: number;
	x?: number;
	start?: number;
	duration?: number;
};

export const flyAndScale = (
	node: Element,
	params: FlyAndScaleParams = { y: -8, x: 0, start: 0.95, duration: 150 }
): TransitionConfig => {
	const style = getComputedStyle(node);
	const transform = style.transform === "none" ? "" : style.transform;

	return {
		duration: params.duration ?? 200,
		delay: 0,
		easing: cubicOut,
		css: (t) => {
			const y = t * (params.y ?? 0);
			const x = t * (params.x ?? 0);
			const scale = (params.start ?? 0.95) + (1 - (params.start ?? 0.95)) * t;

			return `
				transform: ${transform} translate3d(${x}px, ${y}px, 0) scale(${scale});
				opacity: ${t}
			`;
		},
	};
};
