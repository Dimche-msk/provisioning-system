<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { type DiscoveryEngine } from "$lib/migration/engine";
    import {
        Save,
        FileCheck,
        Search,
        ArrowUpDown,
        CheckCircle2,
        XCircle,
        Clock,
        Play,
        Trash2,
        Edit3,
        Check,
        X,
        ArrowLeft,
        Copy,
        ChevronRight,
        Settings2,
        Cpu,
        Library,
        Layers,
        Info,
        Square,
        CheckSquare,
        BarChart3,
        Table,
        ChevronDown,
        ChevronUp,
        Plus,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";

    export let engine: DiscoveryEngine;
    export let vendor: string;
    export let domain: string;

    const dispatch = createEventDispatcher();

    // -- State --
    type MigrationRecord = {
        data: Record<string, string>;
        status: "pending" | "processing" | "success" | "error";
        error?: string;
        isSelected: boolean;
        id: string; // Filename usually
    };

    let activeTab: "audit" | "import" = "audit";
    let records: MigrationRecord[] = [];
    let selectedRecordIndex = 0;
    let searchQuery = "";
    let isProcessing = false;
    let progress = 0;

    // Recommendations state (Restored)
    let localRecommendations = engine.getAnalysisSummary();
    let editingId: number | null = null;
    let editBuffer = { key: "", value: "" };

    // Sorting for recommendations
    let recSortField: "key" | "type" | "isApproved" = "key";
    let recSortDir: 1 | -1 = 1;

    $: sortedRecommendations = [...localRecommendations].sort((a, b) => {
        const valA = a[recSortField];
        const valB = b[recSortField];
        if (valA < valB) return -1 * recSortDir;
        if (valA > valB) return 1 * recSortDir;
        return 0;
    });

    // Recommendation summary for global settings
    $: globalData = localRecommendations
        .filter((r) => r.type === "global" && r.isApproved)
        .reduce((acc, r) => {
            acc[r.key] = r.value || "";
            return acc;
        }, {} as Record<string, string>);

    $: globalConfigText = localRecommendations
        .filter((r) => r.type === "global" && r.isApproved)
        .map((r) => `${r.key} = ${r.value}`)
        .join("\n");

    onMount(() => {
        const rawData = engine.getAllExtractedData();
        records = rawData.map((row) => ({
            data: row,
            status: "pending",
            isSelected: true,
            id: row["phone.raw_filename"],
        }));
    });

    // -- Sync with Engine --
    $: {
        engine.setSavedRecommendations(localRecommendations);
    }

    // -- Derived --
    $: filteredRecords = records.filter((r) => {
        const query = searchQuery.toLowerCase();
        return (
            r.data["phone.mac_address"]?.toLowerCase().includes(query) ||
            r.data["phone.phone_number"]?.toLowerCase().includes(query) ||
            r.id.toLowerCase().includes(query)
        );
    });

    $: selectedRecord = filteredRecords[selectedRecordIndex] || null;
    $: stats = {
        total: records.length,
        selected: records.filter((r) => r.isSelected).length,
        success: records.filter((r) => r.status === "success").length,
        error: records.filter((r) => r.status === "error").length,
    };

    // -- Actions (Audit) --
    function toggleRecSort(field: typeof recSortField) {
        if (recSortField === field) {
            recSortDir *= -1;
        } else {
            recSortField = field;
            recSortDir = 1;
        }
    }

    function startEditRec(rec: any) {
        editingId = rec.id;
        editBuffer = { key: rec.key, value: rec.value || "" };
    }

    function saveEditRec() {
        localRecommendations = localRecommendations.map((r) =>
            r.id === editingId
                ? { ...r, key: editBuffer.key, value: editBuffer.value }
                : r,
        );
        editingId = null;
    }

    function deleteRec(id: number) {
        localRecommendations = localRecommendations.filter((r) => r.id !== id);
    }

    function toggleApprove(id: number) {
        localRecommendations = localRecommendations.map((r) =>
            r.id === id ? { ...r, isApproved: !r.isApproved } : r,
        );
    }

    function toggleType(id: number) {
        localRecommendations = localRecommendations.map((r) =>
            r.id === id
                ? { ...r, type: r.type === "global" ? "dynamic" : "global" }
                : r,
        );
    }

    function addNewParameter() {
        const newId = Math.max(0, ...localRecommendations.map((r) => r.id)) + 1;
        const newRec = {
            id: newId,
            key: "new_key",
            type: "global" as const,
            value: "",
            coverage: 1,
            isApproved: true,
        };
        localRecommendations = [...localRecommendations, newRec];
        startEditRec(newRec);
    }

    // -- Actions (Import) --
    function toggleSelection(record: MigrationRecord) {
        record.isSelected = !record.isSelected;
        records = [...records];
    }

    function toggleAll() {
        const allSelected = stats.selected === filteredRecords.length;
        filteredRecords.forEach((r) => (r.isSelected = !allSelected));
        records = [...records];
    }

    async function processSelected() {
        const toProcess = records.filter((r) => r.isSelected && r.status !== "success");
        if (toProcess.length === 0) {
            toast.info("No records selected for import");
            return;
        }

        isProcessing = true;
        progress = 0;
        let completed = 0;

        for (const record of toProcess) {
            record.status = "processing";
            records = [...records];

            try {
                const response = await fetch("/api/migration/apply", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        domain,
                        vendor,
                        model_id: record.data["phone.model_id"],
                        data: record.data,
                        global_data: globalData,
                    }),
                });

                if (response.status === 201) {
                    record.status = "success";
                    record.isSelected = false; // Auto-uncheck on success
                } else {
                    const errorText = await response.text();
                    record.status = "error";
                    record.error = errorText;
                }
            } catch (err: any) {
                record.status = "error";
                record.error = err.message;
            }

            completed++;
            progress = (completed / toProcess.length) * 100;
            records = [...records];
        }

        isProcessing = false;
        if (stats.error > 0) {
            toast.error(`Finished with ${stats.error} errors`);
        } else {
            toast.success("Batch import completed!");
        }
    }

    async function copyPreview() {
        if (!globalConfigText) return;
        try {
            await navigator.clipboard.writeText(globalConfigText);
            toast.success("Common configuration copied");
        } catch (err) {
            toast.error("Failed to copy configuration");
        }
    }

    function removeRecord(id: string) {
        records = records.filter((r) => r.id !== id);
        if (selectedRecordIndex >= records.length) {
            selectedRecordIndex = Math.max(0, records.length - 1);
        }
    }

    // Helper functions (same as before)
    function getLineData(data: Record<string, string>) {
        const lines: any[] = [];
        for (let i = 0; i < 20; i++) {
            const prefix = `lines[${i}]`;
            if (data[`${prefix}.user_name`] || data[`${prefix}.label`]) {
                lines.push({
                    index: i,
                    user: data[`${prefix}.user_name`],
                    label: data[`${prefix}.label`],
                    auth: data[`${prefix}.auth_name`],
                });
            }
        }
        return lines;
    }

    function getButtonData(data: Record<string, string>) {
        const btns: any[] = [];
        const btnKeys = Object.keys(data).filter(k => k.startsWith('button.'));
        const indices = new Set(btnKeys.map(k => k.split('.')[1]));
        
        indices.forEach(idx => {
            const typeKey = btnKeys.find(k => k.startsWith(`button.${idx}.`) && !k.endsWith('.value') && !k.endsWith('.label'));
            const type = typeKey ? typeKey.split('.')[2] : 'unknown';
            btns.push({
                index: idx,
                type,
                value: data[`button.${idx}.${type}.value`] || data[`button.${idx}.${type}.number`] || 'N/A',
                label: data[`button.${idx}.${type}.label`] || 'N/A'
            });
        });
        return btns.sort((a, b) => Number(a.index) - Number(b.index));
    }
