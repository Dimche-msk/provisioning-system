<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card";
    import { goto } from "$app/navigation";
    import { t } from "svelte-i18n";

    let username = "";
    let password = "";
    let error = "";
    let loading = false;

    async function handleLogin() {
        loading = true;
        error = "";

        try {
            const response = await fetch("/api/login", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ username, password }),
            });

            if (response.ok) {
                goto("/");
            } else {
                error = $t("login.error");
            }
        } catch (e) {
            error = $t("login.error");
        } finally {
            loading = false;
        }
    }
</script>

<div
    class="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900 p-4 transition-colors"
>
    <Card.Root class="w-full max-w-md">
        <Card.Header>
            <Card.Title class="text-2xl font-bold text-center"
                >{$t("app.title")}</Card.Title
            >
            <Card.Description class="text-center"
                >{$t("login.description")}</Card.Description
            >
        </Card.Header>
        <Card.Content class="space-y-4">
            {#if error}
                <div
                    class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative dark:bg-red-900 dark:border-red-700 dark:text-red-100"
                    role="alert"
                >
                    <span class="block sm:inline">{error}</span>
                </div>
            {/if}
            <div class="space-y-2">
                <Label for="username">{$t("login.username")}</Label>
                <Input
                    id="username"
                    type="text"
                    placeholder="login"
                    bind:value={username}
                />
            </div>
            <div class="space-y-2">
                <Label for="password">{$t("login.password")}</Label>
                <Input
                    id="password"
                    type="password"
                    placeholder="••••••••"
                    bind:value={password}
                />
            </div>
        </Card.Content>
        <Card.Footer>
            <Button class="w-full" on:click={handleLogin} disabled={loading}>
                {loading ? $t("login.loading") : $t("login.button")}
            </Button>
        </Card.Footer>
    </Card.Root>
</div>
