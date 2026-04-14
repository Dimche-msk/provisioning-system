<script lang="ts">
    import { onMount } from 'svelte';
    import { ConfigParser, DiscoveryEngine, type DiscoveryStepResult, type FlatConfig } from '$lib/migration/engine';
    import FileSelector from '$lib/components/migration/FileSelector.svelte';
    import DiscoveryWizard from '$lib/components/migration/DiscoveryWizard.svelte';
    import ReviewStep from '$lib/components/migration/ReviewStep.svelte';
    import { toast } from 'svelte-sonner';
    import type { DeviceModel } from '$lib/types';

    let files: { name: string; content: string }[] = [];
    let currentStep: 'setup' | 'upload' | 'discovery' | 'review' = 'setup';
    let engine: DiscoveryEngine | null = null;
    
    let vendors: any[] = [];
    let models: DeviceModel[] = [];
    let domains: string[] = [];
    
    let selectedDomain = "";
    let selectedVendorId = "";
    let selectedModelId = "";
    let loading = false;
    let showAddModal = false;

    onMount(async () => {
        await Promise.all([loadVendors(), loadDomains()]);
    });

    async function loadVendors() {
        try {
            const res = await fetch("/api/vendors");
            if (res.ok) {
                const data = await res.json();
                vendors = data.vendors || [];
                if (vendors.length > 0) selectedVendorId = vendors[0].id;
            }
        } catch (e) {
            toast.error("Failed to load vendors");
        }
    }

    async function loadDomains() {
        try {
            const res = await fetch("/api/domains");
            if (res.ok) {
                const data = await res.json();
                domains = data.domains || [];
                if (domains.length > 0) selectedDomain = domains[0];
            }
        } catch (e) {
            toast.error("Failed to load domains");
        }
    }

    async function loadModels(vId: string) {
        if (!vId) return;
        loading = true;
        try {
            const res = await fetch(`/api/models?vendor=${vId}&include_modules=true`);
            if (res.ok) {
                const data = await res.json();
                models = data.models || [];
                if (models.length > 0) selectedModelId = models[0].id;
            }
        } catch (e) {
            toast.error("Failed to load models");
        } finally {
            loading = false;
        }
    }

    $: if (selectedVendorId) {
        loadModels(selectedVendorId);
    }

    $: selectedVendor = vendors.find(v => v.id === selectedVendorId);
    $: selectedModel = models.find(m => m.id === selectedModelId);

    async function handleFilesSelected(event: CustomEvent<{ files: { name: string; content: string }[]; pattern: string }>) {
        const newFiles = event.detail.files;
        if (newFiles.length === 0) return;
        
        if (currentStep === "upload") {
            files = newFiles;
            engine = new DiscoveryEngine();
            engine.setMacPattern(event.detail.pattern);
            
            // Ingest all files
            const firstFile = ConfigParser.parse(files[0].content, files[0].name);
            engine.setBaseline(firstFile, selectedVendorId, selectedModelId, selectedDomain);
            
            if (files.length > 1) {
                for (let i = 1; i < files.length; i++) {
                    const config = ConfigParser.parse(files[i].content, files[i].name);
                    engine.ingestConfig(config, selectedVendorId, selectedModelId, selectedDomain, files[i].content);
                }
            }
           
            currentStep = "discovery";
        } else {
            // Append mode with de-duplication
            const currentPattern = event.detail.pattern;
            const existingMacs = new Set(
                files.map((f) => ConfigParser.extractMac(f.name, engine?.getMacPattern() || "{MAC}"))
                    .filter(Boolean)
            );

            const filteredNewFiles = newFiles.filter((nf) => {
                const mac = ConfigParser.extractMac(nf.name, currentPattern);
                if (!mac) {
                    console.log(`Skipping file that doesn't match pattern: ${nf.name}`);
                    return false;
                }
                if (existingMacs.has(mac)) {
                    console.log(`Skipping duplicate file with MAC: ${mac}`);
                    return false;
                }
                return true;
            });

            if (filteredNewFiles.length === 0) {
                toast.info("No new unique files found in this batch.");
                showAddModal = false;
                return;
            }

            // Ingest new files with metadata (model is selected per batch)
            for (const nf of filteredNewFiles) {
                const config = ConfigParser.parse(nf.content, nf.name);
                engine.ingestConfig(config, selectedVendorId, selectedModelId, selectedDomain, nf.content);
            }

            files = [...files, ...filteredNewFiles];
            showAddModal = false;
            toast.success(`Added ${filteredNewFiles.length} new unique files. Total: ${files.length}`);
            return;
        }
    }

    function handleReset() {
        files = [];
        engine = null;
        currentStep = 'setup';
    }

    function startUpload() {
        if (!selectedDomain || !selectedVendorId) {
            toast.error("Please complete the setup first");
            return;
        }
        currentStep = 'upload';
    }
