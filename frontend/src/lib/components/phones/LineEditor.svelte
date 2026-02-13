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
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import type { Phone, PhoneLine, DeviceModel, ModelKey } from "$lib/types";

    export let lines: PhoneLine[] = [];
    export let maxSoftKeys = 0;
    export let maxHardKeys = 0;

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
            panel_number: l.panel_number || l.expansion_module_number || 0,
            key_number: l.key_number || 0,
            type: l.type || "Line",
        }));
        selectedLine = null;
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
        let info: Record<string, any> = {};
        try {
            info = JSON.parse(l.additional_info || "{}");
        } catch (e) {}

        const searchStr = [
            l.account_number,
            l.panel_number,
            l.key_number,
            l.type,
            info.display_name,
            info.user_name,
            info.label,
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

    // Background Image logic
    $: selectedKeyType = model?.key_types?.find(kt => kt.id === (editForm?.type || selectedLine?.type));
    
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

    // Get coordinates for highlighting
    $: highlightCoords = (() => {
        if (!selectedLine || !model) return null;
        const mk = model.keys.find(
            (k) =>
                k.index === selectedLine?.key_number &&
                (selectedLine?.panel_number || 0) === 0,
        );
        if (mk && mk.x > 0 && mk.y > 0) {
            return { x: mk.x, y: mk.y };
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
        } catch (e) {
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
        };
        additionalInfo = {};
    }

    function save() {
        if (!editForm || !editForm.account_number) {
            toast.error("Account Number is required");
            return;
        }

        // Ensure numbers are integers
        editForm.account_number = parseInt(String(editForm.account_number), 10);
        editForm.panel_number = parseInt(String(editForm.panel_number), 10);
        editForm.key_number = parseInt(String(editForm.key_number), 10);

        // Validation: Check for duplicates (Type + Panel + Key must be unique)
        for (const line of workingLines) {
            if (originalLine && line === originalLine) continue;

            if (
                line.type === editForm.type &&
                line.panel_number === editForm.panel_number &&
                line.key_number === editForm.key_number
            ) {
                const typeName = model?.key_types?.find(kt => kt.id === editForm.type)?.verbose || editForm.type;
                const panelText = editForm.panel_number === 0 ? "Основная" : `Панель ${editForm.panel_number}`;
                toast.error(
                    `Дубликат: ${typeName}, ${panelText}, Кнопка ${editForm.key_number} уже назначена.`,
                );
                return;
            }
        }

        editForm.additional_info = JSON.stringify(additionalInfo);

        if (originalLine) {
            const idx = workingLines.indexOf(originalLine);
            if (idx !== -1) {
                workingLines[idx] = { ...editForm };
            }
        } else {
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

    function getLineDescription(line: PhoneLine) {
        let info: Record<string, any> = {};
        try {
            info = JSON.parse(line.additional_info || "{}");
        } catch (e) {}

        if (line.type === "Line") {
            return info.display_name || info.label || "";
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
                        <div
                            class="relative w-full h-full flex items-center justify-center"
                        >
                            <img
                                src={activeImageUrl}
                                alt="Phone"
                                class="max-w-full max-h-full object-contain"
                                on:error={() => (imageLoadError = true)}
                            />
                            {#if highlightCoords}
                                <div
                                    class="absolute pointer-events-none flex items-center justify-center"
                                    style="left: {highlightCoords.x}px; top: {highlightCoords.y}px; transform: translate(-50%, -50%);"
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
                            <div class="grid {phone.expansion_modules_count > 0 ? 'grid-cols-4' : 'grid-cols-3'} gap-4">
                                <div class="space-y-2">
                                    <Label>Тип</Label>
                                    <select
                                        class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                        bind:value={editForm.type}
                                    >
                                        {#if model?.key_types && model.key_types.length > 0}
                                            {#each model.key_types as kt}
                                                <option value={kt.id}>{kt.verbose || kt.id}</option>
                                            {/each}
                                        {:else}
                                            <option value="Line">Линия</option>
                                            <option value="Free">Свободно</option>
                                            <option value="topsoftkey">Верхняя клавиша</option>
                                            <option value="softkey">Программная клавиша</option>
                                            <option value="bottomkey">Нижняя клавиша</option>
                                            {#if maxHardKeys > 0}
                                                <option value="hard_key">Клавиша корпуса</option>
                                            {/if}
                                        {/if}
                                    </select>
                                </div>
                                <div class="space-y-2">
                                    <Label>Аккаунт #</Label>
                                    <Input
                                        type="number"
                                        bind:value={editForm.account_number}
                                    />
                                </div>
                                <div class="space-y-2">
                                    <Label>Кнопка #</Label>
                                    <Input
                                        type="number"
                                        min="1"
                                        bind:value={editForm.key_number}
                                    />
                                </div>
                                {#if phone.expansion_modules_count > 0}
                                    <div class="space-y-2">
                                        <Label>Панель #</Label>
                                        <Input
                                            type="number"
                                            min="0"
                                            max={phone.expansion_modules_count || 0}
                                            bind:value={editForm.panel_number}
                                        />
                                    </div>
                                {/if}
                            </div>

                            <!-- Dynamic Fields based on Type -->
                            {#if editForm.type === "Line"}
                                <div class="grid grid-cols-3 gap-4">
                                    <div class="space-y-2">
                                        <Label>Номер линии</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.line_number
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Отображаемое имя</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.display_name
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Имя пользователя</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.user_name
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Имя авторизации</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.auth_name
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Пароль</Label>
                                        <Input
                                            type="password"
                                            bind:value={additionalInfo.password}
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Имя на экране</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.screen_name
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>IP Регистратора 1</Label>
                                        <Input
                                            bind:value={
                                                additionalInfo.registrar1_ip
                                            }
                                        />
                                    </div>
                                    <div class="space-y-2">
                                        <Label>Порт Регистратора 1</Label>
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
                                        <Label>Функция</Label>
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
                                                >Другое</option
                                            >
                                        </select>
                                    </div>

                                    {#if additionalInfo.type && additionalInfo.type !== "custom"}
                                        <div class="grid grid-cols-3 gap-4">
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
                                                                    >Выберите линию</option
                                                                >
                                                                {#each workingLines.filter((l) => l.type === "Line") as line}
                                                                    <option
                                                                        value={line.account_number}
                                                                        >Линия {line.account_number}</option
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
                                        </div>
                                    {:else if additionalInfo.type === "custom"}
                                        <div class="grid grid-cols-2 gap-4">
                                            <div class="space-y-2">
                                                <Label>Метка</Label>
                                                <Input
                                                    bind:value={
                                                        additionalInfo.label
                                                    }
                                                />
                                            </div>
                                            <div class="space-y-2">
                                                <Label>Значение</Label>
                                                <Input
                                                    bind:value={
                                                        additionalInfo.value
                                                    }
                                                />
                                            </div>
                                            <div class="space-y-2">
                                                <Label>Тип</Label>
                                                <Input
                                                    bind:value={
                                                        additionalInfo.custom_type
                                                    }
                                                    placeholder="например: blf"
                                                />
                                            </div>
                                        </div>
                                    {/if}
                                </div>
                            {/if}

                            <div class="space-y-2">
                                <Label>Описание</Label>
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
                                        >Тип</Table.Head
                                    >
                                    <Table.Head>Описание</Table.Head>
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
                                            {line.panel_number === 0
                                                ? "Осн."
                                                : `Расш ${line.panel_number}`} / {line.key_number}
                                        </Table.Cell>
                                        <Table.Cell>
                                            <span
                                                class="capitalize text-xs font-semibold px-2 py-1 rounded bg-muted"
                                            >
                                                {model?.key_types?.find(kt => kt.id === line.type)?.verbose || line.type.replace("_", " ")}
                                            </span>
                                        </Table.Cell>
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
                                            class="text-center py-8 text-muted-foreground"
                                        >
                                            {$t("common.no_results") ||
                                                "No lines configured."}
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
                <Button variant="outline" on:click={close}
                    >{$t("common.cancel") || "Cancel"}</Button
                >
                <Button on:click={saveAll}>OK</Button>
            </div>
        </div>
    </div>
{/if}
