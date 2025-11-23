<script>
    import { cn } from "$lib/utils";
    import { badgeVariants } from "./index.ts";

    export let variant = "default";
    let className = undefined;
    export { className as class };
</script>

<div class={cn(badgeVariants({ variant }), className)} {...$$restProps}>
    <slot />
</div>
