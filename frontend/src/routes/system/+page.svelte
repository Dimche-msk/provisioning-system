<script>
    import { Button } from "$lib/components/ui/button";
    import * as Card from "$lib/components/ui/card";
    import { t } from "svelte-i18n";
    import * as Table from "$lib/components/ui/table";
    import {
        Check,
        RefreshCw,
        UploadCloud,
        AlertTriangle,
        Database,
        RotateCcw,
        Download,
    } from "lucide-svelte";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";
    import { goto } from "$app/navigation";
    import {
        Alert,
        AlertDescription,
        AlertTitle,
    } from "$lib/components/ui/alert";

    let loading = false;
    let applying = false;
    let configPrepared = false;
    let warnings = [];

    let domains = [];
    let deployLoading = {}; // map of domain -> boolean

    let backups = [];
    let backupLoading = false;
    let restoreLoading = false;

    onMount(async () => {
        await loadDomains();
        await loadBackups();
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

    async function loadBackups() {
        try {
            let res = await fetch("/api/system/backups");
            res = await handleResponse(res);
            if (!res) return;
            if (res.ok) {
                const data = await res.json();
                backups = data.backups || [];
            }
        } catch (e) {
            console.error("Failed to load backups", e);
        }
    }

    async function createBackup() {
        backupLoading = true;
        try {
            let res = await fetch("/api/system/backup", { method: "POST" });
            res = await handleResponse(res);
            if (!res) return;
            if (res.ok) {
                toast.success("Backup created successfully");
                await loadBackups();
            } else {
                const data = await res.json();
                toast.error(data.error || "Failed to create backup");
            }
        } catch (e) {
            toast.error("Error creating backup: " + e.message);
        } finally {
            backupLoading = false;
        }
    }

    async function restoreBackup(filename) {
        if (
            !confirm(
                `Are you sure you want to restore from ${filename}? Current data will be lost.`,
            )
        )
            return;

        restoreLoading = true;
        try {
            let res = await fetch("/api/system/restore", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ filename }),
            });
            res = await handleResponse(res);
            if (!res) return;
            if (res.ok) {
                toast.success("Backup restored successfully");
                await loadBackups();
            } else {
                const data = await res.json();
                toast.error(data.error || "Failed to restore backup");
            }
        } catch (e) {
            toast.error("Error restoring backup: " + e.message);
        } finally {
            restoreLoading = false;
        }
    }

    async function prepareConfig() {
        loading = true;
        configPrepared = false;
        warnings = [];

        try {
            let res = await fetch("/api/system/reload", { method: "POST" });
            res = await handleResponse(res);
            if (!res) return;

            const data = await res.json();
            if (res.ok) {
                if (data.warnings && data.warnings.length > 0) {
                    warnings = data.warnings;
                    configPrepared = false; // Require explicit ignore
                    toast.warning("Configuration prepared with warnings", {
                        duration: 5000,
                    });
                } else {
                    toast.success(data.message || "Configuration prepared", {
                        duration: 5000,
                    });
                    configPrepared = true;
                }
                await loadDomains();
            } else {
                toast.error(data.error || "Failed to prepare configuration", {
                    duration: Infinity,
                });
            }
        } catch (e) {
            toast.error("Error preparing configuration: " + e.message, {
                duration: Infinity,
            });
        } finally {
            loading = false;
        }
    }

    function ignoreWarnings() {
        warnings = [];
        configPrepared = true;
    }

    async function applyConfig() {
        applying = true;

        try {
            let res = await fetch("/api/system/apply", { method: "POST" });
            res = await handleResponse(res);
            if (!res) return;

            const data = await res.json();
            if (res.ok) {
                toast.success(data.message || "Configuration applied", {
                    duration: 5000,
                });
                configPrepared = false; // Reset state
            } else {
                toast.error(data.error || "Failed to apply configuration", {
                    duration: Infinity,
                });
            }
        } catch (e) {
            toast.error("Error applying configuration: " + e.message, {
                duration: Infinity,
            });
        } finally {
            applying = false;
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
                <Button on:click={prepareConfig} disabled={loading || applying}>
                    <RefreshCw
                        class="mr-2 h-4 w-4 {loading ? 'animate-spin' : ''}"
                    />
                    {loading ? "Preparing..." : $t("system.reload_button")}
                </Button>

                <Button
                    variant="default"
                    on:click={applyConfig}
                    disabled={loading || applying || !configPrepared}
                >
                    <Check class="mr-2 h-4 w-4" />
                    {applying ? "Applying..." : $t("system.apply_button")}
                </Button>
            </div>

            {#if warnings.length > 0}
                <Alert variant="destructive">
                    <AlertTriangle class="h-4 w-4" />
                    <div class="flex justify-between items-start w-full">
                        <div>
                            <AlertTitle>Warnings</AlertTitle>
                            <AlertDescription>
                                <ul
                                    class="list-disc pl-4 text-sm space-y-1 mt-2"
                                >
                                    {#each warnings as warning}
                                        <li>{warning}</li>
                                    {/each}
                                </ul>
                            </AlertDescription>
                        </div>
                        <Button
                            variant="outline"
                            size="sm"
                            class="ml-4 bg-white text-destructive hover:bg-gray-100 border-destructive/20"
                            on:click={ignoreWarnings}
                        >
                            Ignore
                        </Button>
                    </div>
                </Alert>
            {/if}

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

    <Card.Root>
        <Card.Header>
            <Card.Title>Backup & Restore</Card.Title>
            <Card.Description
                >{$t("backup.description")}.</Card.Description
            >
        </Card.Header>
        <Card.Content class="space-y-6">
            <div class="flex items-center gap-4">
                <Button
                    on:click={createBackup}
                    disabled={backupLoading || restoreLoading}
                >
                    <Database class="mr-2 h-4 w-4" />
                    {backupLoading ? "Creating Backup..." : $t("backup.create") }
                </Button>
            </div>

            <div class="border rounded-md">
                <Table.Root>
                    <Table.Header>
                        <Table.Row>
                            <Table.Head>{$t("common.name")}</Table.Head>
                            <Table.Head>{$t("common.date")}</Table.Head>
                            <Table.Head>{$t("common.size")}</Table.Head>
                            <Table.Head class="text-right">{$t("common.actions")}</Table.Head>
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {#each backups as backup}
                            <Table.Row>
                                <Table.Cell>{backup.name}</Table.Cell>
                                <Table.Cell
                                    >{new Date(
                                        backup.time,
                                    ).toLocaleString()}</Table.Cell
                                >
                                <Table.Cell
                                    >{(backup.size / 1024 / 1024).toFixed(2)} MB</Table.Cell
                                >
                                <Table.Cell class="text-right">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        on:click={() =>
                                            restoreBackup(backup.name)}
                                        disabled={restoreLoading}
                                    >
                                        <RotateCcw class="mr-2 h-4 w-4" />
                                        Restore
                                    </Button>
                                </Table.Cell>
                            </Table.Row>
                        {/each}
                        {#if backups.length === 0}
                            <Table.Row>
                                <Table.Cell
                                    colspan={4}
                                    class="text-center text-muted-foreground"
                                >
                                    No backups found
                                </Table.Cell>
                            </Table.Row>
                        {/if}
                    </Table.Body>
                </Table.Root>
            </div>
        </Card.Content>
    </Card.Root>
</div>
