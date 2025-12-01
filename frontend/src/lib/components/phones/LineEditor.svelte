<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { t } from "svelte-i18n";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Table from "$lib/components/ui/table";
    import {
        Pencil,
        Trash2,
        Plus,
        Search,
        Save,
        X,
        Check,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import type { Phone, PhoneLine } from "$lib/types";

    export let lines: PhoneLine[] = [];
    export let maxSoftKeys = 0;
    export let maxHardKeys = 0;

    export let image = "";

    export let maxLines = 0;
    export let open = false;
    export let phone: Phone = {} as Phone;

    const dispatch = createEventDispatcher();

    let workingLines: PhoneLine[] = [];

    // Reset working copy when opening
    $: if (open) {
        // Deep copy to avoid modifying parent state by reference
        workingLines = JSON.parse(JSON.stringify(lines));
    }

    let searchQuery = "";
    let currentPage = 1;
    let itemsPerPage = 16;

    // Editing state
    let editForm: PhoneLine | null = null;
    let additionalInfo: Record<string, any> = {}; // Parsed JSON

    // Filtered lines
    $: filteredLines = workingLines.filter((l) => {
        const q = searchQuery.toLowerCase();
        // Parse info for search
        let info: Record<string, any> = {};
        try {
            info = JSON.parse(l.additional_info || "{}");
        } catch (e) {}

        const searchStr = [
            l.number,
            l.type,
            info.display_name,
            info.user_name,
            info.label,
            info.value,
        ]
            .join(" ")
            .toLowerCase();

        return searchStr.includes(q);
    });

    $: totalPages = Math.ceil(filteredLines.length / itemsPerPage);
    $: paginatedLines = filteredLines.slice(
        (currentPage - 1) * itemsPerPage,
        currentPage * itemsPerPage,
    );

    $: hasExpansionModules = (phone.expansion_modules_count || 0) > 0;

    let originalLine: PhoneLine | null = null;

    $: imageUrl =
        image && phone.vendor
            ? `/api/vendors-static/${phone.vendor}/static/${image}`
            : "";

    let imageLoadError = false;
    $: if (imageUrl) imageLoadError = false;

    function edit(line: PhoneLine) {
        originalLine = line;
        editForm = { ...line };
        try {
            additionalInfo = JSON.parse(line.additional_info || "{}");
        } catch (e) {
            additionalInfo = {};
        }
    }

    function add() {
        // Calculate next number
        const maxNum = workingLines.reduce(
            (max, l) => (l.number > max ? l.number : max),
            0,
        );

        originalLine = null;
        editForm = {
            type: "line",
            number: maxNum + 1,
            expansion_module_number: 0,
            key_number: 0,
            additional_info: "{}",
        };
        additionalInfo = {};
    }

    function save() {
        if (!editForm || !editForm.number) {
            toast.error("Number is required");
            return;
        }

        // Ensure numbers are integers
        const newNumber = parseInt(String(editForm.number), 10);
        const newExpModule = editForm.expansion_module_number
            ? parseInt(String(editForm.expansion_module_number), 10)
            : 0;
        const newKeyNumber = editForm.key_number
            ? parseInt(String(editForm.key_number), 10)
            : 0;

        editForm.number = newNumber;
        editForm.expansion_module_number = newExpModule;
        editForm.key_number = newKeyNumber;

        // Check Limits
        if (editForm.type === "soft_key") {
            const count = workingLines.filter(
                (l) => l.type === "soft_key" && l !== originalLine,
            ).length;
            if (count >= maxSoftKeys) {
                toast.error(
                    `Maximum number of soft keys (${maxSoftKeys}) reached.`,
                );
                return;
            }
        }
        if (editForm.type === "hard_key") {
            const count = workingLines.filter(
                (l) => l.type === "hard_key" && l !== originalLine,
            ).length;
            if (count >= maxHardKeys) {
                toast.error(
                    `Maximum number of hard keys (${maxHardKeys}) reached.`,
                );
                return;
            }
        }

        // Validation: Check for duplicates
        for (const line of workingLines) {
            // Skip if we are editing this exact line
            if (originalLine && line === originalLine) continue;

            // Check 1: Unique Sequential Number (Per Type)
            if (line.number === newNumber && line.type === editForm.type) {
                toast.error(
                    `Key number ${newNumber} is already in use for type ${editForm.type}.`,
                );
                return;
            }

            // Check 2: Unique Physical Location (Exp Module + Key)
            // Only if Key Number is set (assuming 0 means undefined/auto)
            if (newKeyNumber > 0) {
                const lineExp = line.expansion_module_number || 0;
                const lineKey = line.key_number || 0;

                if (lineExp === newExpModule && lineKey === newKeyNumber) {
                    const locName =
                        newExpModule === 0
                            ? "Main Phone"
                            : `Exp Module ${newExpModule}`;
                    toast.error(
                        `Key ${newKeyNumber} on ${locName} is already assigned.`,
                    );
                    return;
                }
            }
        }

        // Pack additional info
        editForm.additional_info = JSON.stringify(additionalInfo);

        if (originalLine) {
            // Update
            const idx = workingLines.indexOf(originalLine);
            if (idx !== -1) {
                workingLines[idx] = { ...editForm };
            }
        } else {
            // Create
            workingLines = [...workingLines, { ...editForm }];
        }
        originalLine = null;
        editForm = null;
        additionalInfo = {};
    }

    function remove(line: PhoneLine) {
        workingLines = workingLines.filter((l) => l !== line);
    }

    function cancelEdit() {
        originalLine = null;
        editForm = null;
        additionalInfo = {};
    }

    function close() {
        dispatch("close");
    }

    function saveAll() {
        dispatch("save", workingLines);
        close();
    }

    function getLineDescription(line: PhoneLine) {
        let info: Record<string, any> = {};
        try {
            info = JSON.parse(line.additional_info || "{}");
        } catch (e) {}

        if (line.type === "line") {
            return info.display_name || info.user_name || info.auth_name || "";
        } else {
            return info.label || info.value || "";
        }
    }

    let vendors: any[] = [];
    let currentVendorFeatures: any[] = [];

    onMount(async () => {
        await loadVendors();
    });

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

    $: if (phone && vendors.length > 0) {
        const v = vendors.find((v) => v.id === phone.vendor);
        currentVendorFeatures = v ? v.features || [] : [];
    }
</script>

{#if open}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    >
        <div
            class="bg-background dark:bg-slate-900 p-6 rounded-lg shadow-lg max-w-7xl w-full max-h-[90vh] flex flex-col border dark:border-slate-700"
        >
            <div class="flex justify-between items-center mb-4 shrink-0">
                <div>
                    <h2 class="text-lg font-semibold">
                        {$t("lines.title") || "Line Configuration"}. {$t("phone.number")}: {phone.phone_number}
                    </h2>
                    <p class="text-sm text-muted-foreground">
                        {$t("lines.description") ||
                            "Manage additional lines for this phone."}
                        ({lines.length} / {maxLines || "∞"})
                    </p>
                </div>
                <Button variant="ghost" size="icon" on:click={close}>
                    <X class="h-4 w-4" />
                </Button>
            </div>

            <div class="flex gap-6 flex-1 min-h-0">
                {#if imageUrl && !imageLoadError}
                    <div
                        class="w-1/3 bg-slate-50 dark:bg-slate-800/50 rounded-lg p-4 flex items-center justify-center border dark:border-slate-700"
                    >
                        <img
                            src={imageUrl}
                            alt="Phone"
                            class="max-w-full max-h-full object-contain"
                            on:error={() => (imageLoadError = true)}
                        />
                    </div>
                {/if}

                <div class="flex-1 overflow-y-auto pr-2 space-y-4">
                    <!-- Search and Add -->
                    <div class="flex justify-between items-center gap-4">
                        <div class="relative flex-1">
                            <Search
                                class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground"
                            />
                            <Input
                                placeholder={$t("common.search") || "Search..."}
                                class="pl-8"
                                bind:value={searchQuery}
                            />
                        </div>
                        <Button on:click={add} disabled={!!editForm}>
                            <Plus class="mr-2 h-4 w-4" />
                            {$t("common.add") || "Add Line"}
                        </Button>
                    </div>

                    <!-- Editor Form -->
                    {#if editForm}
                        <div
                            class="border-2 rounded-lg p-6 bg-slate-50 dark:bg-slate-800/50 border-slate-200 dark:border-slate-700 shadow-md space-y-4"
                        >
                            <h3
                                class="font-semibold text-lg border-b pb-2 mb-4"
                            >
                                {originalLine
                                    ? $t("lines.edit_item") || "Edit Item"
                                    : $t("lines.new_item") || "New Item"}
                            </h3>
                            <div class="grid grid-cols-3 gap-4">
                                <div class="space-y-2">
                                    <Label>Type</Label>
                                    <select
                                        class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                        bind:value={editForm.type}
                                    >
                                        <option value="line">Line</option>
                                        {#if maxHardKeys > 0}
                                            <option value="hard_key"
                                                >Hard Key</option
                                            >
                                        {/if}
                                        {#if maxSoftKeys > 0}
                                            <option value="soft_key"
                                                >Soft Key</option
                                            >
                                        {/if}
                                    </select>
                                </div>
                                <div class="space-y-2">
                                    <Label>Number (Sequential)</Label>
                                    <Input
                                        type="number"
                                        bind:value={editForm.number}
                                    />
                                </div>
                                {#if hasExpansionModules}
                                    <div class="space-y-2">
                                        <Label>Location</Label>
                                        <select
                                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                            bind:value={
                                                editForm.expansion_module_number
                                            }
                                        >
                                            <option value={0}>Main Phone</option
                                            >
                                            {#each Array(phone.expansion_modules_count || 0) as _, i}
                                                <option value={i + 1}
                                                    >Exp Module {i + 1}</option
                                                >
                                            {/each}
                                        </select>
                                    </div>
                                    {#if editForm.expansion_module_number > 0}
                                        <div class="space-y-2">
                                            <Label>Key # on Module</Label>
                                            <Input
                                                type="number"
                                                min="1"
                                                bind:value={editForm.key_number}
                                            />
                                        </div>
                                    {/if}
                                {/if}
                            </div>

                            <!-- Dynamic Fields based on Type -->
                            {#if editForm.type === "line"}
                                <div class="grid grid-cols-2 gap-4">
                                    <div class="space-y-2">
                                        <Label>Line Number</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.line_number
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Display Name</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.display_name
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>User Name</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.user_name
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Auth Name</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.auth_name
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Password</Label>
                                        <Input
                                            type="password"
                                            bind:value={additionalInfo.password}
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Screen Name</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.screen_name
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Registrar 1 IP</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.registrar1_ip
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Registrar 1 Port</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.registrar1_port
                                            }
                                        />
                                    </div>
                                </div>
                            {:else}
                                <!-- Keys -->
                                <div class="col-span-3 space-y-4">
                                    <div class="space-y-2">
                                        <Label>Function</Label>
                                        <select
                                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                            bind:value={additionalInfo.type}
                                        >
                                            {#each currentVendorFeatures as feature}
                                                <option value={feature.id}
                                                    >{feature.name}</option
                                                >
                                            {/each}
                                            <option value="custom"
                                                >Custom</option
                                            >
                                        </select>
                                    </div>

                                    {#if additionalInfo.type && additionalInfo.type !== "custom"}
                                        {#each currentVendorFeatures.find((f) => f.id === additionalInfo.type)?.params || [] as param}
                                            {#if param.type !== "hidden"}
                                                <div class="space-y-2">
                                                    <Label>{param.label}</Label>
                                                    {#if param.type === "select" && param.source === "lines"}
                                                        <select
                                                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                                            bind:value={
                                                                additionalInfo[
                                                                    param.id
                                                                ]
                                                            }
                                                        >
                                                            <option value=""
                                                                >Select Line</option
                                                            >
                                                            {#each workingLines.filter((l) => l.type === "line") as line}
                                                                <option
                                                                    value={line.number}
                                                                    >Line {line.number}</option
                                                                >
                                                            {/each}
                                                        </select>
                                                    {:else}
                                                        <Input
                                                            bind:value={
                                                                additionalInfo[
                                                                    param.id
                                                                ]
                                                            }
                                                        />
                                                    {/if}
                                                </div>
                                            {/if}
                                        {/each}
                                    {:else if additionalInfo.type === "custom"}
                                        <div class="grid grid-cols-2 gap-4">
                                            <div class="space-y-2">
                                                <Label>Label</Label>
                                                <Input
                                                    bind:value={
                                                        additionalInfo.label
                                                    }
                                                />
                                            </div>
                                            <div class="space-y-2">
                                                <Label>Value</Label>
                                                <Input
                                                    bind:value={
                                                        additionalInfo.value
                                                    }
                                                />
                                            </div>
                                            <div class="space-y-2">
                                                <Label>Type</Label>
                                                <Input
                                                    bind:value={
                                                        additionalInfo.custom_type
                                                    }
                                                    placeholder="e.g. blf"
                                                />
                                            </div>
                                        </div>
                                    {/if}
                                </div>
                            {/if}

                            <div class="space-y-2">
                                <Label>Description</Label>
                                <Input
                                    bind:value={additionalInfo.description}
                                />
                            </div>

                            <div class="flex justify-end gap-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    on:click={cancelEdit}
                                >
                                    <X class="mr-2 h-4 w-4" />
                                    {$t("common.cancel") || "Cancel"}
                                </Button>
                                <Button size="sm" on:click={save}>
                                    <Check class="mr-2 h-4 w-4" />
                                    OK
                                </Button>
                            </div>
                        </div>
                    {/if}

                    <!-- Table -->
                    <div class="border rounded-md">
                        <Table.Root>
                            <Table.Header>
                                <Table.Row>
                                    <Table.Head>#</Table.Head>
                                    <Table.Head>Type</Table.Head>
                                    <Table.Head>Description</Table.Head>
                                    {#if hasExpansionModules}
                                        <Table.Head>Exp/Key</Table.Head>
                                    {/if}
                                    <Table.Head class="text-right"
                                        >Actions</Table.Head
                                    >
                                </Table.Row>
                            </Table.Header>
                            <Table.Body>
                                {#each paginatedLines as line}
                                    <Table.Row>
                                        <Table.Cell>{line.number}</Table.Cell>
                                        <Table.Cell>{line.type}</Table.Cell>
                                        <Table.Cell
                                            >{getLineDescription(
                                                line,
                                            )}</Table.Cell
                                        >
                                        {#if hasExpansionModules}
                                            <Table.Cell
                                                >{line.expansion_module_number ||
                                                    "-"}/{line.key_number ||
                                                    "-"}</Table.Cell
                                            >
                                        {/if}
                                        <Table.Cell class="text-right">
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                on:click={() => edit(line)}
                                                disabled={!!editForm}
                                            >
                                                <Pencil class="h-4 w-4" />
                                            </Button>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                class="text-destructive"
                                                on:click={() => remove(line)}
                                                disabled={!!editForm}
                                            >
                                                <Trash2 class="h-4 w-4" />
                                            </Button>
                                        </Table.Cell>
                                    </Table.Row>
                                {/each}
                                {#if paginatedLines.length === 0}
                                    <Table.Row>
                                        <Table.Cell
                                            colspan={hasExpansionModules
                                                ? 5
                                                : 4}
                                            class="text-center text-muted-foreground"
                                        >
                                            {$t("common.no_results") ||
                                                "No lines found"}
                                        </Table.Cell>
                                    </Table.Row>
                                {/if}
                            </Table.Body>
                        </Table.Root>
                    </div>

                    <!-- Pagination -->
                    {#if totalPages > 1}
                        <div class="flex justify-center gap-2">
                            <Button
                                variant="outline"
                                size="sm"
                                disabled={currentPage === 1}
                                on:click={() => currentPage--}
                            >
                                Previous
                            </Button>
                            <span class="py-2 text-sm"
                                >Page {currentPage} of {totalPages}</span
                            >
                            <Button
                                variant="outline"
                                size="sm"
                                disabled={currentPage === totalPages}
                                on:click={() => currentPage++}
                            >
                                Next
                            </Button>
                        </div>
                    {/if}
                </div>
            </div>

            <div class="flex justify-end gap-2 mt-4 shrink-0">
                <Button variant="outline" on:click={close}
                    >{$t("common.cancel") || "Cancel"}</Button
                >
                <Button on:click={saveAll}>OK</Button>
            </div>
        </div>
    </div>
{/if}
