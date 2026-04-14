<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import {
        ConfigParser,
        type DiscoveryEngine,
        type Diff,
    } from "$lib/migration/engine";
    import {
        ArrowRight,
        ChevronRight,
        ChevronLeft,
        CheckCircle2,
        AlertCircle,
        Play,
        Square,
        Pause,
        ListFilter,
        Trash2,
        Plus,
        Table,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import type { DeviceModel } from "$lib/types";
    import { X, Edit3 } from "lucide-svelte";

    export let files: { name: string; content: string }[] = [];
    export let engine: DiscoveryEngine;
    export let vendor: any;
    export let availableModels: DeviceModel[] = [];
    export let selectedModelId: string = "";

    const dispatch = createEventDispatcher();

    $: currentModelId = currentExtracted["phone.model_id"] || selectedModelId;
    $: currentModel = availableModels.find(m => m.id === currentModelId);

    // Calculate expansion module capacity
    $: supportedExtModels = currentModel?.supported_expansion_modules
        ? availableModels.filter(m => currentModel.supported_expansion_modules.includes(m.id))
        : [];
    $: maxKeysPerPanel = Math.max(0, ...supportedExtModels.map(m => (m.own_soft_keys || 0) + (m.own_hard_keys || 0)));
    $: maxPanels = currentModel?.maximum_expansion_modules || 0;

    let lastModelId = "";
    let lastVendorId = "";
    let MAPPING_OPTIONS: any[] = [];
    let flatMappedOptions: any[] = [];

    $: if (currentModelId !== lastModelId || vendor?.id !== lastVendorId) {
        lastModelId = currentModelId;
        lastVendorId = vendor?.id;
        
        MAPPING_OPTIONS = [
            { group: "Standard Fields", items: [
                { label: "Ignore", value: "ignore" },
                { label: "Phone MAC", value: "phone.mac_address" },
                { label: "Phone Number", value: "phone.phone_number" },
                { label: "Phone Description", value: "phone.description" },
            ]},
            
            // Account Lines
            { 
                group: "Account Lines", 
                items: Array.from({ length: currentModel?.max_account_lines || 6 }).flatMap((_, i) => {
                    const accIdx = i;
                    const lineNum = i + 1;
                    
                    const items = [
                        { label: `Line ${lineNum} Account #`, value: `lines[${accIdx}].account_number` }
                    ];

                    if (vendor?.accounts) {
                        vendor.accounts.forEach((group: any) => {
                            if (group.params) {
                                group.params.forEach((p: any) => {
                                    items.push({ 
                                        label: `Line ${lineNum} ${p.label}`, 
                                        value: `lines[${accIdx}].${p.id}` 
                                    });
                                });
                            }
                        });
                    } else {
                        // Fallback to standard ones if metadata is missing
                        items.push(
                            { label: `Line ${lineNum} Username`, value: `lines[${accIdx}].user_name` },
                            { label: `Line ${lineNum} Password`, value: `lines[${accIdx}].password` },
                            { label: `Line ${lineNum} Label`, value: `lines[${accIdx}].label` }
                        );
                    }
                    return items;
                })
            },
    
            // Global Features (not button associated)
            ...(vendor?.features?.filter((f: any) => !f.associated_with_button).length > 0 ? [{
                group: "Global Features",
                items: vendor.features.filter((f: any) => !f.associated_with_button).flatMap((f: any) => 
                    f.params.map((p: any) => ({
                        label: `${f.name}: ${p.label}`,
                        value: `feature.${f.id}.${p.id}`
                    }))
                )
            }] : []),
    
            // Button Slots (SoftKeys + HardKeys)
            { group: "Programmable Buttons (LineKeys)", items: Array.from({ length: (currentModel?.own_soft_keys || 0) + (currentModel?.own_hard_keys || 0) || 10 }).flatMap((_, i) => {
                const btnIndex = i + 1;
                return (vendor?.features?.filter((f: any) => f.associated_with_button) || []).flatMap((f: any) => 
                    f.params.map((p: any) => ({
                        label: `Button ${btnIndex}: ${f.name} - ${p.label}`,
                        value: `button.${btnIndex}.${f.id}.${p.id}`
                    }))
                );
            })},
    
            // Expansion Modules
            ...Array.from({ length: maxPanels }).map((_, pIdx) => {
                const panel = pIdx + 1;
                return {
                    group: `Expansion Module ${panel}`,
                    items: Array.from({ length: maxKeysPerPanel }).flatMap((_, kIdx) => {
                        const key = kIdx + 1;
                        return (vendor?.features?.filter((f: any) => f.associated_with_button) || []).flatMap((f: any) => 
                            f.params.map((p: any) => ({
                                label: `Key ${key}: ${f.name} - ${p.label}`,
                                value: `button.ext.${panel}.${key}.${f.id}.${p.id}`
                            }))
                        );
                    })
                };
            })
        ];
        
        flatMappedOptions = MAPPING_OPTIONS.flatMap(g => g.items);
    }

    let currentIndex = 0;
    let maxIngestedIndex = 0; // The furthest file we've sent to engine.ingestConfig
    let currentFileName: string = "";
    let currentDiffs: Diff[] = [];
    let currentConflicts: any[] = [];
    let currentExtracted: Record<string, string> = {};
    let isComplete = false;
    let isPlaying = false;
    let autoSkip = true;
    let showMappingTable = false;

    // Shared Selector State
    let showSelector = false;
    let selectorTargetKey: string | null = null;
    let selectorSearch = "";
    
    $: filteredOptions = selectorSearch.trim() === "" 
        ? MAPPING_OPTIONS 
        : MAPPING_OPTIONS.map(group => ({
            ...group,
            items: group.items.filter(item => 
                item.label.toLowerCase().includes(selectorSearch.toLowerCase()) ||
                item.value.toLowerCase().includes(selectorSearch.toLowerCase())
            )
        })).filter(group => group.items.length > 0);

    function openSelector(key: string) {
        selectorTargetKey = key;
        selectorSearch = "";
        showSelector = true;
    }

    function selectMapping(value: string) {
        if (selectorTargetKey) {
            handleMap(selectorTargetKey, value);
        }
        showSelector = false;
        selectorTargetKey = null;
    }

    function getMappingLabel(value: string) {
        if (!value) return "Unmapped";
        if (value === "ignore") return "Ignored";
        const found = flatMappedOptions.find(opt => opt.value === value);
        return found ? found.label : value;
    }

    let editingKeyId: string | null = null;
    let editBuffer: string = "";

    // Helper to get preview data for the ribbon
    function getFileSummary(index: number) {
        if (index > maxIngestedIndex) return null;
        const file = files[index];
        const baseConfig = ConfigParser.parse(file.content, file.name);
        const augmentedConfigs = engine.getAllConfigs();
        const config = augmentedConfigs[index] || baseConfig;
        const data = engine.getExtractedData(config);
        return data["phone.phone_number"] || data["phone.mac_address"] || null;
    }

    function updatePreview() {
        if (currentIndex < files.length) {
            const file = files[currentIndex];
            const baseConfig = ConfigParser.parse(file.content, file.name);
            const augmentedConfigs = engine.getAllConfigs();
            const config = augmentedConfigs[currentIndex] || baseConfig;
            currentExtracted = engine.getExtractedData(config);
        }
    }

    async function processFile() {
        if (currentIndex >= files.length) {
            isComplete = true;
            isPlaying = false;
            return;
        }

        const file = files[currentIndex];
        currentFileName = file.name;
        
        const baseConfig = ConfigParser.parse(file.content, file.name);
        const augmentedConfigs = engine.getAllConfigs();
        const config = augmentedConfigs[currentIndex] || baseConfig;

        // Files are already ingested in +page.svelte before entering discovery
        maxIngestedIndex = files.length - 1;

        const result = engine.analyzeConfig(config, file.content);
        
        currentDiffs = [
            ...result.differences,
            ...result.newKeys.map(key => ({
                key,
                valueA: "---",
                valueB: config[key]
            }))
        ];

        currentConflicts = result.conflicts;
        updatePreview();

        // Check if we can auto-advance
        const allMapped = currentDiffs.every(
            (d) =>
                engine.isKeyMapped(d.key) ||
                engine.isKeyIgnored(d.key) ||
                engine
                    .getAnalysisSummary()
                    .some((r) => r.key === d.key && r.type === "global"),
        );
        const hasConflicts = currentConflicts.length > 0;

        if (allMapped && !hasConflicts && (isPlaying || (autoSkip && currentIndex >= maxIngestedIndex))) {
            // Wait a tiny bit so user sees it moving
            if (isPlaying) await new Promise((r) => setTimeout(r, 50));
            currentIndex++;
            processFile();
        } else {
            // Stop playing if we hit a problem or user intervention is needed
            isPlaying = false;
        }
    }

    function goToPrevious() {
        if (currentIndex > 0) {
            currentIndex--;
            isPlaying = false;
            processFile();
        }
    }

    function goToNext() {
        if (currentIndex < files.length - 1) {
            currentIndex++;
            processFile();
        } else if (currentIndex === files.length - 1) {
            isComplete = true;
        }
    }

    function handleSkip(index: number) {
        const file = files[index];
        engine.removeConfig(file.name);
        
        // Remove from local files array
        files = files.filter((_, i) => i !== index);
        
        // Adjust indices
        if (maxIngestedIndex >= index) {
            maxIngestedIndex--;
        }
        
        if (currentIndex >= files.length && files.length > 0) {
            currentIndex = files.length - 1;
        } else if (files.length === 0) {
            isComplete = true; // Nothing left
            return;
        }
        
        processFile();
        toast.info(`Skipped ${file.name}`);
        dispatch('filesChanged', { files });
    }

    function togglePlay() {
        isPlaying = !isPlaying;
        if (isPlaying) {
            processFile();
        }
    }

    function handleMap(key: string, systemField: string) {
        if (systemField === "ignore") {
            engine.ignoreKey(key);
        } else {
            engine.mapField(key, systemField);
        }
        engine = engine;

        // Refresh analysis for current file based on updated engine mapping
        const file = files[currentIndex];
        const baseConfig = ConfigParser.parse(file.content, file.name);
        const augmentedConfigs = engine.getAllConfigs();
        const config = augmentedConfigs[currentIndex] || baseConfig;
        
        const result = engine.analyzeConfig(config, file.content);
        
        currentDiffs = [
            ...result.differences,
            ...result.newKeys.map(key => ({
                key,
                valueA: "---",
                valueB: config[key]
            }))
        ];

        currentConflicts = result.conflicts;
        updatePreview();
    }

    function startEditKey(key: string) {
        editingKeyId = key;
        editBuffer = key;
    }

    function handleRename() {
        if (!editingKeyId || !editBuffer || editingKeyId === editBuffer) {
            editingKeyId = null;
            return;
        }

        const oldKey = editingKeyId;
        const newKey = editBuffer;

        engine.renameKey(oldKey, newKey);
        engine = engine;
        editingKeyId = null;

        // Refresh current
        const file = files[currentIndex];
        const config = ConfigParser.parse(file.content, file.name);
        const result = engine.analyzeConfig(config, file.content);
        
        currentDiffs = [
            ...result.differences,
            ...result.newKeys.map(key => ({
                key,
                valueA: "---",
                valueB: config[key]
            }))
        ];

        currentConflicts = result.conflicts;
        updatePreview();

        toast.success(`Key renamed from "${oldKey}" to "${newKey}"`);
    }

    onMount(() => {
        processFile();
    });

    // Reset completion if new files are added
    $: if (files.length > currentIndex && isComplete) {
        isComplete = false;
        // Don't auto-process here, let the user trigger it or let it happen naturally
    }

    $: progress = Math.round((currentIndex / files.length) * 100);
    $: mappedFields = engine ? engine.getMappedFields() : {};
</script>

<div class="space-y-6">
    <!-- Header with Ribbon -->
    <div
        class="card bg-card shadow-md border border-border overflow-hidden"
    >
        <div
            class="bg-muted p-2 flex items-center justify-between border-b border-border"
        >
            <div class="flex items-center gap-2 px-2">
                <ListFilter class="w-4 h-4 opacity-50" />
                <span
                    class="text-[10px] font-bold uppercase tracking-wider opacity-50"
                    >File Pipeline</span
                >
            </div>
            <div class="text-[10px] font-mono opacity-50 pr-2">
                {currentIndex + 1} / {files.length}
            </div>
        </div>

        <!-- Navigation Ribbon (Scrollable) -->
        <div
            class="flex overflow-x-auto p-3 gap-2 scrollbar-none bg-background/50"
        >
            {#each files as file, i}
                <button
                    class="flex-shrink-0 flex flex-col items-start p-2 rounded-lg border transition-all min-w-[140px] text-left"
                    class:border-primary={i === currentIndex}
                    class:bg-active={i === currentIndex}
                    class:border-border={i !== currentIndex}
                    class:opacity-40={i > maxIngestedIndex}
                    on:click={() => {
                        currentIndex = i;
                        isPlaying = false;
                        isComplete = false;
                        processFile();
                    }}
                >
                    <div
                        class="text-[10px] font-mono truncate w-full mb-1"
                        title={file.name}
                    >
                        {file.name}
                    </div>
                    <div class="h-4 flex items-center justify-between w-full mt-1">
                        {#if i <= maxIngestedIndex}
                            <span
                                class="text-[8px] opacity-60 truncate max-w-[80px]"
                            >
                                {availableModels.find(m => m.id === engine.getExtractedData(ConfigParser.parse(files[i].content, files[i].name), files[i].content)["phone.model_id"])?.name || "---"}
                            </span>
                        {/if}
                    </div>
                    <div class="h-4 flex items-center justify-between w-full">
                        {#if i <= maxIngestedIndex}
                            <span
                                class="text-[9px] font-bold text-primary truncate max-w-[80px]"
                            >
                                {getFileSummary(i) || "---"}
                            </span>
                        {:else}
                            <span class="text-[8px] italic opacity-40"
                                >Pending...</span
                            >
                        {/if}
                        
                        <button 
                            class="hover:text-error transition-colors p-0.5"
                            on:click|stopPropagation={() => handleSkip(i)}
                            title="Skip this file"
                        >
                            <X class="w-3 h-3" />
                        </button>
                    </div>
                </button>
            {/each}
        </div>

        <!-- Controls Row -->
        <div
            class="p-3 bg-muted/50 flex items-center justify-between border-t border-border"
        >
            <div class="flex items-center gap-2">
                <button
                    class="btn btn-sm btn-circle btn-ghost"
                    disabled={currentIndex === 0}
                    on:click={goToPrevious}
                >
                    <ChevronLeft class="w-5 h-5" />
                </button>

                <div
                    class="flex items-center bg-background rounded-full border border-border overflow-hidden"
                >
                    <button
                        class="btn btn-sm px-4 gap-2 border-none rounded-none hover:bg-primary/10"
                        on:click={togglePlay}
                        class:btn-error={isPlaying}
                        class:text-error={isPlaying}
                    >
                        {#if isPlaying}
                            <Square class="w-3 h-3 fill-current" />
                            <span class="text-[10px] font-bold">STOP</span>
                        {:else}
                            <Play class="w-3 h-3 fill-current" />
                            <span class="text-[10px] font-bold">PLAY</span>
                        {/if}
                    </button>
                </div>

                <button
                    class="btn btn-sm btn-circle btn-ghost"
                    disabled={currentIndex >= files.length - 1 ||
                        currentConflicts.length > 0}
                    on:click={goToNext}
                >
                    <ChevronRight class="w-5 h-5" />
                </button>

                <div class="divider divider-horizontal mx-0"></div>

                <button
                    class="btn btn-xs btn-outline btn-primary gap-1"
                    on:click={() => dispatch("requestAdd")}
                >
                    <Plus class="w-3 h-3" />
                    <span>Add Files</span>
                </button>
            </div>

            <div class="flex items-center gap-4">
                <label class="label cursor-pointer gap-2 py-0">
                    <span
                        class="label-text text-[10px] font-bold uppercase opacity-50"
                        >Auto-skip</span
                    >
                    <input
                        type="checkbox"
                        bind:checked={autoSkip}
                        class="checkbox checkbox-primary checkbox-xs"
                    />
                </label>

                <button
                    class="btn btn-xs btn-outline gap-2"
                    on:click={() => (showMappingTable = !showMappingTable)}
                >
                    <Table class="w-3 h-3" />
                    {showMappingTable ? "Hide Mappings" : "Show Mappings"}
                </button>
            </div>
        </div>
    </div>

    {#if !isComplete}
        <div class="grid grid-cols-1 xl:grid-cols-4 gap-6">
            <div class="xl:col-span-3 space-y-4">
                {#if currentConflicts.length > 0}
                    <div
                        class="alert alert-error shadow-lg animate-in slide-in-from-top-4"
                    >
                        <AlertCircle class="w-6 h-6" />
                        <div>
                            <h3 class="font-bold text-sm uppercase">
                                Mapping Conflict!
                            </h3>
                            <div class="text-xs opacity-90">
                                Multiple keys are mapping to the same field but
                                have different values in this file:
                                <ul
                                    class="list-disc list-inside mt-2 font-mono"
                                >
                                    {#each currentConflicts as conflict}
                                        <li
                                            class="bg-error/20 p-2 rounded mb-1"
                                        >
                                            <strong>{conflict.field}</strong>:
                                            {#each conflict.contributors as c, i}
                                                {c.key} ({c.value}){i <
                                                conflict.contributors.length - 1
                                                    ? " vs "
                                                    : ""}
                                            {/each}
                                        </li>
                                    {/each}
                                </ul>
                            </div>
                        </div>
                    </div>
                {/if}

                <div class="card bg-card shadow-xl border border-border">
                    <div class="card-body">
                        <div class="flex items-center gap-2 mb-4">
                            <AlertCircle class="w-5 h-5 text-warning" />
                            <h2 class="card-title">
                                Discovery: {currentFileName}
                            </h2>
                            <div class="flex-1"></div>
                            <button 
                                class="btn btn-xs btn-ghost text-error gap-1"
                                on:click={() => handleSkip(currentIndex)}
                            >
                                <Trash2 class="w-3 h-3" />
                                Skip this Phone
                            </button>
                        </div>

                        {#if currentDiffs.length > 0}
                            <p class="mb-4 text-muted-foreground">
                                I found some fields that are different from the
                                baseline. Please help me identify them:
                            </p>

                            <div class="overflow-x-auto">
                                <table
                                    class="table table-zebra w-full border border-border"
                                >
                                    <thead class="bg-muted">
                                        <tr>
                                            <th
                                                class="w-1/3 text-primary bg-muted"
                                                >Extraction Pattern</th
                                            >
                                            <th class="bg-muted">Baseline</th
                                            >
                                            <th class="bg-muted">Current</th>
                                            <th
                                                class="w-1/4 text-secondary bg-muted"
                                                >Mapping</th
                                            >
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {#each currentDiffs as diff (diff.key)}
                                            {@const hasConflict =
                                                currentConflicts.some((c) =>
                                                    c.contributors.some(
                                                        (ct) =>
                                                            ct.key === diff.key,
                                                    ),
                                                )}
                                            <tr
                                                class="hover"
                                                class:bg-error={hasConflict}
                                            >
                                                <td>
                                                    <div
                                                        class="flex items-center gap-2"
                                                    >
                                                        {#if editingKeyId === diff.key}
                                                            <div class="flex items-center gap-1 w-full">
                                                                <input 
                                                                    type="text" 
                                                                    bind:value={editBuffer}
                                                                    class="input input-xs input-bordered flex-1 font-mono text-xs"
                                                                    on:keydown={(e) => e.key === 'Enter' && handleRename()}
                                                                    on:blur={handleRename}
                                                                    autoFocus
                                                                />
                                                                <button class="btn btn-xs btn-ghost text-error" on:click={() => editingKeyId = null}>
                                                                    <X class="w-3 h-3" />
                                                                </button>
                                                            </div>
                                                        {:else}
                                                                <button 
                                                                class="flex items-center gap-2 group/key hover:opacity-100 transition-all text-left"
                                                                on:click={() => startEditKey(diff.key)}
                                                                title="Click to refine pattern"
                                                            >
                                                                 <code
                                                                     class="text-primary dark:text-orange-400 font-bold text-[10px] bg-muted/50 dark:bg-orange-950/30 px-2 py-0.5 rounded group-hover/key:ring-1 ring-primary break-all border border-transparent dark:border-orange-900/50"
                                                                     >{diff.key}</code
                                                                 >
                                                                <Edit3 class="w-3 h-3 opacity-0 group-hover/key:opacity-40 transition-opacity" />
                                                            </button>
                                                        {/if}

                                                        <div class="flex items-center gap-1">
                                                            {#if hasConflict}
                                                                <span
                                                                    class="badge badge-error badge-xs animate-pulse uppercase text-[8px]"
                                                                    >Conflict</span
                                                                >
                                                            {:else if engine.isKeyIgnored(diff.key)}
                                                                <span
                                                                    class="badge badge-ghost badge-xs opacity-50 uppercase text-[8px]"
                                                                    >Ignored</span
                                                                >
                                                            {:else if !engine.isKeyMapped(diff.key)}
                                                                <span
                                                                    class="badge badge-error badge-xs uppercase text-[8px]"
                                                                    >Unmapped</span
                                                                >
                                                            {:else}
                                                                <span
                                                                    class="badge badge-success badge-xs uppercase text-[8px]"
                                                                    >Mapped</span
                                                                >
                                                            {/if}
                                                        </div>
                                                    </div>
                                                </td>
                                                <td>
                                                    <span
                                                                class="badge badge-outline badge-sm opacity-70 border-border text-foreground"
                                                        >{diff.valueA}</span
                                                    >
                                                </td>
                                                <td>
                                                    <span
                                                        class="badge badge-primary badge-sm font-mono"
                                                        >{diff.valueB}</span
                                                    >
                                                </td>
                                                <td>
                                                    <button
                                                        class="btn btn-xs w-full justify-between font-normal text-left truncate"
                                                        class:btn-outline={!engine.getMappedFields()[diff.key]}
                                                        class:btn-ghost={engine.getMappedFields()[diff.key] === 'ignore'}
                                                        class:btn-primary={engine.getMappedFields()[diff.key] && engine.getMappedFields()[diff.key] !== 'ignore'}
                                                        on:click={() => openSelector(diff.key)}
                                                    >
                                                        <span class="truncate">
                                                            {getMappingLabel(engine.getMappedFields()[diff.key])}
                                                        </span>
                                                        <Edit3 class="w-2.5 h-2.5 opacity-50 shrink-0" />
                                                    </button>
                                                </td>
                                            </tr>
                                        {/each}
                                    </tbody>
                                </table>
                            </div>
                        {:else}
                            <div class="text-center py-8">
                                <CheckCircle2
                                    class="w-12 h-12 text-success mx-auto mb-4"
                                />
                                <p class="font-semibold text-lg">
                                    No New Fields Found
                                </p>
                                <p class="text-sm opacity-60">
                                    This file is already compatible with your
                                    current mapping.
                                </p>

                                {#if currentConflicts.length === 0}
                                    <button
                                        class="btn btn-primary btn-wide mt-6"
                                        on:click={goToNext}
                                    >
                                        Next File <ChevronRight
                                            class="w-4 h-4 ml-2"
                                        />
                                    </button>
                                {:else}
                                    <div class="alert alert-error mt-6 text-sm">
                                        <AlertCircle class="w-4 h-4" />
                                        <span
                                            >Resolve conflicts in the preview
                                            panel to proceed.</span
                                        >
                                    </div>
                                {/if}
                            </div>
                        {/if}
                    </div>
                </div>
            </div>
            <!-- End of xl:col-span-3/4 -->

            <!-- Side Panel: Mappings & Preview -->
            <div class="xl:col-span-1 space-y-4">
                {#if showMappingTable}
                    <div
                        class="card bg-card border border-border shadow-lg animate-in slide-in-from-right-4 transition-all"
                    >
                        <div class="card-body p-4">
                            <h3
                                class="font-bold text-xs uppercase opacity-50 flex items-center gap-2 mb-2"
                            >
                                <Table class="w-3 h-3" />
                                Active Mappings
                            </h3>
                            <div class="max-h-[300px] overflow-y-auto">
                                <table class="table table-xs w-full">
                                    <thead>
                                        <tr>
                                            <th class="px-0">Config Key</th>
                                            <th class="px-0">System Field</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {#each Object.entries(mappedFields) as [key, field]}
                                            <tr>
                                                <td class="px-0"
                                                    ><code
                                                        class="text-[10px] opacity-70"
                                                        >{key}</code
                                                    ></td
                                                >
                                                <td class="px-0"
                                                    ><span
                                                        class="badge badge-ghost badge-sm text-[10px]"
                                                        >{field}</span
                                                    ></td
                                                >
                                            </tr>
                                        {/each}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                {/if}

                <!-- Preview Card -->
                <div
                    class="card bg-muted border border-border shadow-lg h-fit sticky top-4"
                >
                    <div class="card-body p-4">
                        <h3
                            class="font-bold flex items-center gap-2 border-b border-border pb-2 mb-4"
                        >
                            <CheckCircle2 class="w-4 h-4 text-success" />
                            Resulting Phone Record
                        </h3>

                        <div class="space-y-4 text-sm">
                            <div
                                class="bg-background p-3 rounded-lg border border-border"
                            >
                                <div
                                    class="text-[10px] opacity-50 uppercase font-bold mb-2"
                                >
                                    Phone Identity
                                </div>
                                <div class="flex justify-between mb-1">
                                    <span class="opacity-70">MAC:</span>
                                    <span class="font-mono text-primary"
                                        >{currentExtracted[
                                            "phone.mac_address"
                                        ] || "---"}</span
                                    >
                                </div>
                                <div class="flex justify-between">
                                    <span class="opacity-70">Number:</span>
                                    <span class="font-bold"
                                        >{currentExtracted[
                                            "phone.phone_number"
                                        ] || "---"}</span
                                    >
                                </div>
                            </div>

                            {#each [0, 1, 2, 3, 4, 5] as i}
                                {#if currentExtracted[`lines[${i}].user_name`] || currentExtracted[`lines[${i}].password`] || currentExtracted[`lines[${i}].label`]}
                                    <div
                                        class="bg-background p-3 rounded-lg border border-border border-l-4 border-l-secondary"
                                    >
                                        <div
                                            class="text-[10px] opacity-50 uppercase font-bold mb-2"
                                        >
                                            Line {i + 1}
                                        </div>
                                        <div class="flex justify-between mb-1">
                                            <span class="opacity-70">User:</span
                                            >
                                            <span class="font-mono"
                                                >{currentExtracted[
                                                    `lines[${i}].user_name`
                                                ] || "---"}</span
                                            >
                                        </div>
                                        <div class="flex justify-between mb-1">
                                            <span class="opacity-70"
                                                >Label:</span
                                            >
                                            <span class="font-bold"
                                                >{currentExtracted[
                                                    `lines[${i}].label`
                                                ] || "---"}</span
                                            >
                                        </div>
                                        <div class="flex justify-between">
                                            <span class="opacity-70"
                                                >Password:</span
                                            >
                                            <span class="font-mono"
                                                >{currentExtracted[
                                                    `lines[${i}].password`
                                                ]
                                                    ? "***"
                                                    : "---"}</span
                                            >
                                        </div>
                                    </div>
                                {/if}
                            {/each}

                            {#if !Object.values(currentExtracted).some((v) => v)}
                                <div class="text-center py-8 opacity-30 italic">
                                    No data extracted yet. Please map some
                                    fields.
                                </div>
                            {/if}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    {:else}
        <div
            class="card bg-success/10 border border-success/30 text-center p-12"
        >
            <CheckCircle2 class="w-16 h-16 text-success mx-auto mb-4" />
            <h2 class="text-2xl font-bold mb-2">Analysis Complete!</h2>
            <p class="mb-4">
                All {files.length} files have been processed. I've extracted the
                data based on your mapping.
            </p>

            <div class="flex justify-center gap-4 mt-8">
                <button
                    class="btn btn-outline gap-2"
                    on:click={() => dispatch("requestAdd")}
                >
                    <Plus class="w-4 h-4 mr-2" />
                    Add More Files
                </button>
                <button
                    class="btn btn-ghost gap-2 text-error"
                    on:click={() => dispatch("addMore")}
                >
                    <Trash2 class="w-4 h-4 mr-2" />
                    Reset & Start Over
                </button>
                <button
                    class="btn btn-success px-8"
                    on:click={() => dispatch("complete")}
                >
                    Review Extracted Data
                    <ArrowRight class="w-4 h-4 ml-2" />
                </button>
            </div>
        </div>
    {/if}
</div>

<!-- Mapping Selector Modal -->
{#if showSelector}
    <div class="fixed inset-0 z-[60] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
        <div class="bg-card border shadow-2xl rounded-xl w-full max-w-2xl max-h-[80vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
            <!-- Modal Header -->
            <div class="px-6 py-4 border-b flex items-center justify-between bg-muted/30">
                <div>
                    <h3 class="font-bold text-lg">Select Mapping</h3>
                    <p class="text-xs opacity-60">Mapping for pattern: <code class="bg-muted px-1 rounded">{selectorTargetKey}</code></p>
                </div>
                <button class="btn btn-sm btn-ghost btn-circle" on:click={() => showSelector = false}>
                    <X class="w-5 h-5" />
                </button>
            </div>

            <!-- Search Bar -->
            <div class="p-4 border-b">
                <div class="relative">
                    <ListFilter class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 opacity-50" />
                    <input 
                        type="text" 
                        bind:value={selectorSearch}
                        placeholder="Search keys, features, lines... (e.g. 'exp 1 key 5' or 'label')"
                        class="input input-bordered w-full pl-10 focus:ring-2 ring-primary/20"
                        on:keydown={(e) => e.key === 'Escape' && (showSelector = false)}
                    />
                </div>
            </div>

            <!-- Options List -->
            <div class="flex-1 overflow-y-auto p-2 scrollbar-thin">
                {#if filteredOptions.length === 0}
                    <div class="text-center py-12 opacity-30 italic">
                        No matches found for "{selectorSearch}"
                    </div>
                {:else}
                    {#each filteredOptions as group}
                        <div class="mb-4">
                            <div class="px-2 py-1 text-[10px] font-bold uppercase tracking-wider opacity-40 bg-muted/50 rounded mb-1">
                                {group.group}
                            </div>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-1 px-1">
                                {#each group.items as opt}
                                    <button 
                                        class="flex items-center justify-between px-3 py-2 text-sm rounded-lg hover:bg-primary hover:text-white transition-colors text-left group/opt"
                                        on:click={() => selectMapping(opt.value)}
                                    >
                                        <span class="truncate">{opt.label}</span>
                                        <code class="text-[10px] opacity-40 group-hover/opt:text-white/70 truncate max-w-[120px] ml-2">{opt.value}</code>
                                    </button>
                                {/each}
                            </div>
                        </div>
                    {/each}
                {/if}
            </div>

            <!-- Footer -->
            <div class="px-6 py-3 border-t bg-muted/10 flex justify-end gap-2">
                <button class="btn btn-sm btn-ghost" on:click={() => showSelector = false}>Cancel</button>
                <button class="btn btn-sm btn-error btn-outline" on:click={() => selectMapping('ignore')}>Ignore Pattern</button>
            </div>
        </div>
    </div>
{/if}

<style>
    .scrollbar-thin::-webkit-scrollbar {
        width: 6px;
    }
    .scrollbar-thin::-webkit-scrollbar-track {
        background: transparent;
    }
    .scrollbar-thin::-webkit-scrollbar-thumb {
        background: rgba(0, 0, 0, 0.1);
        border-radius: 10px;
    }
    :global(.dark) .scrollbar-thin::-webkit-scrollbar-thumb {
        background: rgba(255, 255, 255, 0.1);
    }
</style>
