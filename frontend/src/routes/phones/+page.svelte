<script lang="ts">
    import { t } from "svelte-i18n";
    import * as Card from "$lib/components/ui/card";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import { toast } from "svelte-sonner";
    import { onMount } from "svelte";
    import PhoneForm from "$lib/components/phones/PhoneForm.svelte";
    import {
        Search,
        ChevronLeft,
        ChevronRight,
        X,
        Upload,
        Plus,
        Trash2,
    } from "lucide-svelte";
    import type { Phone, Vendor } from "$lib/types";
    import { formatMacInput } from "$lib/utils";

    let domains: string[] = [];
    let vendors: Vendor[] = [];

    let filters: {
        domain: string;
        vendor: string;
        model_id: string;
        mac: string;
        number: string;
        q: string;
    } = {
        domain: "",
        vendor: "",
        model_id: "",
        mac: "",
        number: "",
        q: "",
    };

    let models: any[] = [];

    let phones: Phone[] = [];
    let total = 0;
    let page = 1;
    let limit = 20;
    let loading = false;

    let editingPhone: Phone | null = null;
    let showEditDialog = false;

    $: isFiltered = Object.values(filters).some((v) => v !== "");

    onMount(async () => {
        await Promise.all([loadDomains(), loadVendors()]);
        await search();
    });
    $: if (filters.vendor) {
        loadModels(filters.vendor);
    } else {
        models = [];
        filters.model_id = "";
    }

    async function loadDomains() {
        try {
            const res = await fetch("/api/domains");
            if (res.ok) {
                const data = await res.json();
                domains = data.domains || [];
            }
        } catch (e: any) {
            console.error("Failed to load domains", e);
        }
    }

    async function loadVendors() {
        try {
            const res = await fetch("/api/vendors");
            if (res.ok) {
                const data = await res.json();
                vendors = data.vendors || [];
            }
        } catch (e: any) {
            console.error("Failed to load vendors", e);
        }
    }

    async function loadModels(vendor: string) {
        try {
            const res = await fetch(`/api/models?vendor=${vendor}`);
            if (res.ok) {
                const data = await res.json();
                models = data.models || [];
            }
        } catch (e: any) {
            console.error("Failed to load models", e);
        }
    }
    async function search(newPage = 1) {
        loading = true;
        page = newPage;

        const params = new URLSearchParams({
            page: page.toString(),
            limit: limit.toString(),
            ...filters,
        });

        // Remove empty filters
        for (const [key, value] of Object.entries(filters)) {
            if (!value) params.delete(key);
        }

        try {
            const res = await fetch(`/api/phones?${params.toString()}`);
            if (res.ok) {
                const data = await res.json();
                phones = data.phones || [];
                total = data.total || 0;
            } else {
                toast.error("Failed to load phones");
            }
        } catch (e: any) {
            toast.error("Error loading phones: " + e.message);
        } finally {
            loading = false;
        }
    }

    function clearFilters() {
        filters = {
            domain: "",
            vendor: "",
            model_id: "",
            mac: "",
            number: "",
            q: "",
        };
        search(1);
    }

    function nextPage() {
        if (page * limit < total) {
            search(page + 1);
        }
    }

    function prevPage() {
        if (page > 1) {
            search(page - 1);
        }
    }

    function formatMac(mac: string | null) {
        if (!mac) return "";
        // Remove any existing separators
        const clean = mac.replace(/[^a-fA-F0-9]/g, "");
        // Add colons every 2 chars
        return clean.match(/.{1,2}/g)?.join(":") || mac;
    }

    function editPhone(phone: Phone) {
        console.debug("Start Phone Editor", phone);
        editingPhone = { ...phone }; // Clone
        showEditDialog = true;
    }

    function createPhone() {
        editingPhone = {
            domain: domains[0] || "",
            vendor: "",
            model_id: "",
            mac_address: "",
            phone_number: "",
            ip_address: "",
            description: "",
            lines: [],
            expansion_module_model: "",
            type: "phone",
        };
        showEditDialog = true;
    }

    function onSave(e: CustomEvent) {
        showEditDialog = false;
        search(page); // Refresh
    }

    function getLinesDescription(phone: Phone) {
        const parts = [];
        if (phone.description) {
            parts.push(phone.description);
        }

        if (phone.type === "phone" && phone.lines) {
            const descriptions = [];
            for (const line of phone.lines) {
                if (line.type !== "line") continue;

                let info: Record<string, any> = {};
                if (typeof line.additional_info === "string") {
                    try {
                        info = JSON.parse(line.additional_info);
                    } catch (e: any) {
                        /* ignore */
                    }
                } else if (typeof line.additional_info === "object") {
                    info = line.additional_info || {};
                }
                let text =
                    info.display_name ||
                    info.screen_name ||
                    info.user_name ||
                    info.auth_name ||
                    "";

                if (text) {
                    if (text.length > 20) {
                        text = text.substring(0, 20) + "...";
                    }
                    descriptions.push(text);
                }
            }
            if (descriptions.length > 0) {
                parts.push(descriptions.join(", "));
            }
        }

        return parts.join(": ");
    }

    async function deletePhone(phone: Phone) {
        if (
            !confirm(
                $t("phone.confirm_delete") ||
                    "Are you sure you want to delete this phone?",
            )
        ) {
            return;
        }
        try {
            const res = await fetch(`/api/phones/${phone.id}`, {
                method: "DELETE",
            });
            if (res.ok) {
                toast.success($t("phone.deleted") || "Phone deleted");
                search(page);
            } else {
                toast.error(
                    $t("phone.delete_failed") || "Failed to delete phone",
                );
            }
        } catch (e: any) {
            toast.error("Error deleting phone: " + e.message);
        }
    }
