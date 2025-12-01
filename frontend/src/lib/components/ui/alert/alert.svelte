<script lang="ts">
    import { type VariantProps, tv } from "tailwind-variants";
    import type { HTMLAttributes } from "svelte/elements";

    const alertVariants = tv({
        base: "relative w-full rounded-lg border p-4 [&>svg~*]:pl-7 [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground",
        variants: {
            variant: {
                default: "bg-background text-foreground",
                destructive:
                    "border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive",
            },
        },
        defaultVariants: {
            variant: "default",
        },
    });

    type Variant = VariantProps<typeof alertVariants>["variant"];

    interface $$Props extends HTMLAttributes<HTMLDivElement> {
        variant?: Variant;
    }

    let className: string | undefined | null = undefined;
    export { className as class };
    export let variant: Variant = "default";
</script>

<div
    role="alert"
    class={alertVariants({ variant, className })}
    {...$$restProps}
>
    <slot />
</div>
