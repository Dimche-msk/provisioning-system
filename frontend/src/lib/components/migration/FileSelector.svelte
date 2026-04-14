<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { Upload, Settings } from 'lucide-svelte';
    import { ConfigParser } from '$lib/migration/engine';

    const dispatch = createEventDispatcher();
    let dragging = false;
    export let macPattern = '{MAC}';
    let fileList: FileList | null = null;
    let extractedCount = 0;
    let previews: { name: string; mac: string | null }[] = [];
    let failedFiles: string[] = [];

    async function handleFileChange(event: Event) {
        const input = event.target as HTMLInputElement;
        if (!input.files) return;
        fileList = input.files;
        updatePreview();

        // Auto-confirm if we already have a functional pattern and NO failures
        if (macPattern !== '{MAC}' && extractedCount === fileList.length && failedFiles.length === 0) {
            handleConfirm();
        }
    }

    function updatePreview() {
        if (!fileList) return;
        let count = 0;
        previews = [];
        failedFiles = [];
        
        for (let i = 0; i < fileList.length; i++) {
            const mac = ConfigParser.extractMac(fileList[i].name, macPattern);
            if (mac) {
                count++;
            } else {
                failedFiles.push(fileList[i].name);
            }
            
            if (i < 5 || (!mac && failedFiles.length <= 5)) {
                previews.push({ name: fileList[i].name, mac });
            }
        }
        extractedCount = count;
    }

    async function handleConfirm() {
        if (!fileList) return;
        const results: { name: string; content: string }[] = [];
        for (let i = 0; i < fileList.length; i++) {
            const file = fileList[i];
            if (file.name.startsWith('.')) continue; // Skip hidden files
            
            // Strictly only include files that pass the MAC pattern
            const mac = ConfigParser.extractMac(file.name, macPattern);
            if (!mac) continue;

            const content = await file.text();
            results.push({ name: file.name, content });
        }
        dispatch('selected', { files: results, pattern: macPattern });
    }

    function handleDrop(event: DragEvent) {
        dragging = false;
        event.preventDefault();
        if (event.dataTransfer?.files) {
            fileList = event.dataTransfer.files;
            updatePreview();
        }
    }
</script>

<div class="space-y-6">
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div 
        class="flex flex-col items-center justify-center cursor-pointer min-h-[200px] border-2 border-dashed border-border rounded-xl"
        class:bg-muted={dragging}
        on:dragover|preventDefault={() => dragging = true}
        on:dragleave={() => dragging = false}
        on:drop|preventDefault={handleDrop}
        on:click={() => document.getElementById('file-input')?.click()}
    >
        <Upload class="w-12 h-12 mb-2 text-primary" />
        <h3 class="text-lg font-semibold mb-1">Select folder or drag files here</h3>
        
        <input 
            id="file-input"
            type="file" 
            multiple 
            webkitdirectory
            class="hidden" 
            on:change={handleFileChange}
        />
        
        <button class="btn btn-primary btn-sm mt-2">Browse Files</button>
    </div>

    {#if fileList}
        <div class="bg-card p-6 rounded-xl border border-border text-left space-y-4 shadow-sm animate-in fade-in slide-in-from-top-4">
            <h4 class="font-bold flex items-center gap-2">
                <Settings class="w-4 h-4" />
                Configure Extraction
            </h4>
            
            <div class="form-control w-full">
                <label class="label" for="mac-pattern">
                    <span class="label-text">MAC Address Pattern in Filename</span>
                    <span class="label-text-alt text-muted-foreground">e.g. SEP&#123;MAC&#125;.xml</span>
                </label>
                <input 
                    id="mac-pattern"
                    type="text" 
                    bind:value={macPattern} 
                    on:input={updatePreview}
                    placeholder="e.g. &#123;MAC&#125;.cfg or SEP&#123;MAC&#125;.xml"
                    class="input input-bordered w-full font-mono text-sm bg-background text-foreground border-input"
                />
            </div>

            <div class="bg-muted p-4 rounded-xl border border-border space-y-3">
                <div class="flex justify-between items-center bg-card p-3 rounded-lg border border-border">
                    <div>
                        <div class="text-[10px] uppercase font-bold opacity-50">Validation Results</div>
                        <div class="text-xl font-mono font-bold" class:text-success={extractedCount === fileList.length} class:text-warning={extractedCount > 0 && extractedCount < fileList.length} class:text-error={extractedCount === 0}>
                            {extractedCount} <span class="text-sm opacity-50">/ {fileList.length} valid</span>
                        </div>
                    </div>
                    <div class="radial-progress text-primary" style="--value:{Math.round((extractedCount / fileList.length) * 100)}; --size:3rem; --thickness: 4px;">
                        <span class="text-[10px] font-bold">{Math.round((extractedCount / fileList.length) * 100)}%</span>
                    </div>
                </div>

                {#if failedFiles.length > 0}
                    <div class="alert alert-warning p-2 rounded-lg text-xs gap-2">
                        <Settings class="w-4 h-4" />
                        <span>{failedFiles.length} files do not match the MAC pattern and will be ignored.</span>
                    </div>
                {/if}

                <div class="space-y-1">
                    <div class="text-[10px] uppercase font-bold opacity-50 px-1">Extraction Preview</div>
                    <div class="bg-muted/30 rounded-lg overflow-hidden border border-border">
                        <table class="table table-xs w-full">
                            <thead>
                                <tr class="bg-muted/50">
                                    <th>Filename</th>
                                    <th>Resulting MAC</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each previews as preview}
                                    <tr>
                                        <td class="truncate max-w-[150px] font-mono opacity-70">{preview.name}</td>
                                        <td>
                                            {#if preview.mac}
                                                <span class="badge badge-primary badge-xs font-mono">{preview.mac}</span>
                                            {:else}
                                                <span class="text-[10px] text-error italic">no match</span>
                                            {/if}
                                        </td>
                                    </tr>
                                {/each}
                                {#if fileList.length > 5}
                                    <tr>
                                        <td colspan="2" class="text-center opacity-40 italic text-[10px]">... and {fileList.length - 5} more files</td>
                                    </tr>
                                {/if}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            <div class="flex justify-end gap-2 pt-2">
                <button class="btn btn-ghost btn-sm" on:click={() => { fileList = null; }}>Cancel</button>
                <button class="btn btn-primary btn-sm" on:click={handleConfirm} disabled={extractedCount === 0}>
                    Import {extractedCount} Valid Files
                </button>
            </div>
        </div>
    {/if}
</div>
