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
        Archive,
        Trash2,
    } from "lucide-svelte";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";
    import { goto } from "$app/navigation";
    import {
        Alert,
        AlertDescription,
        AlertTitle,
    } from "$lib/components/ui/alert";
    import {
        licenseInfo,
        isPro,
        fetchLicenseStatus,
    } from "$lib/stores/licenseStore";
    import { Shield, LifeBuoy } from "lucide-svelte";

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
        await fetchLicenseStatus();
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
            toast.error($t("system.load_domains_error"));
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

    async function createBackup(type) {
        backupLoading = true;
        try {
            const url =
                type === "db"
                    ? "/api/system/backups/create/db"
                    : "/api/system/backups/create/cfg";
            let res = await fetch(url, { method: "POST" });
            res = await handleResponse(res);
            if (!res) return;
            if (res.ok) {
                toast.success(
                    type === "db"
                        ? $t("backup.create_db_success")
                        : $t("backup.create_cfg_success"),
                );
                await loadBackups();
            } else {
                const data = await res.json();
                toast.error(data.error || $t("system.reload_error"));
            }
        } catch (e) {
            toast.error($t("system.reload_error") + ": " + e.message);
        } finally {
            backupLoading = false;
        }
    }

    async function downloadBackup(filename) {
        window.location.href = `/api/system/backups/download/${filename}`;
    }

    async function uploadBackup(event) {
        const file = event.target.files[0];
        if (!file) return;

        const formData = new FormData();
        formData.append("backup", file);

        backupLoading = true;
        try {
            let res = await fetch("/api/system/backups/upload", {
                method: "POST",
                body: formData,
            });
            res = await handleResponse(res);
            if (!res) return;

            if (res.ok) {
                toast.success($t("backup.upload_success"));
                await loadBackups();
            } else {
                const data = await res.json();
                toast.error(data.error || $t("system.deploy_error"));
            }
        } catch (e) {
            toast.error($t("system.deploy_error") + ": " + e.message);
        } finally {
            backupLoading = false;
            event.target.value = ""; // Reset file input
        }
    }

    async function restoreBackup(backupInfo) {
        const isDB = backupInfo.type === "db";
        const typeLabel = isDB ? $t("backup.db_type") : $t("backup.cfg_type");

        let confirmMsg = $t("backup.restore_confirm", {
            values: {
                type: typeLabel,
                name: backupInfo.name,
                target: isDB ? $t("common.data") : $t("phone.settings"),
            },
        });
        if (!confirm(confirmMsg)) return;

        restoreLoading = true;
        try {
            const url = isDB
                ? "/api/system/backups/restore/db"
                : "/api/system/backups/restore/cfg";
            let res = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ filename: backupInfo.name }),
            });
            res = await handleResponse(res);
            if (!res) return;
            if (res.ok) {
                if (isDB) {
                    toast.success($t("backup.restore_db_success"));
                } else {
                    toast.success($t("backup.restore_cfg_success"), {
                        description: $t("backup.restore_cfg_instruction"),
                        duration: 10000,
                    });
                }
                await loadBackups();
            } else {
                const data = await res.json();
                toast.error(data.error || $t("system.reload_error"));
            }
        } catch (e) {
            toast.error($t("system.reload_error") + ": " + e.message);
        } finally {
            restoreLoading = false;
        }
    }

    async function deleteBackup(filename) {
        if (
            !confirm(
                $t("backup.delete_confirm") || `Delete backup ${filename}?`,
            )
        )
            return;

        backupLoading = true;
        try {
            let res = await fetch(`/api/system/backups/${filename}`, {
                method: "DELETE",
            });
            res = await handleResponse(res);
            if (!res) return;

            if (res.ok) {
                toast.success($t("backup.delete_success") || "Backup deleted");
                await loadBackups();
            } else {
                const data = await res.json();
                toast.error(data.error || "Failed to delete backup");
            }
        } catch (e) {
            toast.error("Error: " + e.message);
        } finally {
            backupLoading = false;
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
                    toast.warning($t("system.config_prepared_warnings"), {
                        duration: 5000,
                    });
                } else {
                    toast.success(
                        data.message || $t("system.config_prepared_success"),
                        {
                            duration: 5000,
                        },
                    );
                    configPrepared = true;
                }
                await loadDomains();
            } else {
                toast.error(data.error || $t("system.reload_error"), {
                    duration: Infinity,
                });
            }
        } catch (e) {
            toast.error($t("system.prepare_error") + ": " + e.message, {
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
                toast.success(
                    data.message || $t("system.config_apply_success"),
                    {
                        duration: 5000,
                    },
                );
                configPrepared = false; // Reset state
            } else {
                toast.error(data.error || $t("system.reload_error"), {
                    duration: Infinity,
                });
            }
        } catch (e) {
            toast.error($t("system.apply_error") + ": " + e.message, {
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
                toast.error(data.error || $t("system.deploy_error"), {
                    duration: Infinity,
                });
            }
        } catch (e) {
            toast.error($t("system.deploy_error") + ": " + e.message, {
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

    async function uploadLicense(event) {
        /** @type {HTMLInputElement} */
        const target = event.target;
        const file = target?.files?.[0];
        if (!file) return;

        const formData = new FormData();
        formData.append("license", file);

        try {
            let res = await fetch("/api/system/license/upload", {
                method: "POST",
                body: formData,
            });
            res = await handleResponse(res);
            if (!res) return;

            if (res.ok) {
                toast.success($t("system.license_uploaded"));
                await fetchLicenseStatus();
            } else {
                const data = await res.json();
                toast.error(data.error || $t("system.upload_error"));
            }
        } catch (e) {
            toast.error(
                $t("system.upload_error") +
                    ": " +
                    (e instanceof Error ? e.message : String(e)),
            );
        } finally {
            event.target.value = ""; // Reset
        }
    }

    async function generateSupportBundle() {
        try {
            let res = await fetch("/api/system/support/bundle", {
                method: "POST",
            });
            res = await handleResponse(res);
            if (!res) return;

            if (res.ok) {
                toast.success($t("system.support_bundle_ready"));
                await loadBackups();
            } else {
                const data = await res.json();
                toast.error(data.error || $t("system.reload_error"));
            }
        } catch (e) {
            toast.error(
                $t("system.reload_error") +
                    ": " +
                    (e instanceof Error ? e.message : String(e)),
            );
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
                    {loading
                        ? $t("system.preparing")
                        : $t("system.reload_button")}
                </Button>

                <Button
                    variant="default"
                    on:click={applyConfig}
                    disabled={loading || applying || !configPrepared}
                >
                    <Check class="mr-2 h-4 w-4" />
                    {applying
                        ? $t("system.applying")
                        : $t("system.apply_button")}
                </Button>
            </div>

            {#if warnings.length > 0}
                <Alert variant="destructive">
                    <AlertTriangle class="h-4 w-4" />
                    <div class="flex justify-between items-start w-full">
                        <div>
                            <AlertTitle
                                >{$t("system.warnings_title")}</AlertTitle
                            >
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
                            {$t("system.ignore_button")}
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
            <Card.Title class="flex items-center gap-2">
                <Shield class="w-5 h-5 text-primary" />
                {$t("system.license")}
            </Card.Title>
            <Card.Description>{$t("system.license_desc")}</Card.Description>
        </Card.Header>
        <Card.Content class="space-y-6">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="space-y-1.5">
                    <div
                        class="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-900/50 rounded-lg"
                    >
                        <span class="text-xs opacity-60"
                            >{$t("system.license_tier")}:</span
                        >
                        <span
                            class="font-bold border px-2 py-0.5 text-xs rounded-full bg-white dark:bg-slate-800"
                            class:text-green-600={$isPro}
                            class:border-green-600={$isPro}
                        >
                            {#if $licenseInfo.tier === "Free" && $licenseInfo.license_key}
                                {$t("system.key_uploaded") || "Key Uploaded"}
                            {:else}
                                {$licenseInfo.tier}
                            {/if}
                        </span>
                    </div>

                    {#if $licenseInfo.issued_to}
                        <div
                            class="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-900/50 rounded-lg"
                        >
                            <span class="text-xs opacity-60"
                                >{$t("system.license_issued_to")}:</span
                            >
                            <span class="font-medium text-sm"
                                >{$licenseInfo.issued_to}</span
                            >
                        </div>
                    {/if}

                    <div
                        class="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-900/50 rounded-lg"
                    >
                        <span class="text-xs opacity-60"
                            >{$t("system.support_level") ||
                                "Support Level"}:</span
                        >
                        <span class="font-medium text-sm"
                            >{$licenseInfo.support_level}</span
                        >
                    </div>

                    {#if $licenseInfo.valid_from}
                        <div
                            class="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-900/50 rounded-lg"
                        >
                            <span class="text-xs opacity-60"
                                >{$t("system.license_valid_from")}:</span
                            >
                            <span class="font-mono text-xs"
                                >{new Date(
                                    $licenseInfo.valid_from,
                                ).toLocaleDateString()}</span
                            >
                        </div>
                    {/if}

                    {#if $licenseInfo.expiry}
                        <div
                            class="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-900/50 rounded-lg"
                        >
                            <span class="text-xs opacity-60"
                                >{$t("system.license_expiry")}:</span
                            >
                            <span
                                class="font-mono text-xs"
                                class:text-destructive={new Date(
                                    $licenseInfo.expiry,
                                ) < new Date()}
                            >
                                {new Date(
                                    $licenseInfo.expiry,
                                ).toLocaleDateString()}
                            </span>
                        </div>
                    {/if}

                    {#if $licenseInfo.license_key}
                        <div
                            class="flex flex-col gap-0.5 p-2 bg-slate-50 dark:bg-slate-900/50 rounded-lg"
                        >
                            <span
                                class="text-[10px] opacity-40 uppercase tracking-widest"
                                >{$t("system.license_key")}:</span
                            >
                            <span
                                class="font-mono text-[10px] break-all opacity-60"
                                >{$licenseInfo.license_key}</span
                            >
                        </div>
                    {/if}

                    <div class="flex flex-col gap-2">
                        <Button
                            variant={$isPro ? "default" : "outline"}
                            class="w-full"
                            on:click={() => {
                                if ($isPro) {
                                    toast.info(
                                        $t("system.opening_support_ticket"),
                                    );
                                    // window.open('https://support.example.com', '_blank');
                                } else {
                                    toast.warning(
                                        $t("system.pro_required_for_support"),
                                    );
                                }
                            }}
                        >
                            <LifeBuoy class="mr-2 h-4 w-4" />
                            {$t("system.request_support")}
                        </Button>
                        <Button
                            variant="ghost"
                            size="sm"
                            class="opacity-70"
                            on:click={generateSupportBundle}
                        >
                            <Archive class="mr-2 h-4 w-4" />
                            {$t("system.generate_diagnostic_bundle")}
                        </Button>
                    </div>
                </div>

                <div class="space-y-4">
                    <div
                        class="p-4 border border-dashed rounded-lg flex flex-col items-center justify-center gap-3"
                    >
                        <Shield class="w-8 h-8 opacity-20" />
                        <div class="text-center">
                            <p class="text-sm font-semibold">
                                {$t("system.upload_license_key")}
                            </p>
                            <p class="text-xs opacity-60">
                                {$t("system.license_upload_hint")}
                            </p>
                        </div>
                        <input
                            type="file"
                            id="license-upload"
                            class="hidden"
                            on:change={uploadLicense}
                        />
                        <Button
                            variant="outline"
                            size="sm"
                            on:click={() =>
                                document
                                    .getElementById("license-upload")
                                    ?.click()}
                        >
                            <UploadCloud class="mr-2 h-4 w-4" />
                            {$t("common.upload")}
                        </Button>
                    </div>
                </div>
            </div>
        </Card.Content>
    </Card.Root>

    <Card.Root>
        <Card.Header>
            <Card.Title>{$t("backup.create")}</Card.Title>
            <Card.Description>{$t("backup.description")}.</Card.Description>
        </Card.Header>
        <Card.Content class="space-y-6">
            <div class="flex flex-wrap items-center gap-4">
                <Button
                    on:click={() => createBackup("db")}
                    disabled={backupLoading || restoreLoading}
                >
                    <Database class="mr-2 h-4 w-4" />
                    {backupLoading
                        ? $t("common.creating")
                        : $t("backup.create_db")}
                </Button>

                <Button
                    on:click={() => createBackup("cfg")}
                    disabled={backupLoading || restoreLoading}
                >
                    <Archive class="mr-2 h-4 w-4" />
                    {backupLoading
                        ? $t("common.creating")
                        : $t("backup.create_cfg")}
                </Button>

                <div class="flex-1"></div>

                <div class="relative">
                    <input
                        type="file"
                        id="backup-upload"
                        class="hidden"
                        accept=".zip"
                        on:change={uploadBackup}
                    />
                    <Button
                        variant="ghost"
                        on:click={() =>
                            document.getElementById("backup-upload").click()}
                        disabled={backupLoading || restoreLoading}
                    >
                        <UploadCloud class="mr-2 h-4 w-4" />
                        {$t("backup.upload")}
                    </Button>
                </div>
            </div>

            <div class="border rounded-md">
                <Table.Root>
                    <Table.Header>
                        <Table.Row>
                            <Table.Head>{$t("common.name")}</Table.Head>
                            <Table.Head>{$t("backup.type")}</Table.Head>
                            <Table.Head>{$t("common.date")}</Table.Head>
                            <Table.Head>{$t("common.size")}</Table.Head>
                            <Table.Head class="text-right"
                                >{$t("common.actions")}</Table.Head
                            >
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {#each backups as backup}
                            <Table.Row>
                                <Table.Cell class="font-mono text-xs"
                                    >{backup.name}</Table.Cell
                                >
                                <Table.Cell>
                                    {#if backup.type === "db"}
                                        <span
                                            class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                                        >
                                            {$t("backup.db_type")}
                                        </span>
                                    {:else if backup.type === "config"}
                                        <span
                                            class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200"
                                        >
                                            {$t("backup.cfg_type")}
                                        </span>
                                    {:else if backup.type === "support"}
                                        <span
                                            class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200"
                                        >
                                            {$t("backup.support_type")}
                                        </span>
                                    {/if}
                                </Table.Cell>
                                <Table.Cell class="whitespace-nowrap"
                                    >{new Date(
                                        backup.time,
                                    ).toLocaleString()}</Table.Cell
                                >
                                <Table.Cell
                                    >{(backup.size / 1024 / 1024).toFixed(2)} MB</Table.Cell
                                >
                                <Table.Cell
                                    class="text-right flex justify-end gap-2"
                                >
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        on:click={() =>
                                            downloadBackup(backup.name)}
                                    >
                                        <Download class="h-4 w-4" />
                                    </Button>

                                    <Button
                                        variant="outline"
                                        size="sm"
                                        class="text-destructive hover:text-destructive hover:bg-destructive/10"
                                        on:click={() =>
                                            deleteBackup(backup.name)}
                                        disabled={backupLoading}
                                    >
                                        <Trash2 class="h-4 w-4" />
                                    </Button>

                                    {#if backup.type !== "support"}
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            on:click={() =>
                                                restoreBackup(backup)}
                                            disabled={backupLoading ||
                                                restoreLoading}
                                        >
                                            <RotateCcw class="mr-2 h-4 w-4" />
                                            {$t("common.restore")}
                                        </Button>
                                    {/if}
                                </Table.Cell>
                            </Table.Row>
                        {/each}
                        {#if backups.length === 0}
                            <Table.Row>
                                <Table.Cell
                                    colspan={5}
                                    class="text-center text-muted-foreground py-8"
                                >
                                    {$t("common.no_results")}
                                </Table.Cell>
                            </Table.Row>
                        {/if}
                    </Table.Body>
                </Table.Root>
            </div>
        </Card.Content>
    </Card.Root>
</div>