</script>

<div class="flex flex-col h-[850px] border border-border rounded-2xl bg-card overflow-hidden shadow-2xl animate-in fade-in zoom-in-95 duration-300">
    <!-- Header Area -->
    <div class="p-6 border-b bg-muted/20 flex items-center justify-between gap-6 shrink-0">
        <div class="flex items-center gap-4">
            <button class="btn btn-ghost btn-sm" on:click={() => dispatch("back")} disabled={isProcessing}>
                <ArrowLeft class="w-4 h-4" />
            </button>
            <div>
                <h2 class="text-xl font-bold flex items-center gap-2">
                    <FileCheck class="w-6 h-6 text-primary" />
                    Review & Import
                </h2>
                <div class="flex items-center gap-4 mt-1">
                    <div class="tabs tabs-boxed bg-background border p-0.5">
                        <button 
                            class="tab tab-sm gap-1 {activeTab === 'audit' ? 'tab-active' : ''}" 
                            on:click={() => activeTab = 'audit'}
                            disabled={isProcessing}
                        >
                            <BarChart3 class="w-3 h-3" /> Audit & Strategy
                        </button>
                        <button 
                            class="tab tab-sm gap-1 {activeTab === 'import' ? 'tab-active' : ''}" 
                            on:click={() => activeTab = 'import'}
                        >
                            <Play class="w-3 h-3" /> Device Explorer
                        </button>
                    </div>
                </div>
            </div>
        </div>

        {#if isProcessing}
            <div class="flex-1 max-w-md space-y-2">
                <div class="flex justify-between text-xs font-medium">
                    <span>Processing {stats.success + stats.error + 1} of {stats.selected}...</span>
                    <span>{Math.round(progress)}%</span>
                </div>
                <div class="w-full bg-secondary h-2 rounded-full overflow-hidden">
                    <div class="bg-primary h-full transition-all duration-300" style="width: {progress}%"></div>
                </div>
            </div>
        {:else if activeTab === 'import'}
            <div class="flex items-center gap-2">
                <div class="grid grid-cols-3 gap-1 px-4 py-2 bg-background/50 rounded-lg border text-[10px] font-mono">
                    <div class="flex flex-col border-r pr-2">
                        <span class="opacity-50">SELECTED</span>
                        <span class="font-bold text-sm">{stats.selected}</span>
                    </div>
                    <div class="flex flex-col border-r px-2">
                        <span class="opacity-50">SUCCESS</span>
                        <span class="font-bold text-sm text-success">{stats.success}</span>
                    </div>
                    <div class="flex flex-col pl-2">
                        <span class="opacity-50">ERRORS</span>
                        <span class="font-bold text-sm text-error">{stats.error}</span>
                    </div>
                </div>
                <button 
                    class="btn btn-primary shadow-lg shadow-primary/20" 
                    on:click={processSelected}
                    disabled={stats.selected === 0}
                >
                    <Play class="w-4 h-4 mr-2" />
                    Start Insertion
                </button>
            </div>
        {/if}
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col overflow-hidden">
        {#if activeTab === 'audit'}
            <!-- Tab 1: Audit & Strategy -->
            <div class="flex-1 flex overflow-hidden">
                <!-- Recommendations Table -->
                <div class="flex-1 p-6 overflow-y-auto custom-scrollbar space-y-6">
                    <div class="card border bg-background/50 overflow-hidden">
                        <div class="p-4 border-b bg-muted/30 flex items-center justify-between">
                            <div class="flex items-center gap-2">
                                <BarChart3 class="w-4 h-4 text-primary" />
                                <h4 class="font-bold text-sm">Discovered Template Parameters</h4>
                            </div>
                            <div class="flex items-center gap-2">
                                <button class="btn btn-xs btn-primary gap-1" on:click={addNewParameter}>
                                    <Plus class="w-3 h-3" /> New parameter
                                </button>
                                <span class="badge badge-sm badge-outline opacity-50">{localRecommendations.length} parameters found</span>
                            </div>
                        </div>
                        <div class="overflow-x-auto">
                            <table class="table table-sm table-zebra w-full">
                                <thead class="bg-muted/50">
                                    <tr class="text-[10px] uppercase opacity-60">
                                        <th class="w-10">Use</th>
                                        <th class="cursor-pointer hover:bg-muted" on:click={() => toggleRecSort("key")}>Key</th>
                                        <th class="cursor-pointer hover:bg-muted" on:click={() => toggleRecSort("type")}>Type</th>
                                        <th>Coverage</th>
                                        <th>Value / Role</th>
                                        <th class="text-right">Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each sortedRecommendations as rec (rec.id)}
                                        <tr class="hover group {editingId === rec.id ? 'bg-primary/5' : ''}">
                                            <td>
                                                <button 
                                                    class="btn btn-xs btn-ghost p-0 h-6 w-6 {rec.isApproved ? 'text-success' : 'opacity-20'}"
                                                    on:click={() => toggleApprove(rec.id)}
                                                >
                                                    {#if rec.isApproved}<CheckSquare class="w-4 h-4" />{:else}<Square class="w-4 h-4" />{/if}
                                                </button>
                                            </td>
                                            <td class="font-mono text-xs">
                                                {#if editingId === rec.id}
                                                    <input type="text" bind:value={editBuffer.key} class="input input-xs input-bordered w-full font-mono" />
                                                {:else}
                                                    <span class={!rec.isApproved ? 'line-through opacity-30' : ''}>{rec.key}</span>
                                                {/if}
                                            </td>
                                            <td>
                                                <button 
                                                    class="badge badge-xs cursor-pointer hover:scale-105 transition-transform {rec.type === 'global' ? 'badge-info' : 'badge-warning'}"
                                                    on:click={() => toggleType(rec.id)}
                                                >
                                                    {rec.type}
                                                </button>
                                            </td>
                                            <td class="text-xs opacity-60">{(rec.coverage * 100).toFixed(1)}%</td>
                                            <td>
                                                {#if editingId === rec.id}
                                                    <input type="text" bind:value={editBuffer.value} class="input input-xs input-bordered w-full font-mono" />
                                                {:else if rec.type === 'global'}
                                                    <code class="text-[10px] bg-muted px-1 rounded truncate max-w-[150px] inline-block">{rec.value}</code>
                                                {:else}
                                                    <span class="text-[9px] uppercase font-bold text-warning">Dynamic Variable</span>
                                                {/if}
                                            </td>
                                            <td class="text-right">
                                                <div class="flex justify-end gap-1">
                                                    {#if editingId === rec.id}
                                                        <button class="btn btn-xs btn-success btn-square" on:click={saveEditRec}><Check class="w-3 h-3" /></button>
                                                        <button class="btn btn-xs btn-ghost btn-square" on:click={() => editingId = null}><X class="w-3 h-3" /></button>
                                                    {:else}
                                                        <button class="btn btn-xs btn-ghost btn-square opacity-0 group-hover:opacity-100" on:click={() => startEditRec(rec)}><Edit3 class="w-3 h-3" /></button>
                                                        <button class="btn btn-xs btn-ghost btn-square text-error opacity-0 group-hover:opacity-100" on:click={() => deleteRec(rec.id)}><Trash2 class="w-3 h-3" /></button>
                                                    {/if}
                                                </div>
                                            </td>
                                        </tr>
                                    {/each}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>

                <!-- Global Preview Sidebar -->
                <div class="w-96 border-l bg-muted/10 p-6 space-y-4 shrink-0 overflow-y-auto custom-scrollbar">
                    <div class="flex items-center justify-between">
                        <h4 class="font-bold text-sm flex items-center gap-2">
                            <Settings2 class="w-4 h-4 text-primary" /> Global Config
                        </h4>
                        <button class="btn btn-ghost btn-xs gap-1" on:click={copyPreview} disabled={!globalConfigText}>
                            <Copy class="w-3 h-3" /> Copy
                        </button>
                    </div>
                    <div class="bg-muted rounded-xl p-4 font-mono text-[11px] min-h-[400px] border shadow-inner relative group">
                        <pre class="whitespace-pre-wrap">{globalConfigText || '# No global keys defined yet'}</pre>
                        {#if globalConfigText}
                            <div class="absolute inset-0 bg-primary/5 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none"></div>
                        {/if}
                    </div>
                    <div class="p-3 bg-primary/5 rounded-lg border border-primary/10 space-y-2">
                        <div class="flex items-center gap-2 text-[10px] font-bold text-primary uppercase">
                            <Info class="w-3 h-3" /> Note
                        </div>
                        <p class="text-[10px] opacity-70 leading-relaxed">
                            These settings are common to ALL devices in this batch. They will be injected into every line configuration automatically.
                        </p>
                    </div>
                </div>
            </div>
        {:else}
            <!-- Tab 2: Device Explorer (Restored from previous step) -->
            <div class="flex flex-1 overflow-hidden">
                <div class="w-80 border-r bg-muted/10 flex flex-col shrink-0">
                    <div class="p-3 border-b space-y-2 bg-muted/20">
                        <div class="relative">
                            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 opacity-50" />
                            <input 
                                type="text" 
                                bind:value={searchQuery}
                                placeholder="Search MAC or Number..." 
                                class="input input-sm input-bordered w-full pl-9 bg-background focus:ring-2 ring-primary/20"
                            />
                        </div>
                        <div class="flex items-center justify-between px-1">
                            <button class="text-[10px] flex items-center gap-1 opacity-70 hover:opacity-100 uppercase font-bold" on:click={toggleAll}>
                                {#if stats.selected === filteredRecords.length && filteredRecords.length > 0}
                                    <CheckSquare class="w-3 h-3 text-primary" /> Unselect All
                                {:else}
                                    <Square class="w-3 h-3" /> Select All
                                {/if}
                            </button>
                            <span class="text-[10px] opacity-40 italic uppercase">{filteredRecords.length} records</span>
                        </div>
                    </div>

                    <div class="flex-1 overflow-y-auto custom-scrollbar">
                        {#each filteredRecords as record, idx}
                            <div 
                                class="group p-4 border-b cursor-pointer transition-all hover:bg-muted/50 relative {selectedRecordIndex === idx ? 'bg-primary/5 ring-1 ring-inset ring-primary/20 shadow-inner' : ''}"
                                on:click={() => selectedRecordIndex = idx}
                                on:keydown={(e) => e.key === 'Enter' && (selectedRecordIndex = idx)}
                                tabindex="0"
                                role="button"
                            >
                                <div class="flex items-start gap-3">
                                    <button 
                                        class="shrink-0 mt-0.5 text-muted-foreground hover:text-primary transition-colors {isProcessing ? 'pointer-events-none opacity-50' : ''}"
                                        on:click|stopPropagation={() => toggleSelection(record)}
                                    >
                                        {#if record.isSelected}<CheckSquare class="w-4 h-4 text-primary" />{:else}<Square class="w-4 h-4" />{/if}
                                    </button>
                                    <div class="flex-1 min-w-0">
                                        <div class="flex items-center justify-between gap-2 mb-1">
                                            <span class="font-mono text-xs font-bold truncate">{record.data["phone.mac_address"] || "NO-MAC"}</span>
                                            {#if record.status === 'success'}<CheckCircle2 class="w-4 h-4 text-success" />
                                            {:else if record.status === 'error'}<XCircle class="w-4 h-4 text-error" />
                                            {:else if record.status === 'processing'}<div class="loading loading-spinner loading-xs text-primary"></div>
                                            {:else}<Clock class="w-4 h-4 opacity-20" />{/if}
                                        </div>
                                        <div class="flex items-center justify-between text-[10px] opacity-60">
                                            <span class="truncate">{record.data["phone.phone_number"] || "No Line"}</span>
                                            <span class="font-mono">{record.data["phone.model_id"]?.split('-').pop()?.toUpperCase()}</span>
                                        </div>
                                    </div>
                                </div>
                                {#if record.error}
                                    <div class="mt-2 text-[10px] text-error font-medium truncate group-hover:whitespace-normal group-hover:overflow-visible group-hover:z-20 group-hover:bg-error/10 group-hover:p-1 rounded">
                                        {record.error}
                                    </div>
                                {/if}
                            </div>
                        {/each}
                    </div>
                </div>

                <div class="flex-1 flex flex-col min-w-0 bg-background/50">
                    {#if selectedRecord}
                        <div class="flex-1 overflow-y-auto p-8 custom-scrollbar space-y-6">
                            <!-- Detail View Header -->
                            <div class="flex items-start justify-between">
                                <div class="flex items-center gap-4">
                                    <div class="w-16 h-16 bg-primary/10 rounded-2xl flex items-center justify-center text-primary border border-primary/20 shadow-sm"><Cpu class="w-8 h-8" /></div>
                                    <div>
                                        <h3 class="text-2xl font-bold tracking-tight">{selectedRecord.data["phone.mac_address"] || "Unknown Device"}</h3>
                                        <div class="flex items-center gap-2 text-sm opacity-60">
                                            <span class="badge badge-sm badge-outline">{selectedRecord.data["phone.vendor"]}</span>
                                            <span class="badge badge-sm badge-outline font-mono">{selectedRecord.data["phone.model_id"]}</span>
                                            <span class="flex items-center gap-1 border-l pl-2 ml-2"><Library class="w-3 h-3 text-secondary" /> {domain}</span>
                                        </div>
                                    </div>
                                </div>
                                <button class="btn btn-ghost btn-sm text-error h-10 w-10 p-0" on:click={() => removeRecord(selectedRecord.id)} disabled={isProcessing}><Trash2 class="w-4 h-4" /></button>
                            </div>

                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <!-- Details remain the same -->
                                <div class="card bg-muted/20 border border-border overflow-hidden">
                                    <div class="p-4 border-b bg-muted/30 flex items-center gap-2"><CheckSquare class="w-4 h-4 text-primary" /><h4 class="font-bold text-sm">Discovered Accounts</h4></div>
                                    <div class="p-4 space-y-3">
                                        {#each getLineData(selectedRecord.data) as line}
                                            <div class="flex items-center justify-between p-3 bg-background rounded-xl border border-border/50">
                                                <div class="flex items-center gap-3">
                                                    <div class="w-8 h-8 rounded-lg bg-secondary/10 flex items-center justify-center text-secondary text-xs font-bold border border-secondary/20">{line.index + 1}</div>
                                                    <div><div class="text-sm font-bold">{line.user || 'Unknown User'}</div><div class="text-[10px] opacity-60">Label: {line.label || 'None'}</div></div>
                                                </div><Layers class="w-4 h-4 opacity-20" />
                                            </div>
                                        {/each}
                                    </div>
                                </div>
                                <div class="card bg-muted/20 border border-border overflow-hidden">
                                    <div class="p-4 border-b bg-muted/30 flex items-center gap-2"><Layers class="w-4 h-4 text-secondary" /><h4 class="font-bold text-sm">Hardware Mappings</h4></div>
                                    <div class="p-4 space-y-2">
                                        {#each getButtonData(selectedRecord.data) as btn}
                                            <div class="flex items-center justify-between p-2 pl-3 bg-secondary/5 rounded-lg border border-secondary/10 hover:bg-secondary/10 transition-colors">
                                                <div class="flex items-center gap-3"><span class="text-[10px] font-mono opacity-40">#{btn.index}</span><span class="text-xs font-medium uppercase tracking-wider">{btn.type.replace('_', ' ')}</span></div>
                                                <div class="text-[10px] bg-background px-2 py-1 rounded border font-mono">{btn.value}</div>
                                            </div>
                                        {/each}
                                    </div>
                                </div>
                            </div>

                            <!-- Dynamic & Global Features Tile -->
                            <div class="card border border-border bg-gradient-to-br from-muted/30 to-background overflow-hidden">
                                <div class="p-4 border-b bg-muted/20 flex items-center justify-between">
                                    <div class="flex items-center gap-2"><Settings2 class="w-4 h-4 text-primary" /><h4 class="font-bold text-sm">Device Profile Features</h4></div>
                                </div>
                                <div class="p-4 grid grid-cols-2 lg:grid-cols-3 gap-2">
                                    {#each Object.entries(globalData) as [key, value]}
                                        <div class="flex flex-col p-2 rounded bg-background/50 border border-border/50">
                                            <span class="text-[9px] opacity-40 uppercase font-bold truncate" title={key}>{key.split('.').pop()}</span>
                                            <span class="text-xs font-mono truncate">{value}</span>
                                        </div>
                                    {/each}
                                    <!-- Plus any specific dynamic fields for this device that are not global -->
                                    {#each Object.entries(selectedRecord.data).filter(([k]) => k.startsWith('feature.') && !globalData[k.replace('feature.', '')]) as [key, value]}
                                        <div class="flex flex-col p-2 rounded bg-warning/5 border border-warning/20">
                                            <span class="text-[9px] text-warning uppercase font-bold truncate">{key.split('.').pop()}</span>
                                            <span class="text-xs font-mono truncate">{value}</span>
                                        </div>
                                    {/each}
                                </div>
                            </div>
                        </div>
                    {:else}
                        <div class="flex-1 flex flex-col items-center justify-center p-12 text-center space-y-4 opacity-30">
                            <div class="w-24 h-24 rounded-full border-4 border-dashed border-current flex items-center justify-center"><ChevronRight class="w-12 h-12" /></div>
                            <p class="text-lg font-medium">Select a device from the list to preview</p>
                        </div>
                    {/if}
                </div>
            </div>
        {/if}
    </div>
</div>

<style>
    .custom-scrollbar::-webkit-scrollbar { width: 4px; }
    .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
    .custom-scrollbar::-webkit-scrollbar-thumb { background: hsl(var(--bc) / 0.1); border-radius: 10px; }
    .custom-scrollbar::-webkit-scrollbar-thumb:hover { background: hsl(var(--bc) / 0.2); }
</style>
