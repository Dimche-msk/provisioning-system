<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { goto } from "$app/navigation";
    import ThemeToggle from "./ThemeToggle.svelte";
    import LanguageSwitcher from "./LanguageSwitcher.svelte";
    import { t } from "svelte-i18n";
    async function handleLogout() {
        try {
            await fetch("/api/logout", { method: "POST" });
            goto("/login");
        } catch (e) {
            console.error("Logout failed", e);
            goto("/login");
        }
    }
</script>

<header
    class="bg-white dark:bg-gray-950 border-b dark:border-gray-800 p-4 flex justify-between items-center shadow-sm transition-colors"
>
    <div class="font-bold text-xl text-gray-900 dark:text-gray-100">
        {$t("app.title")}
    </div>
    <div class="flex items-center gap-2">
        <ThemeToggle />
        <LanguageSwitcher />
        <Button variant="outline" on:click={handleLogout}>{$t("logout")}</Button
        >
    </div>
</header>
