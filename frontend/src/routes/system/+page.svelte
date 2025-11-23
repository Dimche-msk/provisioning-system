<script>
    import { Button } from "$lib/components/ui/button";
    import * as Card from "$lib/components/ui/card";
    import { t } from "svelte-i18n";
    import { RefreshCw, UploadCloud } from "lucide-svelte";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";
    import { goto } from "$app/navigation";

    let loading = false;

    let domains = [];
    let deployLoading = {}; // map of domain -> boolean

    onMount(async () => {
        await loadDomains();
    });

    async function handleResponse(res) {
        if (res.status === 401) {
            toast.error($t("auth.session_expired") || "Session expired");
            goto("/login");
            return null;
        }
        return res;
    }

    async function loadDomains() {
        try {
            let res = await fetch("/api/domains");
            res = await handleResponse(res);
            if (!res) return;

            if (res.ok) {
                const data = await res.json();
                domains = data.domains || [];
            }
        } catch (e) {
            console.error("Failed to load domains", e);
            toast.error("Failed to load domains");
        }
    }

    async function reloadConfig() {
        loading = true;

        try {
            let res = await fetch("/api/system/reload", { method: "POST" });
            res = await handleResponse(res);
            if (!res) return;

            const data = await res.json();
            if (res.ok) {
                toast.success(data.message || $t("system.reload_success"), {
                    duration: 10000,
                });
                await loadDomains(); // Reload domains in case they changed
            } else {
                toast.error(data.error || $t("system.reload_error"), {
                    duration: Infinity,
                });
            }
        } catch (e) {
            toast.error($t("system.reload_error"), {
                duration: Infinity,
            });
        } finally {
            loading = false;
        }
    }

    async function deploy(domain) {
        deployLoading[domain] = true;

        try {
            let res = await fetch("/api/deploy", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ domain }),
            });
            res = await handleResponse(res);
            if (!res) return;

            const data = await res.json();

            if (res.ok) {
                toast.success(data.message, {
                    duration: 10000,
                });
            } else {
                toast.error(data.error || "Deploy failed", {
                    duration: Infinity,
                });
            }
        } catch (e) {
            toast.error("Deploy failed: " + e.message, {
                duration: Infinity,
            });
        } finally {
            deployLoading[domain] = false;
        }
    }

    async function deployAll() {
        for (const domain of domains) {
            await deploy(domain);
        }
    }
</script>

<div class="p-6 space-y-6">
    <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
        {$t("menu.system")}
    </h1>

    <Card.Root>
        <Card.Header>
            <Card.Title>{$t("system.config_title")}</Card.Title>
            <Card.Description>{$t("system.config_desc")}</Card.Description>
        </Card.Header>
        <Card.Content class="space-y-6">
            <div class="flex items-center gap-4">
                <Button on:click={reloadConfig} disabled={loading}>
                    <RefreshCw
                        class="mr-2 h-4 w-4 {loading ? 'animate-spin' : ''}"
                    />
                    {loading
                        ? $t("system.reloading")
                        : $t("system.reload_button")}
                </Button>
            </div>

            {#if domains.length > 0}
                <div class="border-t pt-4">
                    <h3 class="text-lg font-semibold mb-4">
                        {$t("system.deploy_title")}
                    </h3>
                    <div class="flex flex-wrap gap-3">
                        {#each domains as domain}
                            <Button
                                variant="outline"
                                on:click={() => deploy(domain)}
                                disabled={deployLoading[domain]}
                            >
                                <UploadCloud class="mr-2 h-4 w-4" />
                                {deployLoading[domain]
                                    ? $t("system.deploying")
                                    : $t("system.deploy_button", {
                                          values: { domain },
                                      })}
                            </Button>
                        {/each}

                        {#if domains.length > 1}
                            <Button variant="default" on:click={deployAll}>
                                <UploadCloud class="mr-2 h-4 w-4" />
                                {$t("system.deploy_all")}
                            </Button>
                        {/if}
                    </div>
                </div>
            {/if}
        </Card.Content>
    </Card.Root>
</div>
