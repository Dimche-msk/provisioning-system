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
        Target,
        FileUp,
        FileDown,
    } from "lucide-svelte";
    import * as XLSX from "xlsx";
    import { toast } from "svelte-sonner";
    import { Switch } from "$lib/components/ui/switch";
    import * as Dialog from "$lib/components/ui/dialog";
    import type { Phone, PhoneLine, DeviceModel, ModelKey } from "$lib/types";

    export let lines: PhoneLine[] = [];
    export let maxSoftKeys = 0;
    maxSoftKeys;
    export let maxHardKeys = 0;
    maxHardKeys;

    export let image = "";

    export let maxLines = 0;
    export let open = false;
    export let phone: Phone = {} as Phone;
    export let model: DeviceModel | null = null;

    const dispatch = createEventDispatcher();

    let workingLines: PhoneLine[] = [];
    let selectedLine: PhoneLine | null = null;

    // Reset working copy when opening
    $: if (open && lines) {
        workingLines = JSON.parse(JSON.stringify(lines)).map((l: any) => ({
            ...l,
            account_number: l.account_number || l.number || 1,
            panel_number: l.panel_number === null ? null : l.panel_number || 0,
            key_number: l.key_number === null ? null : l.key_number || 0,
            type: l.type || "Line",
        }));
        selectedLine = null;
    }

    let searchQuery = "";
    let currentPage = 1;
    let itemsPerPage = 16;

    $: isLineEditorFiltered = searchQuery.length > 0;

    // Editing state
    let editForm: PhoneLine | null = null;
    let additionalInfo: Record<string, any> = {}; // Parsed JSON
    let showEditDialog = false;

    $: if (editForm) {
        showEditDialog = true;
    } else {
        showEditDialog = false;
    }

    function handleDialogChange(open: boolean) {
        if (!open) {
            cancelEdit();
        }
    }

    // Filtered lines
    $: filteredLines = workingLines.filter((l) => {
        const q = searchQuery.toLowerCase().trim();
        if (!q) return true;

        let info: Record<string, any> = {};
        try {
            info = typeof l.additional_info === "string" 
                ? JSON.parse(l.additional_info) 
                : (l.additional_info || {});
        } catch (e) {}

        // 1. Check direct fields
        if (String(l.account_number).includes(q)) return true;
        if (l.panel_number !== null && String(l.panel_number).includes(q)) return true;
        if (l.key_number !== null && String(l.key_number).includes(q)) return true;
        if (l.type.toLowerCase().includes(q)) return true;

        // 2. Check all additional info values
        for (const val of Object.values(info)) {
            if (String(val).toLowerCase().includes(q)) return true;
        }

        // 3. Check feature/type names
        const feature = currentVendorFeatures.find(f => f.id === l.type);
        if (feature?.name?.toLowerCase().includes(q)) return true;
        
        const keyType = model?.key_types?.find(kt => kt.id === l.type);
        if (keyType?.verbose?.toLowerCase().includes(q)) return true;

        // 4. Special case: search for "panel X key Y" without spaces if user types it
        const pk = `${l.panel_number}${l.key_number}`;
        if (l.panel_number !== null && pk.includes(q)) return true;

        return false;
    });

    $: totalPages = Math.ceil(filteredLines.length / itemsPerPage);
    $: paginatedLines = filteredLines.slice(
        (currentPage - 1) * itemsPerPage,
        currentPage * itemsPerPage,
    );

    $: hasExpansionModules = (phone.expansion_modules_count || 0) > 0;

    let originalLine: PhoneLine | null = null;

    // Background Image logic
    $: selectedKeyType = model?.key_types?.find(
        (kt) => kt.id === (editForm?.type || selectedLine?.type),
    );

    $: baseImageUrl =
        image && phone.vendor
            ? `/api/vendors-static/${phone.vendor}/static/${image}`
            : "";

    $: typeImageUrl =
        selectedKeyType?.image && phone.vendor
            ? `/api/vendors-static/${phone.vendor}/static/${selectedKeyType.image}`
            : "";

    // Find custom image for selected line
    $: currentModelKey =
        model && selectedLine
            ? model.keys.find(
                  (k) =>
                      k.index === selectedLine?.key_number &&
                      (selectedLine?.panel_number || 0) === 0,
              )
            : null;

    $: myImageUrl =
        currentModelKey?.my_image && phone.vendor
            ? `/api/vendors-static/${phone.vendor}/static/${currentModelKey.my_image}`
            : "";

    $: activeImageUrl = myImageUrl || typeImageUrl || baseImageUrl;

    let imageLoadError = false;
    $: if (activeImageUrl) imageLoadError = false;

    let naturalWidth = 0;
    let naturalHeight = 0;

    // Get coordinates for highlighting
    $: highlightPos = (() => {
        if (!selectedLine || !model || !naturalWidth || !naturalHeight)
            return null;
        const mk = model.keys.find(
            (k) =>
                k.index === selectedLine?.key_number &&
                k.type?.toLowerCase() === selectedLine?.type?.toLowerCase() &&
                (selectedLine?.panel_number || 0) === 0,
        );
        if (mk && mk.x > 0 && mk.y > 0) {
            return {
                left: (mk.x / naturalWidth) * 100,
                top: (mk.y / naturalHeight) * 100,
            };
        }
        return null;
    })();

    function selectLine(line: PhoneLine) {
        selectedLine = line;
    }

    function edit(line: PhoneLine) {
        originalLine = line;
        selectedLine = line;
        editForm = { ...line };
        try {
            additionalInfo = JSON.parse(line.additional_info || "{}");
            const currentType = line.type === "Line" ? "account" : line.type;
            const features =
                currentType === "account"
                    ? currentVendorAccounts
                    : currentVendorFeatures;
            const feature = features.find((f: any) => f.id === currentType);
            if (feature?.params) {
                for (const p of feature.params) {
                    if (
                        p.type === "boolean" &&
                        p.params &&
                        additionalInfo[p.id] === true
                    ) {
                        additionalInfo[p.id] = {};
                    }
                }
            }
        } catch (e) {
            console.warn("Failed to parse additional_info", e);
            additionalInfo = {};
        }
    }

    function add() {
        originalLine = null;
        selectedLine = null;
        editForm = {
            type: "Line",
            account_number: 1,
            panel_number: 0,
            key_number: 1,
            additional_info: "{}",
        } as PhoneLine;
        additionalInfo = {};
    }

    function addFunction() {
        originalLine = null;
        selectedLine = null;

        // Find first available non-button feature as default
        // Priority:
        // 1. Global features not already present
        // 2. Account-associated features
        const existingGlobalFeatures = new Set(
            phone?.lines
                ?.filter((l) => l.panel_number === null)
                .map((l) => l.type) || [],
        );
        const defaultFeature = currentVendorFeatures.find((f) => {
            if (f.associated_with_button || f.id === "Line") return false;
            if (!f.associated_with_account) {
                // Global feature -> check if exists
                return !existingGlobalFeatures.has(f.id);
            }
            // Account feature -> always allow
            return true;
        });

        editForm = {
            type: defaultFeature?.id || "custom",
            account_number: 1,
            panel_number: null,
            key_number: null,
            additional_info: "{}",
        } as any;
        additionalInfo = {};
    }

    function save() {
        if (!editForm || !editForm.account_number) {
            toast.error("Account Number is required");
            return;
        }

        // Ensure numbers are integers if they are not null
        if (editForm.account_number !== null)
            editForm.account_number = parseInt(
                String(editForm.account_number),
                10,
            );
        if (editForm.panel_number !== null)
            editForm.panel_number = parseInt(String(editForm.panel_number), 10);
        if (editForm.key_number !== null)
            editForm.key_number = parseInt(String(editForm.key_number), 10);

        // Validation: Check for duplicates (Type + Panel + Key must be unique)
        // General features don't have panel/key, so they might have multiple entries of same type?
        // User said: "может быть несколько записай для одного аппарата с одним и тем же типом доп. функции"
        for (const line of workingLines) {
            if (originalLine && line === originalLine) continue;

            if (
                editForm?.panel_number !== null &&
                line.type === editForm?.type &&
                line.panel_number === editForm?.panel_number &&
                line.key_number === editForm?.key_number
            ) {
                const typeName =
                    model?.key_types?.find((kt) => kt.id === editForm?.type)
                        ?.verbose || editForm?.type;
                const panelText =
                    editForm?.panel_number === 0
                        ? "Основная"
                        : `Панель ${editForm?.panel_number}`;
                toast.error(
                    `Дубликат: ${typeName}, ${panelText}, Кнопка ${editForm.key_number} уже назначена.`,
                );
                return;
            }
        }

        editForm.additional_info = JSON.stringify(additionalInfo);

        if (editForm && originalLine) {
            const idx = workingLines.indexOf(originalLine);
            if (idx !== -1 && editForm) {
                workingLines[idx] = { ...editForm };
            }
        } else if (editForm) {
            workingLines = [...workingLines, { ...editForm }];
        }
        originalLine = null;
        editForm = null;
        additionalInfo = {};
    }

    function remove(line: PhoneLine) {
        workingLines = workingLines.filter((l) => l !== line);
        if (selectedLine === line) selectedLine = null;
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

    function exportToExcel() {
        const data = workingLines.map((l) => {
            let info: any = {};
            try {
                info = typeof l.additional_info === "string" 
                    ? JSON.parse(l.additional_info) 
                    : (l.additional_info || {});
            } catch (e) {}

            return {
                Type: l.type,
                Account: l.account_number,
                Panel: l.panel_number,
                Key: l.key_number,
                ...info,
            };
        });

        const ws = XLSX.utils.json_to_sheet(data);
        const wb = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(wb, ws, "Lines");
        const filename = `lines_${phone.phone_number || phone.mac_address || "export"}.xlsx`;
        XLSX.writeFile(wb, filename);
        toast.success($t("lines.export_excel") + " OK");
    }

    let importConflicts: { existing: PhoneLine; new: PhoneLine }[] = [];
    let pendingImport: PhoneLine[] = [];
    let showConflictDialog = false;

    async function importFromExcel(e: Event) {
        const target = e.target as HTMLInputElement;
        const file = target.files?.[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (evt) => {
            try {
                const bstr = evt.target?.result;
                const wb = XLSX.read(bstr, { type: "binary" });
                const wsname = wb.SheetNames[0];
                const ws = wb.Sheets[wsname];
                const data = XLSX.utils.sheet_to_json(ws);

                const newLines: PhoneLine[] = data.map((row: any) => {
                    const line: PhoneLine = {
                        type: row.Type || "Line",
                        account_number: parseInt(row.Account) || 1,
                        panel_number: row.Panel === undefined || row.Panel === null || row.Panel === "" ? null : parseInt(row.Panel),
                        key_number: row.Key === undefined || row.Key === null || row.Key === "" ? null : parseInt(row.Key),
                        additional_info: "{}",
                    } as PhoneLine;

                    const info: any = {};
                    const reserved = ["Type", "Account", "Panel", "Key"];
                    for (const [k, v] of Object.entries(row)) {
                        if (!reserved.includes(k)) {
                            info[k] = v;
                        }
                    }
                    line.additional_info = JSON.stringify(info);
                    return line;
                });

                // Find conflicts
                const conflicts: { existing: PhoneLine; new: PhoneLine }[] = [];
                const nonConflicting: PhoneLine[] = [];

                for (const nl of newLines) {
                    const existing = workingLines.find(
                        (el) =>
                            el.panel_number === nl.panel_number &&
                            el.key_number === nl.key_number &&
                            el.panel_number !== null, // Only buttons can conflict
                    );

                    if (existing) {
                        conflicts.push({ existing, new: nl });
                    } else {
                        nonConflicting.push(nl);
                    }
                }

                if (conflicts.length > 0) {
                    importConflicts = conflicts;
                    pendingImport = nonConflicting;
                    showConflictDialog = true;
                } else {
                    workingLines = [...workingLines, ...newLines];
                    toast.success(`${$t("phones.import")} OK: ${newLines.length}`);
                }
            } catch (err: any) {
                toast.error("Failed to parse Excel: " + err.message);
            }
            target.value = ""; // Reset input
        };
        reader.readAsBinaryString(file);
    }

    function resolveConflicts(resolution: "overwrite" | "skip") {
        if (resolution === "overwrite") {
            const updatedLines = [...workingLines];
            for (const conflict of importConflicts) {
                const idx = updatedLines.indexOf(conflict.existing);
                if (idx !== -1) {
                    updatedLines[idx] = conflict.new;
                }
            }
            workingLines = [...updatedLines, ...pendingImport];
            toast.success(`${$t("phones.import")} OK: ${importConflicts.length + pendingImport.length} (${$t("common.overwrite")})`);
        } else {
            workingLines = [...workingLines, ...pendingImport];
            toast.success(`${$t("phones.import")} OK: ${pendingImport.length} (${$t("common.skip")})`);
        }
        showConflictDialog = false;
        importConflicts = [];
        pendingImport = [];
    }

    function getLineDescription(line: PhoneLine) {
        let info: Record<string, any> = {};
        try {
            info = JSON.parse(line.additional_info || "{}");
        } catch (e) {}

        if (line.type === "Line") {
            return info.display_name || info.label || "";
        } else {
            const feature = currentVendorFeatures.find(
                (f) => f.id === line.type || f.id === info.type,
            );
            return info.label || info.value || feature?.name || line.type;
        }
    }

    function getLineValue(line: PhoneLine) {
        let info: Record<string, any> = {};
        try {
            info = typeof line.additional_info === "string" 
                ? JSON.parse(line.additional_info) 
                : (line.additional_info || {});
        } catch (e) {}

        const values: string[] = [];
        // Common numeric fields to look for first
        const priorityFields = ['user_name', 'auth_name', 'value', 'number', 'line_number', 'extension', 'directory_number'];
        
        for (const field of priorityFields) {
            if (info[field] && /\d/.test(String(info[field]))) {
                values.push(String(info[field]));
            }
        }

        // If nothing found in priority fields, look at all fields
        if (values.length === 0) {
            for (const [key, val] of Object.entries(info)) {
                if (priorityFields.includes(key)) continue;
                if ((typeof val === 'string' || typeof val === 'number') && /\d/.test(String(val))) {
                    const sVal = String(val);
                    // Skip very long strings (likely not numbers/IDs) and potential passwords
                    if (sVal.length < 24 && !key.toLowerCase().includes('pass')) {
                        values.push(sVal);
                    }
                }
            }
        }

        // Deduplicate and join
        return [...new Set(values)].join(", ");
    }

    let vendors: any[] = [];
    let currentVendorFeatures: any[] = [];
    let currentVendorAccounts: any[] = [];

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
        currentVendorAccounts = v ? v.accounts || [] : [];
    }

    let currentEditFeature: any = null;
    $: {
        currentEditFeature =
            editForm &&
            currentVendorFeatures.find((f) => f.id === editForm?.type);

        if (
            editForm &&
            !originalLine &&
            currentEditFeature?.associated_with_button &&
            editForm.panel_number === null
        ) {
            // No automatic fixing of panel_number here - the filtering prevents this state
            // and we want to keep it as null if it's a general function
        }
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
                        {$t("lines.title") || "Line Configuration"}. {$t(
                            "phone.number",
                        )}: {phone.phone_number}
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
                {#if activeImageUrl}
                    <div
                        class="w-1/3 bg-slate-50 dark:bg-slate-800/50 rounded-lg p-4 border dark:border-slate-700 relative overflow-hidden"
                    >
                        <div class="relative inline-block">
                            <img
                                src={activeImageUrl}
                                alt="Phone"
                                class="max-w-full max-h-full object-contain"
                                bind:naturalWidth
                                bind:naturalHeight
                                on:error={() => (imageLoadError = true)}
                            />
                            {#if highlightPos}
                                <div
                                    class="absolute pointer-events-none flex items-center justify-center"
                                    style="left: {highlightPos.left}%; top: {highlightPos.top}%; transform: translate(-50%, -50%);"
                                >
                                    <div class="relative">
                                        <!-- Ring animation -->
                                        <div
                                            class="absolute inset-0 rounded-full border-4 border-red-500 animate-ping opacity-75"
                                        ></div>
                                        <Target
                                            class="h-8 w-8 text-red-600 drop-shadow-[0_0_5px_rgba(255,255,255,0.8)]"
                                        />
                                    </div>
                                </div>
                            {/if}
                        </div>
                        {#if imageLoadError}
                            <div
                                class="absolute inset-0 flex items-center justify-center bg-muted/50"
                            >
                                <span class="text-sm text-muted-foreground"
                                    >Image not found</span
                                >
                            </div>
                        {/if}
                    </div>
                {/if}

                <div class="flex-1 overflow-y-auto pr-2 space-y-4">
                    <!-- Search and Add -->
                    <div class="flex justify-between items-center gap-4">
                        <div
                            class="relative flex-1 group transition-all duration-300 {isLineEditorFiltered
                                ? 'ring-2 ring-blue-500/50 rounded-md'
                                : ''}"
                        >
                            <Search
                                class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground"
                            />
                            <Input
                                placeholder={$t("common.search") || "Search..."}
                                class="pl-8 pr-8"
                                bind:value={searchQuery}
                            />
                            {#if isLineEditorFiltered}
                                <button
                                    class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                                    on:click={() => (searchQuery = "")}
                                >
                                    <X class="h-4 w-4" />
                                </button>
                            {/if}
                        </div>
                        <div class="flex gap-2">
                            <Button
                                on:click={exportToExcel}
                                variant="outline"
                                title={$t("lines.export_excel")}
                            >
                                <FileDown class="h-4 w-4" />
                            </Button>
                            <div class="relative">
                                <Input
                                    type="file"
                                    accept=".xlsx, .xls"
                                    class="hidden"
                                    id="excel-import"
                                    on:change={importFromExcel}
                                />
                                <Button
                                    variant="outline"
                                    on:click={() =>
                                        document
                                            .getElementById("excel-import")
                                            ?.click()}
                                    title={$t("lines.import_excel")}
                                >
                                    <FileUp class="h-4 w-4" />
                                </Button>
                            </div>
                            <Button
                                on:click={add}
                                variant="outline"
                                disabled={!!editForm}
                            >
                                <Plus class="mr-2 h-4 w-4" />
                                {phone?.type === 'gateway' 
                                    ? ($t("add") || "Add Line") 
                                    : ($t("common.add_key") || "Add Key")}
                            </Button>
                            {#if phone?.type === 'phone'}
                                <Button
                                    on:click={addFunction}
                                    variant="outline"
                                    disabled={!!editForm}
                                >
                                    <Plus class="mr-2 h-4 w-4" />
                                    {$t("common.add_function") || "Add Function"}
                                </Button>
                            {/if}
                        </div>
                    </div>

                    <!-- Editor Form -->
                    <!-- Editor Dialog -->
                    <Dialog.Root open={showEditDialog} onOpenChange={handleDialogChange}>
                        <Dialog.Content class="max-w-4xl max-h-[90vh] overflow-y-auto">
                            {#if editForm}
                                <Dialog.Header>
                                    <Dialog.Title>
                                        {originalLine
                                            ? $t("lines.edit_item") || "Edit Item"
                                            : $t("lines.new_item") || "New Item"}
                                    </Dialog.Title>
                                </Dialog.Header>

                                <div class="space-y-6 py-4">
                                    <div class="grid grid-cols-4 gap-4 items-end">
                                        {#if phone?.type === 'phone'}
                                            <div class="space-y-1.5">
                                                <Label class="text-xs text-muted-foreground uppercase font-bold">Тип</Label>
                                                <select
                                                    class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                                                    bind:value={editForm.type}
                                                >
                                                    {#if editForm.panel_number !== null}
                                                        <option value="Line">Линия</option>
                                                        {#each currentVendorFeatures.filter((f) => f.associated_with_button) as f}
                                                            <option value={f.id}>{f.name}</option>
                                                        {/each}
                                                    {:else}
                                                        {#each currentVendorFeatures.filter( (f) => {
                                                                if (f.associated_with_button || f.id === "Line") return false;
                                                                if (!f.associated_with_account) {
                                                                    const alreadyExists = phone?.lines?.some((l) => l.type === f.id && l.panel_number === null);
                                                                    return !alreadyExists || (originalLine && originalLine.type === f.id);
                                                                }
                                                                return true;
                                                            }, ) as f}
                                                            <option value={f.id}>{f.name}</option>
                                                        {/each}
                                                    {/if}
                                                    <option value="custom">Другое</option>
                                                </select>
                                            </div>
                                        {/if}

                                        {#if editForm.panel_number !== null || currentEditFeature?.associated_with_button}
                                            <div class="space-y-1.5">
                                                <Label class="text-xs text-muted-foreground uppercase font-bold">Аккаунт #</Label>
                                                <Input class="h-9" type="number" bind:value={editForm.account_number} />
                                            </div>
                                            <div class="space-y-1.5">
                                                <Label class="text-xs text-muted-foreground uppercase font-bold">Кнопка</Label>
                                                {#if model && model.keys && model.keys.length > 0 && Number(editForm.panel_number) === 0}
                                                    <select
                                                        class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                                                        bind:value={editForm.key_number}
                                                    >
                                                        <option value={0}>-- Выберите кнопку --</option>
                                                        {#each model.keys as mk}
                                                            <option value={mk.index}>{mk.label} ({mk.type})</option>
                                                        {/each}
                                                    </select>
                                                {:else}
                                                    <Input class="h-9" type="number" min="1" bind:value={editForm.key_number} />
                                                {/if}
                                            </div>
                                            {#if phone?.expansion_modules_count && phone.expansion_modules_count > 0}
                                                <div class="space-y-1.5">
                                                    <Label class="text-xs text-muted-foreground uppercase font-bold">Устройство</Label>
                                                    <select
                                                        class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                                                        bind:value={editForm.panel_number}
                                                    >
                                                        <option value={0}>Телефон</option>
                                                        {#each Array.from({ length: phone.expansion_modules_count }, (_, i) => i + 1) as i}
                                                            <option value={i}>Панель {i}</option>
                                                        {/each}
                                                    </select>
                                                </div>
                                            {/if}
                                        {:else if currentEditFeature?.associated_with_account}
                                            <div class="space-y-1.5 {phone?.type === 'gateway' ? 'col-span-3' : 'col-span-1'}">
                                                <Label class="text-xs text-muted-foreground uppercase font-bold">Аккаунт #</Label>
                                                <Input class="h-9" type="number" bind:value={editForm.account_number} />
                                            </div>
                                        {/if}
                                    </div>

                                    <!-- Dynamic Fields -->
                                    {#if editForm.type === "Line"}
                                        <div class="grid grid-cols-2 gap-4">
                                            {#if currentVendorAccounts && currentVendorAccounts.length > 0}
                                                {#each currentVendorAccounts.find((a) => a.id === "account")?.params || [] as param}
                                                    {#if param.type !== "hidden"}
                                                        <div class="space-y-2">
                                                            <Label>{param.label}</Label>
                                                            {#if param.type === "boolean"}
                                                                <div class="flex h-10 items-center">
                                                                    <Switch checked={!!additionalInfo[param.id]} on:change={(e) => {
                                                                        const checked = e.detail;
                                                                        if (checked) {
                                                                            if (param.params && param.params.length > 0) {
                                                                                if (!additionalInfo[param.id] || typeof additionalInfo[param.id] !== "object") {
                                                                                    additionalInfo[param.id] = {};
                                                                                }
                                                                            } else {
                                                                                additionalInfo[param.id] = true;
                                                                            }
                                                                        } else {
                                                                            additionalInfo[param.id] = false;
                                                                        }
                                                                    }} />
                                                                </div>
                                                                {#if additionalInfo[param.id] && typeof additionalInfo[param.id] === "object" && param.params}
                                                                    <div class="pl-4 border-l-2 border-primary/20 space-y-4 pt-2 col-span-full">
                                                                        {#each param.params as subParam}
                                                                            <div class="space-y-2">
                                                                                <Label>{subParam.label}</Label>
                                                                                <Input bind:value={additionalInfo[param.id][subParam.id]} />
                                                                            </div>
                                                                        {/each}
                                                                    </div>
                                                                {/if}
                                                            {:else}
                                                                <Input type={param.type === "password" ? "password" : "text"} bind:value={additionalInfo[param.id]} />
                                                            {/if}
                                                        </div>
                                                    {/if}
                                                {/each}
                                            {/if}
                                        </div>
                                    {:else}
                                        <!-- Features -->
                                        <div class="space-y-4">
                                            {#if currentEditFeature}
                                                <div class="grid grid-cols-2 gap-4">
                                                    {#each currentEditFeature.params || [] as param}
                                                        {#if param.type !== "hidden"}
                                                            <div class="space-y-2">
                                                                <Label>{param.label}</Label>
                                                                {#if param.type === "select"}
                                                                    <select class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={additionalInfo[param.id]}>
                                                                        <option value="">Выберите...</option>
                                                                        {#if param.source === "lines"}
                                                                            {#each workingLines.filter((l) => l.type === "Line") as line}
                                                                                <option value={line.account_number}>Линия {line.account_number}</option>
                                                                            {/each}
                                                                        {:else if param.options}
                                                                            {#each param.options as opt}
                                                                                <option value={opt.value}>{opt.label}</option>
                                                                            {/each}
                                                                        {/if}
                                                                    </select>
                                                                {:else if param.type === "boolean"}
                                                                    <div class="flex h-10 items-center">
                                                                        <Switch checked={!!additionalInfo[param.id]} on:change={(e) => {
                                                                            const checked = e.detail;
                                                                            if (checked) {
                                                                                if (param.params && param.params.length > 0) {
                                                                                    if (!additionalInfo[param.id] || typeof additionalInfo[param.id] !== "object") {
                                                                                        additionalInfo[param.id] = {};
                                                                                    }
                                                                                } else {
                                                                                    additionalInfo[param.id] = true;
                                                                                }
                                                                            } else {
                                                                                additionalInfo[param.id] = false;
                                                                            }
                                                                        }} />
                                                                    </div>
                                                                    {#if additionalInfo[param.id] && typeof additionalInfo[param.id] === "object" && param.params}
                                                                        <div class="pl-4 border-l-2 border-primary/20 space-y-4 pt-2 col-span-full">
                                                                            {#each param.params as subParam}
                                                                                <div class="space-y-2">
                                                                                    <Label>{subParam.label}</Label>
                                                                                    <Input bind:value={additionalInfo[param.id][subParam.id]} />
                                                                                </div>
                                                                            {/each}
                                                                        </div>
                                                                    {/if}
                                                                {:else}
                                                                    <Input bind:value={additionalInfo[param.id]} />
                                                                {/if}
                                                            </div>
                                                        {/if}
                                                    {/each}
                                                </div>
                                            {:else if editForm?.type === "custom"}
                                                <div class="grid grid-cols-2 gap-4">
                                                    <div class="space-y-2"><Label>Метка</Label><Input bind:value={additionalInfo.label} /></div>
                                                    <div class="space-y-2"><Label>Значение</Label><Input bind:value={additionalInfo.value} /></div>
                                                    <div class="space-y-2"><Label>Тип</Label><Input bind:value={additionalInfo.custom_type} placeholder="например: blf" /></div>
                                                </div>
                                            {/if}
                                        </div>
                                    {/if}

                                    <div class="space-y-2">
                                        <Label>Описание</Label>
                                        <Input bind:value={additionalInfo.description} />
                                    </div>
                                </div>

                                <Dialog.Footer>
                                    <Button variant="outline" on:click={cancelEdit}>
                                        <X class="mr-2 h-4 w-4" />
                                        {$t("common.cancel") || "Cancel"}
                                    </Button>
                                    <Button on:click={save}>
                                        <Check class="mr-2 h-4 w-4" />
                                        OK
                                    </Button>
                                </Dialog.Footer>
                            {/if}
                        </Dialog.Content>
                    </Dialog.Root>

                    <!-- Table -->
                    <div class="border rounded-md overflow-hidden">
                        <Table.Root>
                            <Table.Header>
                                <Table.Row>
                                    <Table.Head class="w-[80px]"
                                        >Акк #</Table.Head
                                    >
                                    <Table.Head class="w-[120px]"
                                        >Панель / Кнопка</Table.Head
                                    >
                                    <Table.Head class="w-[100px]"
                                        >{$t("lines.type") || "Тип"}</Table.Head
                                    >
                                    <Table.Head>{$t("common.value") || "Значение"}</Table.Head>
                                    <Table.Head>{$t("common.description") || "Описание"}</Table.Head>
                                    <Table.Head class="text-right"
                                        >Действия</Table.Head
                                    >
                                </Table.Row>
                            </Table.Header>
                            <Table.Body>
                                {#each paginatedLines as line}
                                    <Table.Row
                                        on:click={() => selectLine(line)}
                                        class="cursor-pointer transition-colors {selectedLine ===
                                        line
                                            ? 'bg-blue-50 dark:bg-blue-900/20'
                                            : ''}"
                                    >
                                        <Table.Cell class="font-medium">
                                            {line.account_number}
                                        </Table.Cell>
                                        <Table.Cell>
                                            {#if line.panel_number !== null && line.key_number !== null}
                                                {line.panel_number === 0
                                                    ? "Осн."
                                                    : `Расш ${line.panel_number}`}
                                                /
                                                {line.key_number}
                                            {:else}
                                                <span
                                                    class="text-muted-foreground italic"
                                                    >Общая</span
                                                >
                                            {/if}
                                        </Table.Cell>
                                        <Table.Cell>
                                            <span
                                                class="capitalize text-xs font-semibold px-2 py-1 rounded bg-muted"
                                            >
                                                {model?.key_types?.find(
                                                    (kt) => kt.id === line.type,
                                                )?.verbose ||
                                                    line.type.replace("_", " ")}
                                            </span>
                                        </Table.Cell>
                                        <Table.Cell
                                            >{getLineValue(line)}</Table.Cell
                                        >
                                        <Table.Cell
                                            >{getLineDescription(
                                                line,
                                            )}</Table.Cell
                                        >
                                        <Table.Cell class="text-right">
                                            <div class="flex justify-end gap-1">
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    on:click={(e) => {
                                                        e.stopPropagation();
                                                        edit(line);
                                                    }}
                                                    disabled={!!editForm}
                                                >
                                                    <Pencil class="h-4 w-4" />
                                                </Button>
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    class="text-destructive hover:text-destructive hover:bg-destructive/10"
                                                    on:click={(e) => {
                                                        e.stopPropagation();
                                                        remove(line);
                                                    }}
                                                    disabled={!!editForm}
                                                >
                                                    <Trash2 class="h-4 w-4" />
                                                </Button>
                                            </div>
                                        </Table.Cell>
                                    </Table.Row>
                                {/each}
                                {#if paginatedLines.length === 0}
                                    <Table.Row>
                                        <Table.Cell
                                            colspan={5}
                                            class="text-center py-12"
                                        >
                                            <div
                                                class="flex flex-col items-center justify-center space-y-2"
                                            >
                                                <div
                                                    class="p-3 bg-muted rounded-full text-muted-foreground"
                                                >
                                                    <Search class="h-6 w-6" />
                                                </div>
                                                <p
                                                    class="text-sm font-medium text-muted-foreground"
                                                >
                                                    {$t("common.no_results") ||
                                                        "No lines found."}
                                                </p>
                                                {#if isLineEditorFiltered}
                                                    <Button
                                                        variant="link"
                                                        size="sm"
                                                        class="h-auto p-0"
                                                        on:click={() =>
                                                            (searchQuery = "")}
                                                    >
                                                        {$t("common.clear") ||
                                                            "Clear Search"}
                                                    </Button>
                                                {/if}
                                            </div>
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
                                Назад
                            </Button>
                            <span class="py-2 text-sm"
                                >Страница {currentPage} из {totalPages}</span
                            >
                            <Button
                                variant="outline"
                                size="sm"
                                disabled={currentPage === totalPages}
                                on:click={() => currentPage++}
                            >
                                Вперед
                            </Button>
                        </div>
                    {/if}
                </div>
            </div>

            <div class="flex justify-end gap-2 mt-4 shrink-0">
                <Button
                    variant="outline"
                    on:click={close}
                    disabled={!!editForm}
                >
                    {$t("common.cancel") || "Cancel"}
                </Button>
                <Button on:click={saveAll} disabled={!!editForm}>OK</Button>
            </div>
        </div>
    </div>
{/if}

{#if showConflictDialog}
    <div
        class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60"
    >
        <div class="bg-background dark:bg-slate-900 p-6 rounded-lg shadow-2xl border dark:border-slate-700 max-w-lg w-full">
            <h3 class="text-lg font-semibold mb-2">{$t("lines.conflicts_found")}</h3>
            <p class="text-sm text-muted-foreground mb-4">
                {$t("lines.conflicts_desc", { values: { count: importConflicts.length } })}
            </p>

            <div class="max-h-48 overflow-y-auto mb-6 border rounded p-2 text-xs space-y-1">
                {#each importConflicts as conflict}
                    <div class="flex justify-between border-b pb-1">
                        <span>{$t("phone.exp_count")} {conflict.existing.panel_number}, {$t("common.add_key")} {conflict.existing.key_number}</span>
                        <span class="text-muted-foreground">→ {conflict.new.type}</span>
                    </div>
                {/each}
            </div>

            <div class="flex justify-end gap-3">
                <Button variant="outline" on:click={() => (showConflictDialog = false)}>{$t("common.cancel")}</Button>
                <Button variant="secondary" on:click={() => resolveConflicts("skip")}>{$t("common.skip")}</Button>
                <Button variant="destructive" on:click={() => resolveConflicts("overwrite")}>{$t("common.overwrite")}</Button>
            </div>
        </div>
    </div>
{/if}