</script>

<div class="container mx-auto py-8">
    <div class="flex items-center justify-between mb-8">
        <div>
            <h1 class="text-3xl font-bold">Migration Wizard</h1>
            <p class="text-muted-foreground">Import legacy configuration files and teach the system the logic.</p>
        </div>
    </div>

    {#if currentStep === 'setup'}
        <div class="max-w-2xl mx-auto">
            <div class="card bg-card border border-border shadow-lg">
                <div class="card-body gap-6">
                    <h2 class="text-xl font-bold flex items-center gap-2">
                        <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-primary-foreground text-sm">1</span>
                        Session Setup
                    </h2>
                    
                    <div class="grid grid-cols-1 gap-4">
                        <div class="space-y-2">
                            <label class="text-sm font-medium">
                                Target Domain
                                <select bind:value={selectedDomain} class="select select-bordered w-full bg-background mt-1">
                                    {#each domains as domain}
                                        <option value={domain}>{domain}</option>
                                    {/each}
                                </select>
                            </label>
                        </div>

                        <div class="grid grid-cols-1 gap-4">
                            <div class="space-y-2">
                                <label class="text-sm font-medium">
                                    Hardware Vendor
                                    <select bind:value={selectedVendorId} class="select select-bordered w-full bg-background mt-1">
                                        {#each vendors as v}
                                            <option value={v.id}>{v.name}</option>
                                        {/each}
                                    </select>
                                </label>
                            </div>
                        </div>
                    </div>

                    <div class="pt-4">
                        <button class="btn btn-primary w-full" on:click={startUpload} disabled={loading}>
                            Continue to Upload
                        </button>
                    </div>
                </div>
            </div>
        </div>
    {:else if currentStep === 'upload'}
        <div class="max-w-4xl mx-auto space-y-6">
            <div class="card bg-card border border-border">
                <div class="card-body">
                    <h3 class="text-lg font-bold mb-4 italic text-primary">Target Hardware for this Batch: {selectedVendor?.name}</h3>
                    <div class="space-y-2">
                        <label class="text-sm font-medium">
                            Select Model for these files
                            <select bind:value={selectedModelId} class="select select-bordered w-full bg-background mt-1" disabled={loading}>
                                {#each models.filter(m => m.type !== 'expansion-module') as m}
                                    <option value={m.id}>{m.name}</option>
                                {/each}
                            </select>
                        </label>
                    </div>
                </div>
            </div>

            <div class="card bg-muted p-12 text-center border-2 border-dashed border-border group hover:border-primary/50 transition-colors">
                <FileSelector on:selected={handleFilesSelected} />
            </div>
        </div>
    {:else if currentStep === 'discovery' && engine}
        <DiscoveryWizard 
            {files} 
            {engine} 
            vendor={selectedVendor}
            availableModels={models}
            {selectedModelId}
            on:complete={() => currentStep = 'review'} 
            on:addMore={handleReset}
            on:requestAdd={() => showAddModal = true}
            on:filesChanged={(e) => files = e.detail.files}
        />
    {:else if currentStep === 'review' && engine}
        <ReviewStep 
            {engine} 
            vendor={selectedVendorId} 
            domain={selectedDomain}
            on:applied={handleReset} 
            on:back={() => currentStep = 'discovery'}
        />
    {/if}

    {#if showAddModal}
        <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm animate-in fade-in duration-200">
            <div class="card bg-card w-full max-w-2xl shadow-2xl border border-border">
                <div class="card-body">
                    <div class="flex items-center justify-between mb-4">
                        <h3 class="text-lg font-bold">Append More Files</h3>
                        <button class="btn btn-sm btn-circle btn-ghost" on:click={() => showAddModal = false}>✕</button>
                    </div>
                    <div class="space-y-4 mb-6">
                        <label class="text-sm font-medium">
                            Hardware Model for new files
                            <select bind:value={selectedModelId} class="select select-bordered w-full bg-background mt-1" disabled={loading}>
                                {#each models.filter(m => m.type !== 'expansion-module') as m}
                                    <option value={m.id}>{m.name}</option>
                                {/each}
                            </select>
                        </label>
                    </div>
                    
                    <FileSelector 
                        macPattern={engine?.getMacPattern() || '{MAC}'} 
                        on:selected={handleFilesSelected} 
                    />
                    <div class="modal-action">
                        <button class="btn btn-ghost" on:click={() => showAddModal = false}>Cancel</button>
                    </div>
                </div>
            </div>
        </div>
    {/if}
</div>