</script>

<div class="p-6 space-y-6">
    <div class="flex justify-between items-center">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
            {$t("menu.phones") || "Phones"}
        </h1>
    </div>

    <Card.Root>
        <Card.Header>
            <Card.Title
                >{$t("phone.search_title") || "Search Phones"}</Card.Title
            >
        </Card.Header>
        <Card.Content class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-5 gap-4">
                <div class="space-y-2">
                    <Label for="s_domain">{$t("phone.domain")}</Label>
                    <select
                        id="s_domain"
                        bind:value={filters.domain}
                        class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                        <option value=""
                            >{$t("common.all_domains") || "Все"}</option
                        >
                        {#each domains as d}
                            <option value={d}>{d}</option>
                        {/each}
                    </select>
                </div>
                <div class="space-y-2 md:col-span-1">
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                        <div class="space-y-2">
                            <Label for="s_vendor">{$t("phone.vendor")}</Label>
                            <select
                                id="s_vendor"
                                bind:value={filters.vendor}
                                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                            >
                                <option value=""
                                    >{$t("common.all_vendors") || "Все"}</option
                                >
                                {#each vendors as v}
                                    <option value={v.id}>{v.name}</option>
                                {/each}
                            </select>
                        </div>
                        {#if filters.vendor}
                            <div class="space-y-2">
                                <Label for="s_model"
                                    >{$t("phone.model") || "Модель"}</Label
                                >
                                <select
                                    id="s_model"
                                    bind:value={filters.model_id}
                                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                >
                                    <option value=""
                                        >{$t("common.all_models") ||
                                            "Все"}</option
                                    >
                                    {#each models as m}
                                        <option value={m.id}>{m.name}</option>
                                    {/each}
                                </select>
                            </div>
                        {/if}
                    </div>
                </div>
                <div class="space-y-2">
                    <Label for="s_mac">{$t("phone.mac")}</Label>
                    <div class="relative group">
                        <Input
                            id="s_mac"
                            class="pr-8"
                            value={filters.mac}
                            on:input={(e) => {
                                filters.mac = formatMacInput(
                                    e.currentTarget.value,
                                );
                                e.currentTarget.value = filters.mac;
                            }}
                            placeholder="00:11:..."
                        />
                        {#if filters.mac}
                            <button
                                class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                                on:click={() => {
                                    filters.mac = "";
                                    search(1);
                                }}
                            >
                                <X class="h-4 w-4" />
                            </button>
                        {/if}
                    </div>
                </div>
                <div class="space-y-2">
                    <Label for="s_number">{$t("phone.number")}</Label>
                    <Input
                        id="s_number"
                        bind:value={filters.number}
                        placeholder="101"
                    />
                </div>
                <div class="space-y-2">
                    <Label for="s_q">Сквозной поиск</Label>
                    <div class="relative group">
                        <Input
                            id="s_q"
                            class="pr-8"
                            bind:value={filters.q}
                            placeholder="Поиск по номеру, описанию..."
                        />
                        {#if filters.q}
                            <button
                                class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                                on:click={() => {
                                    filters.q = "";
                                    search(1);
                                }}
                            >
                                <X class="h-4 w-4" />
                            </button>
                        {/if}
                    </div>
                </div>
            </div>
            <div class="flex justify-end gap-2">
                {#if isFiltered}
                    <Button
                        variant="outline"
                        on:click={clearFilters}
                        disabled={loading}
                    >
                        <X class="mr-2 h-4 w-4" />
                        Сбросить фильтр
                    </Button>
                {/if}
                <Button on:click={() => search(1)} disabled={loading}>
                    <Search class="mr-2 h-4 w-4" />
                    {$t("common.search") || "Search"}
                </Button>
            </div>
        </Card.Content>
    </Card.Root>

    <div class="flex justify-between items-center">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
            {$t("menu.phones") || "Phones"}
        </h1>
        <div class="flex gap-2">
            <Button variant="outline" href="/phones/import">
                <Upload class="mr-2 h-4 w-4" />
                {$t("phones.import") || "Import"}
            </Button>
            <Button on:click={createPhone}>
                <Plus class="mr-2 h-4 w-4" />
                {$t("phones.create") || "Create Phone"}
            </Button>
        </div>
    </div>

    <Card.Root>
        <Card.Content class="p-0">
            <div class="relative w-full overflow-auto">
                <table class="w-full caption-bottom text-sm">
                    <thead class="[&_tr]:border-b">
                        <tr
                            class="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted"
                        >
                            <th
                                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
                                >{$t("phone.number")}</th
                            >
                            <th
                                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
                                >{$t("common.description") || "Description"}</th
                            >
                            <th
                                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
                                >{$t("phone.mac")}</th
                            >
                            <th
                                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
                                >{$t("phone.model") || "Model"}</th
                            >
                            <th
                                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
                                >{$t("phone.vendor")}</th
                            >
                            <th
                                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
                                >{$t("phone.domain")}</th
                            >
                            <th
                                class="h-12 px-2 text-right align-middle font-medium text-muted-foreground w-[1%] whitespace-nowrap"
                            >
                                {$t("common.actions") || "Actions"}
                            </th>
                        </tr>
                    </thead>
                    <tbody class="[&_tr:last-child]:border-0">
                        {#if phones.length === 0}
                            <tr>
                                <td
                                    colspan="7"
                                    class="p-4 text-center text-muted-foreground"
                                >
                                    {$t("common.no_results") ||
                                        "No results found"}
                                </td>
                            </tr>
                        {:else}
                            {#each phones as phone}
                                <tr
                                    class="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted cursor-pointer"
                                    on:click={() => editPhone(phone)}
                                >
                                    <td class="p-4 align-middle">
                                        {#if phone.type === "gateway"}
                                            {phone.ip_address}
                                        {:else}
                                            {phone.phone_number}
                                        {/if}
                                    </td>
                                    <td class="p-4 align-middle">
                                        {getLinesDescription(phone)}
                                    </td>
                                    <td class="p-4 align-middle"
                                        >{formatMac(phone.mac_address)}</td
                                    >
                                    <td class="p-4 align-middle"
                                        >{phone.model_name ||
                                            phone.model_id}</td
                                    >
                                    <td class="p-4 align-middle"
                                        >{phone.vendor_name || phone.vendor}</td
                                    >
                                    <td class="p-4 align-middle"
                                        >{phone.domain}</td
                                    >
                                    <td class="p-2 align-middle text-right">
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            class="text-destructive hover:text-destructive hover:bg-destructive/10"
                                            on:click={(e) => {
                                                e.stopPropagation();
                                                deletePhone(phone);
                                            }}
                                        >
                                            <Trash2 class="h-4 w-4" />
                                        </Button>
                                    </td>
                                </tr>
                            {/each}
                        {/if}
                    </tbody>
                </table>
            </div>

            <div class="flex items-center justify-end space-x-2 py-4 px-4">
                <div class="flex-1 text-sm text-muted-foreground">
                    {$t("common.total") || "Total"}: {total}
                </div>
                <div class="space-x-2">
                    <Button
                        variant="outline"
                        size="sm"
                        on:click={prevPage}
                        disabled={page <= 1 || loading}
                    >
                        <ChevronLeft class="h-4 w-4" />
                        {$t("common.prev") || "Previous"}
                    </Button>
                    <Button
                        variant="outline"
                        size="sm"
                        on:click={nextPage}
                        disabled={page * limit >= total || loading}
                    >
                        {$t("common.next") || "Next"}
                        <ChevronRight class="h-4 w-4" />
                    </Button>
                </div>
            </div>
        </Card.Content>
    </Card.Root>

    {#if showEditDialog}
        <div
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        >
            <div
                class="bg-background dark:bg-slate-900 p-6 rounded-lg shadow-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto border dark:border-slate-700 relative"
            >
                <Button
                    variant="ghost"
                    size="icon"
                    class="absolute right-4 top-4"
                    on:click={() => (showEditDialog = false)}
                >
                    <X class="h-4 w-4" />
                </Button>
                {#if editingPhone}
                    <PhoneForm
                        mode={editingPhone.id ? "edit" : "create"}
                        phone={editingPhone}
                        on:save={onSave}
                        on:cancel={() => (showEditDialog = false)}
                    />
                {/if}
            </div>
        </div>
    {/if}
</div>
