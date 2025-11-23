<script>
    import { t } from "svelte-i18n";
    import * as Card from "$lib/components/ui/card";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import { toast } from "svelte-sonner";
    import { onMount } from "svelte";
    import PhoneForm from "$lib/components/phones/PhoneForm.svelte";
    import { Search, ChevronLeft, ChevronRight, X } from "lucide-svelte";

    let domains = [];
    let vendors = [];

    let filters = {
        domain: "",
        vendor: "",
        mac: "",
        number: "",
        caller_id: "",
    };

    let phones = [];
    let total = 0;
    let page = 1;
    let limit = 20;
    let loading = false;

    // Edit state
    let editingPhone = null;
    let showEditDialog = false;

    onMount(async () => {
        await Promise.all([loadDomains(), loadVendors()]);
        await search();
    });

    async function loadDomains() {
        try {
            const res = await fetch("/api/domains");
            if (res.ok) {
                const data = await res.json();
                domains = data.domains || [];
            }
        } catch (e) {
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
        } catch (e) {
            console.error("Failed to load vendors", e);
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
        } catch (e) {
            toast.error("Error loading phones: " + e.message);
        } finally {
            loading = false;
        }
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

    function formatMac(mac) {
        if (!mac) return "";
        // Remove any existing separators
        const clean = mac.replace(/[^a-fA-F0-9]/g, "");
        // Add colons every 2 chars
        return clean.match(/.{1,2}/g)?.join(":") || mac;
    }

    function editPhone(phone) {
        console.debug("Start Phone Editor", phone)
        editingPhone = { ...phone }; // Clone
        showEditDialog = true;
    }

    function onSave(e) {
        showEditDialog = false;
        search(page); // Refresh
    }
</script>

<div class="p-6 space-y-6">
    <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
        {$t("menu.phones") || "Phones"}
    </h1>

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
                        class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                        bind:value={filters.domain}
                    >
                        <option value="">{$t("common.all") || "All"}</option>
                        {#each domains as d}
                            <option value={d}>{d}</option>
                        {/each}
                    </select>
                </div>
                <div class="space-y-2">
                    <Label for="s_vendor">{$t("phone.vendor")}</Label>
                    <select
                        id="s_vendor"
                        class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                        bind:value={filters.vendor}
                    >
                        <option value="">{$t("common.all") || "All"}</option>
                        {#each vendors as v}
                            <option value={v.id}>{v.name}</option>
                        {/each}
                    </select>
                </div>
                <div class="space-y-2">
                    <Label for="s_mac">{$t("phone.mac")}</Label>
                    <Input
                        id="s_mac"
                        bind:value={filters.mac}
                        placeholder="00:11:..."
                    />
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
                    <Label for="s_callerid">{$t("phone.caller_id")}</Label>
                    <Input
                        id="s_callerid"
                        bind:value={filters.caller_id}
                        placeholder="John..."
                    />
                </div>
            </div>
            <div class="flex justify-end">
                <Button on:click={() => search(1)} disabled={loading}>
                    <Search class="mr-2 h-4 w-4" />
                    {$t("common.search") || "Search"}
                </Button>
            </div>
        </Card.Content>
    </Card.Root>

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
                                >{$t("phone.caller_id")}</th
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
                        </tr>
                    </thead>
                    <tbody class="[&_tr:last-child]:border-0">
                        {#if phones.length === 0}
                            <tr>
                                <td
                                    colspan="6"
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
                                    <td class="p-4 align-middle"
                                        >{phone.phone_number}</td
                                    >
                                    <td class="p-4 align-middle"
                                        >{phone.caller_id}</td
                                    >
                                    <td class="p-4 align-middle"
                                        >{formatMac(phone.mac_address)}</td
                                    >
                                    <td class="p-4 align-middle"
                                        >{phone.model_id}</td
                                    >
                                    <td class="p-4 align-middle"
                                        >{phone.vendor}</td
                                    >
                                    <td class="p-4 align-middle"
                                        >{phone.domain}</td
                                    >
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
                class="bg-background p-6 rounded-lg shadow-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto border relative"
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
                        mode="edit"
                        phone={editingPhone}
                        on:save={onSave}
                    />
                {/if}
            </div>
        </div>
    {/if}
</div>
