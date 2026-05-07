<script lang="ts">
    import { onMount } from "svelte";
    import { t } from "svelte-i18n";
    import { toast } from "svelte-sonner";
    import * as Card from "$lib/components/ui/card";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import { Checkbox } from "$lib/components/ui/checkbox";
    import { Save, AlertCircle, Info, Plus, Trash2, ChevronDown, ChevronRight, FileCode, X } from "lucide-svelte";
    import * as Tabs from "$lib/components/ui/tabs";
    import { Alert, AlertDescription, AlertTitle } from "$lib/components/ui/alert";

    let config: any = {
        server: {},
        auth: {},
        database: {},
        domains: []
    };
    let sample = "";
    let metadata: Record<string, any> = {};
    let loading = true;
    let saving = false;
    let activeTab = "general";
    let vendors: any[] = [];
    let selectedVendorId = "";
    let templateContent = "";
    let vendorTemplates: string[] = [];
    let selectedTemplateFile = "";
    $: selectedVendor = vendors.find(v => v.id === selectedVendorId);
    $: if (selectedVendorId) {
        loadVendorTemplates();
    }
    $: if (selectedVendorId && selectedTemplateFile) {
        loadTemplateFile();
    }

    onMount(async () => {
        await loadData();
    });

    async function loadData() {
        loading = true;
        try {
            const [configRes, sampleRes] = await Promise.all([
                fetch("/api/system/config"),
                fetch("/api/system/config/sample")
            ]);

            if (configRes.ok) {
                config = await configRes.json();
            }

            if (sampleRes.ok) {
                sample = await sampleRes.text();
                parseMetadata(sample);
            }
        } catch (e: any) {
            console.error("Failed to load config data", e);
            toast.error($t("templates.load_error") || "Failed to load configuration");
        } finally {
            loading = false;
        }
        loadVendors();
    }

    async function loadVendors() {
        try {
            const res = await fetch("/api/vendors");
            if (res.ok) {
                const data = await res.json();
                vendors = data.vendors || [];
                if (vendors.length > 0 && !selectedVendorId) {
                    selectedVendorId = vendors[0].id;
                }
            }
        } catch (e) {
            console.error("Failed to load vendors", e);
        }
    }

    function parseMetadata(yamlText: string) {
        const lines = yamlText.split("\n");
        let currentSection = "";
        const newMetadata: Record<string, any> = {};
        
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            
            // Check for section
            const sectionMatch = line.match(/#\s*\[section:\s*([^,\]]+)(?:,\s*help:\s*([^\]]+))?\]/i);
            if (sectionMatch) {
                currentSection = sectionMatch[1].trim();
                continue;
            }

            // Check for metadata tag [...]
            const tagMatch = line.match(/#\s*\[(.*?)\]/);
            if (!tagMatch) continue;

            const tagContent = tagMatch[1];
            if (!tagContent.includes("label:")) continue;

            const parts: Record<string, string> = {};
            // Improved parsing: find all key:value pairs
            // This regex looks for word: followed by either a quoted string or text until the next key: or end of string
            const pairRegex = /(\w+):\s*("(?:[^"\\]|\\.)*"|.+?)(?=\s*,\s*\w+:|$)/g;
            let match;
            while ((match = pairRegex.exec(tagContent)) !== null) {
                const k = match[1].toLowerCase().trim();
                let v = match[2].trim();
                // Remove quotes if present
                if (v.startsWith('"') && v.endsWith('"')) {
                    v = v.substring(1, v.length - 1);
                }
                parts[k] = v;
            }

            // Find the key this metadata belongs to (look at current line, then next few lines)
            let key = "";
            for (let j = i; j < Math.min(i + 4, lines.length); j++) {
                const searchLine = lines[j];
                const keyMatch = searchLine.match(/^\s*(?:-\s+)?([\w_]+):/);
                if (keyMatch) {
                    key = keyMatch[1];
                    break;
                }
            }

            if (key) {
                newMetadata[key] = {
                    label: parts.label,
                    type: parts.type || "string",
                    help: parts.help,
                    options: parts.options ? parts.options.split(",").map(o => o.trim()) : [],
                    readonly: parts.readonly === "true",
                    section: currentSection
                };
            }
        }
        metadata = newMetadata;
    }

    async function saveConfig() {
        if (!confirm($t("templates.confirm_save"))) return;

        saving = true;
        try {
            const res = await fetch("/api/system/config", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(config)
            });

            if (res.ok) {
                toast.success($t("templates.save_success"));
                await loadData();
            } else {
                const data = await res.json();
                toast.error(data.error || $t("templates.save_error"));
            }
        } catch (e: any) {
            toast.error($t("templates.save_error") + ": " + e.message);
        } finally {
            saving = false;
        }
    }

    function addDomain() {
        if (!config.domains) config.domains = [];
        config.domains = [...config.domains, {
            name: "new-domain",
            generate_random_password: true,
            deploy_commands: [],
            delete_commands: [],
            variables: {}
        }];
    }

    function removeDomain(index: number) {
        config.domains = config.domains.filter((_: any, i: number) => i !== index);
    }

    function addVariable(domainIndex: number) {
        const domain = config.domains[domainIndex];
        if (!domain.variables) domain.variables = {};
        const key = prompt("Variable name:");
        if (key && !domain.variables[key]) {
            domain.variables[key] = "";
            config = { ...config };
        }
    }

    function removeVariable(domainIndex: number, key: string) {
        delete config.domains[domainIndex].variables[key];
        config = { ...config };
    }

    function addCommand(domainIndex: number, type: string) {
        const domain = config.domains[domainIndex];
        if (!domain[type]) domain[type] = [];
        domain[type] = [...domain[type], ""];
        config = { ...config };
    }

    function removeCommand(domainIndex: number, type: string, cmdIndex: number) {
        config.domains[domainIndex][type] = config.domains[domainIndex][type].filter((_: any, i: number) => i !== cmdIndex);
        config = { ...config };
    }

    async function saveVendorData(type: 'features' | 'accounts') {
        if (!selectedVendor) return;
        const confirmMsg = type === 'features' ? $t("templates.confirm_features_save") : $t("templates.confirm_accounts_save");
        if (!confirm(confirmMsg || `Save ${type} for ${selectedVendor.name}?`)) return;

        saving = true;
        try {
            const res = await fetch(`/api/vendors/${selectedVendor.id}/${type}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(selectedVendor[type])
            });

            if (res.ok) {
                toast.success($t("templates.save_success"));
            } else {
                const data = await res.json();
                toast.error(data.error || $t("templates.save_error"));
            }
        } catch (e: any) {
            toast.error($t("templates.save_error") + ": " + e.message);
        } finally {
            saving = false;
        }
    }

    function addFeature(type: 'features' | 'accounts') {
        if (!selectedVendor) return;
        const newFeature = {
            id: "new_feature",
            name: "New Feature",
            params: []
        };
        selectedVendor[type] = [...selectedVendor[type], newFeature];
        vendors = [...vendors];
    }

    function removeFeature(type: 'features' | 'accounts', index: number) {
        selectedVendor[type] = selectedVendor[type].filter((_: any, i: number) => i !== index);
        vendors = [...vendors];
    }

    function addParam(type: 'features' | 'accounts', featureIndex: number) {
        const feature = selectedVendor[type][featureIndex];
        const newParam = {
            id: "new_param",
            label: "New Param",
            type: "string",
            config_template: ""
        };
        feature.params = [...(feature.params || []), newParam];
        vendors = [...vendors];
    }

    function removeParam(type: 'features' | 'accounts', featureIndex: number, paramIndex: number) {
        selectedVendor[type][featureIndex].params = selectedVendor[type][featureIndex].params.filter((_: any, i: number) => i !== paramIndex);
        vendors = [...vendors];
    }

    async function loadVendorTemplates() {
        if (!selectedVendorId) return;
        templateContent = ""; // Clear current content while loading
        try {
            const res = await fetch(`/api/vendors/${selectedVendorId}/templates`);
            if (res.ok) {
                const data = await res.json();
                const files = data.files || [];
                vendorTemplates = files;
                
                // Auto-select the main template if it exists in the list
                const mainTpl = selectedVendor?.phone_config_template;
                
                if (mainTpl && files.includes(mainTpl)) {
                    selectedTemplateFile = mainTpl;
                } else if (files.length > 0) {
                    selectedTemplateFile = files[0];
                } else {
                    selectedTemplateFile = "";
                }
                
                // Manually trigger load if file stayed the same but vendor changed
                loadTemplateFile();
            }
        } catch (e) {
            console.error("Failed to load vendor templates", e);
        }
    }

    async function loadTemplateFile() {
        if (!selectedVendorId || !selectedTemplateFile) return;
        try {
            const res = await fetch(`/api/vendors/${selectedVendorId}/templates/file?file=${encodeURIComponent(selectedTemplateFile)}`);
            if (res.ok) {
                templateContent = await res.text();
            } else {
                templateContent = "";
            }
        } catch (e) {
            console.error("Failed to load template file", e);
        }
    }

    async function saveTemplate() {
        if (!selectedVendorId || !selectedTemplateFile) return;
        if (!confirm($t("templates.confirm_template_save") || "Save template changes?")) return;

        saving = true;
        try {
            const res = await fetch(`/api/vendors/${selectedVendorId}/templates/file?file=${encodeURIComponent(selectedTemplateFile)}`, {
                method: "POST",
                headers: { "Content-Type": "text/plain" },
                body: templateContent
            });

            if (res.ok) {
                toast.success($t("templates.save_success"));
            } else {
                const data = await res.json();
                toast.error(data.error || $t("templates.save_error"));
            }
        } catch (e: any) {
            toast.error($t("templates.save_error") + ": " + e.message);
        } finally {
            saving = false;
        }
    }
</script>

<div class="p-6 space-y-6">
    <div class="flex justify-between items-center">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
            {$t("templates.title")}
        </h1>
        <Button on:click={saveConfig} disabled={saving || loading}>
            <Save class="mr-2 h-4 w-4" />
            {saving ? $t("common.saving") : $t("templates.save_button")}
        </Button>
    </div>

    <Alert>
        <Info class="h-4 w-4" />
        <AlertTitle>{$t("common.notice") || "Notice"}</AlertTitle>
        <AlertDescription>
            {$t("templates.backup_notice")}
        </AlertDescription>
    </Alert>

    {#if loading}
        <div class="flex justify-center p-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
    {:else if config}
        <Tabs.Root value="general" class="w-full">
            <Tabs.List class="grid w-full grid-cols-2 max-w-md">
                <Tabs.Trigger value="general">{$t("templates.general_settings")}</Tabs.Trigger>
                <Tabs.Trigger value="functions">{$t("templates.function_settings")}</Tabs.Trigger>
            </Tabs.List>

            <Tabs.Content value="general" class="space-y-6 pt-6">
                <!-- Server Section -->
                <Card.Root>
                    <Card.Header>
                        <Card.Title>{metadata.listen_address?.section || "Server"}</Card.Title>
                    </Card.Header>
                    <Card.Content class="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {#if config?.server}
                            {#each Object.keys(config?.server || {}) as key}
                                <div class="space-y-2">
                                    <Label for={"server-" + key} class="flex items-center gap-2">
                                        {metadata[key]?.label || key}
                                        {#if metadata[key]?.readonly}
                                            <span class="text-[10px] bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded uppercase font-bold">Readonly</span>
                                        {/if}
                                    </Label>
                                    
                                    {#if metadata[key]?.type === "boolean"}
                                        <div class="flex items-center space-x-2 pt-2">
                                            <Checkbox id={"server-" + key} bind:checked={config.server[key]} disabled={metadata[key]?.readonly} />
                                            <p class="text-sm text-muted-foreground">{metadata[key]?.help || ""}</p>
                                        </div>
                                    {:else if metadata[key]?.type === "select"}
                                        <select 
                                            id={"server-" + key}
                                            bind:value={config.server[key]}
                                            disabled={metadata[key]?.readonly}
                                            class="w-full h-10 px-3 py-2 rounded-md border border-input bg-background text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                                        >
                                            {#each metadata[key]?.options || [] as opt}
                                                <option value={opt}>{opt}</option>
                                            {/each}
                                        </select>
                                    {:else}
                                        <Input 
                                            id={"server-" + key} 
                                            type={metadata[key]?.type === "number" ? "number" : "text"}
                                            bind:value={config.server[key]} 
                                            disabled={metadata[key]?.readonly}
                                            placeholder={metadata[key]?.help}
                                        />
                                    {/if}
                                    
                                    {#if metadata[key]?.help && metadata[key]?.type !== "boolean"}
                                        <p class="text-xs text-muted-foreground italic">{metadata[key]?.help}</p>
                                    {/if}

                                    {#if metadata[key]?.readonly}
                                        <p class="text-[10px] text-amber-600">{$t("templates.readonly_notice")}</p>
                                    {/if}
                                </div>
                            {/each}
                        {/if}
                    </Card.Content>
                </Card.Root>

                <!-- Auth Section -->
                <Card.Root>
                    <Card.Header>
                        <Card.Title>{metadata.admin_user?.section || "Authentication"}</Card.Title>
                    </Card.Header>
                    <Card.Content class="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {#if config?.auth}
                            {#each Object.keys(config?.auth || {}) as key}
                                <div class="space-y-2">
                                    <Label for={"auth-" + key} class="flex items-center gap-2">
                                        {metadata[key]?.label || key}
                                        {#if metadata[key]?.readonly}
                                            <span class="text-[10px] bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded uppercase font-bold">Readonly</span>
                                        {/if}
                                    </Label>
                                    <Input 
                                        id={"auth-" + key} 
                                        type={metadata[key]?.type === "password" ? "password" : "text"}
                                        bind:value={config.auth[key]} 
                                        disabled={metadata[key]?.readonly}
                                        placeholder={metadata[key]?.type === "password" ? "********" : metadata[key]?.help}
                                    />
                                    {#if metadata[key]?.help}
                                        <p class="text-xs text-muted-foreground italic">{metadata[key]?.help}</p>
                                    {/if}
                                </div>
                            {/each}
                        {/if}
                    </Card.Content>
                </Card.Root>

                <!-- Database Section -->
                <Card.Root>
                    <Card.Header>
                        <Card.Title>{metadata.path?.section || "Database"}</Card.Title>
                    </Card.Header>
                    <Card.Content class="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {#if config?.database}
                            {#each Object.keys(config?.database || {}) as key}
                                <div class="space-y-2">
                                    <Label for={"db-" + key} class="flex items-center gap-2">
                                        {metadata[key]?.label || key}
                                        {#if metadata[key]?.readonly}
                                            <span class="text-[10px] bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded uppercase font-bold">Readonly</span>
                                        {/if}
                                    </Label>
                                    <Input 
                                        id={"db-" + key} 
                                        bind:value={config.database[key]} 
                                        disabled={metadata[key]?.readonly}
                                    />
                                    {#if metadata[key]?.help}
                                        <p class="text-xs text-muted-foreground italic">{metadata[key]?.help}</p>
                                    {/if}
                                </div>
                            {/each}
                        {/if}
                    </Card.Content>
                </Card.Root>

                <!-- Domains Section -->
                <div class="space-y-4">
                    <div class="flex justify-between items-center">
                        <div class="space-y-1">
                            <h2 class="text-2xl font-bold">{metadata.domains?.label || "Domains Management"}</h2>
                            {#if metadata.domains?.help}
                                <p class="text-sm text-muted-foreground">{metadata.domains?.help}</p>
                            {/if}
                        </div>
                        <Button variant="outline" size="sm" disabled={metadata.domains?.readonly} on:click={addDomain}>
                            <Plus class="mr-2 h-4 w-4" />
                            {$t("templates.add_domain")}
                        </Button>
                    </div>

                    {#each config?.domains || [] as domain, i}
                        <Card.Root>
                            <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
                                <Card.Title class="text-sm font-medium">Domain: {domain.name}</Card.Title>
                                <Button variant="ghost" size="sm" class="text-destructive" disabled={metadata.domains?.readonly} on:click={() => removeDomain(i)}>
                                    <Trash2 class="h-4 w-4" />
                                </Button>
                            </Card.Header>
                            <Card.Content class="space-y-4">
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div class="space-y-2">
                                        <Label class="flex items-center gap-2">
                                            {metadata.name?.label || "Domain Name"}
                                            {#if metadata.name?.readonly}
                                                <span class="text-[10px] bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded uppercase font-bold">Readonly</span>
                                            {/if}
                                        </Label>
                                        <Input bind:value={domain.name} disabled={metadata.name?.readonly} placeholder={metadata.name?.help} />
                                        {#if metadata.name?.help}
                                            <p class="text-xs text-muted-foreground italic">{metadata.name?.help}</p>
                                        {/if}
                                    </div>
                                    <div class="flex flex-col justify-end space-y-2 pb-2">
                                        <div class="flex items-center space-x-2">
                                            <Checkbox id={"rand-pass-" + i} bind:checked={domain.generate_random_password} disabled={metadata.generate_random_password?.readonly} />
                                            <Label for={"rand-pass-" + i} class="flex items-center gap-2">
                                                {metadata.generate_random_password?.label || "Auto-generate Passwords"}
                                                {#if metadata.generate_random_password?.readonly}
                                                    <span class="text-[10px] bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded uppercase font-bold">Readonly</span>
                                                {/if}
                                            </Label>
                                        </div>
                                        {#if metadata.generate_random_password?.help}
                                            <p class="text-xs text-muted-foreground italic">{metadata.generate_random_password?.help}</p>
                                        {/if}
                                    </div>
                                </div>

                                <!-- Commands -->
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div class="space-y-2">
                                        <div class="flex justify-between items-center">
                                            <Label class="flex items-center gap-2">
                                                {metadata.deploy_commands?.label || "Deploy Commands"}
                                                {#if metadata.deploy_commands?.readonly}
                                                    <span class="text-[10px] bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded uppercase font-bold">Readonly</span>
                                                {/if}
                                            </Label>
                                            <Button variant="ghost" size="sm" disabled={metadata.deploy_commands?.readonly} on:click={() => addCommand(i, 'deploy_commands')}>
                                                <Plus class="h-3 w-3" />
                                            </Button>
                                        </div>
                                        <div class="space-y-2">
                                            {#each domain.deploy_commands || [] as cmd, ci}
                                                <div class="flex flex-col gap-1 border rounded-md p-2 bg-muted/20 relative group/cmd">
                                                    <textarea 
                                                        bind:value={domain.deploy_commands[ci]} 
                                                        readonly={metadata.deploy_commands?.readonly} 
                                                        rows="3"
                                                        class="w-full bg-transparent text-xs font-mono resize-y focus:outline-none {metadata.deploy_commands?.readonly ? 'opacity-70' : ''}"
                                                        placeholder="Enter command..."
                                                    ></textarea>
                                                    {#if !metadata.deploy_commands?.readonly}
                                                        <Button variant="ghost" size="sm" class="absolute top-1 right-1 opacity-0 group-hover/cmd:opacity-100 h-6 w-6 p-0" on:click={() => removeCommand(i, 'deploy_commands', ci)}>
                                                            <Trash2 class="h-3 w-3" />
                                                        </Button>
                                                    {/if}
                                                </div>
                                            {/each}
                                            {#if !domain.deploy_commands || domain.deploy_commands.length === 0}
                                                <p class="text-xs text-muted-foreground italic">No commands defined.</p>
                                            {/if}
                                        </div>
                                        {#if metadata.deploy_commands?.help}
                                            <p class="text-xs text-muted-foreground italic">{metadata.deploy_commands?.help}</p>
                                        {/if}
                                    </div>

                                    <div class="space-y-2">
                                        <div class="flex justify-between items-center">
                                            <Label class="flex items-center gap-2">
                                                {metadata.delete_commands?.label || "Delete Commands"}
                                                {#if metadata.delete_commands?.readonly}
                                                    <span class="text-[10px] bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded uppercase font-bold">Readonly</span>
                                                {/if}
                                            </Label>
                                            <Button variant="ghost" size="sm" disabled={metadata.delete_commands?.readonly} on:click={() => addCommand(i, 'delete_commands')}>
                                                <Plus class="h-3 w-3" />
                                            </Button>
                                        </div>
                                        <div class="space-y-2">
                                            {#each domain.delete_commands || [] as cmd, ci}
                                                <div class="flex flex-col gap-1 border rounded-md p-2 bg-muted/20 relative group/cmd">
                                                    <textarea 
                                                        bind:value={domain.delete_commands[ci]} 
                                                        readonly={metadata.delete_commands?.readonly} 
                                                        rows="3"
                                                        class="w-full bg-transparent text-xs font-mono resize-y focus:outline-none {metadata.delete_commands?.readonly ? 'opacity-70' : ''}"
                                                        placeholder="Enter command..."
                                                    ></textarea>
                                                    {#if !metadata.delete_commands?.readonly}
                                                        <Button variant="ghost" size="sm" class="absolute top-1 right-1 opacity-0 group-hover/cmd:opacity-100 h-6 w-6 p-0" on:click={() => removeCommand(i, 'delete_commands', ci)}>
                                                            <Trash2 class="h-3 w-3" />
                                                        </Button>
                                                    {/if}
                                                </div>
                                            {/each}
                                            {#if !domain.delete_commands || domain.delete_commands.length === 0}
                                                <p class="text-xs text-muted-foreground italic">No commands defined.</p>
                                            {/if}
                                        </div>
                                        {#if metadata.delete_commands?.help}
                                            <p class="text-xs text-muted-foreground italic">{metadata.delete_commands?.help}</p>
                                        {/if}
                                    </div>
                                </div>

                                <!-- Variables -->
                                <div class="space-y-2 border-t pt-4">
                                    <div class="flex justify-between items-center">
                                        <Label class="flex items-center gap-2">
                                            {metadata.variables?.label || "Variables"}
                                            {#if metadata.variables?.readonly}
                                                <span class="text-[10px] bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded uppercase font-bold">Readonly</span>
                                            {/if}
                                        </Label>
                                        <Button variant="outline" size="sm" disabled={metadata.variables?.readonly} on:click={() => addVariable(i)}>
                                            <Plus class="h-3 w-3 mr-1" /> Add Var
                                        </Button>
                                    </div>
                                    {#if metadata.variables?.help}
                                        <p class="text-xs text-muted-foreground italic mb-2">{metadata.variables?.help}</p>
                                    {/if}
                                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                        {#each Object.keys(domain.variables || {}) as key}
                                            <div class="flex flex-col gap-1 p-2 border rounded-md relative group">
                                                <span class="text-[10px] font-bold text-amber-600 dark:text-amber-500">{key}</span>
                                                <Input bind:value={domain.variables[key]} class="h-8 text-sm" />
                                                <button 
                                                    class="absolute top-1 right-1 opacity-0 group-hover:opacity-100 text-destructive p-1 hover:bg-destructive/10 rounded"
                                                    on:click={() => removeVariable(i, key)}
                                                >
                                                    <Trash2 class="h-3 w-3" />
                                                </button>
                                            </div>
                                        {/each}
                                    </div>
                                </div>
                            </Card.Content>
                        </Card.Root>
                    {/each}
                </div>
            </Tabs.Content>

            <Tabs.Content value="functions" class="pt-6 space-y-6">
                <div class="flex gap-4 items-center bg-muted/30 p-4 rounded-lg">
                    <Label class="whitespace-nowrap">Select Vendor:</Label>
                    <select 
                        bind:value={selectedVendorId}
                        class="max-w-xs w-full h-10 px-3 py-2 rounded-md border border-input bg-background text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                    >
                        {#each vendors as v}
                            <option value={v.id}>{v.name}</option>
                        {/each}
                    </select>
                </div>

                {#if selectedVendor}
                    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                        <!-- Features Column -->
                        <div class="space-y-4">
                            <div class="flex justify-between items-center">
                                <h3 class="text-xl font-bold flex items-center gap-2">
                                    <FileCode class="w-5 h-5" />
                                    Features (Buttons)
                                </h3>
                                <div class="flex gap-2">
                                    <Button variant="outline" size="sm" on:click={() => addFeature('features')}>
                                        <Plus class="w-4 h-4 mr-1" /> Add
                                    </Button>
                                    <Button size="sm" on:click={() => saveVendorData('features')} disabled={saving}>
                                        <Save class="w-4 h-4 mr-1" /> Save
                                    </Button>
                                </div>
                            </div>

                            <div class="space-y-3">
                                {#each selectedVendor.features || [] as feature, fi}
                                    <Card.Root>
                                        <Card.Header class="p-4 pb-2 border-b bg-muted/10">
                                            <div class="flex justify-between items-center">
                                                <div class="space-y-1">
                                                    <h4 class="text-sm font-bold text-primary">{feature.name || "Unnamed Feature"}</h4>
                                                    <p class="text-[10px] font-mono opacity-60">ID: {feature.id}</p>
                                                </div>
                                                <Button variant="ghost" size="sm" class="text-destructive h-8 w-8 p-0" on:click={() => removeFeature('features', fi)}>
                                                    <Trash2 class="h-4 w-4" />
                                                </Button>
                                            </div>
                                        </Card.Header>
                                        <Card.Content class="p-4 space-y-4">
                                            <div class="grid grid-cols-2 gap-3">
                                                <div class="space-y-1">
                                                    <Label class="text-[10px] uppercase opacity-70">Feature ID</Label>
                                                    <Input bind:value={feature.id} placeholder="e.g. blf" class="h-8 font-mono text-xs" />
                                                </div>
                                                <div class="space-y-1">
                                                    <Label class="text-[10px] uppercase opacity-70">Display Name</Label>
                                                    <Input bind:value={feature.name} placeholder="Feature Name" class="h-8 text-xs" />
                                                </div>
                                            </div>

                                            <div class="flex gap-4">
                                                <div class="flex items-center space-x-2">
                                                    <Checkbox id={"feat-acc-"+fi} bind:checked={feature.associated_with_account} />
                                                    <Label for={"feat-acc-"+fi} class="text-[10px] uppercase cursor-pointer">Account Linked</Label>
                                                </div>
                                                <div class="flex items-center space-x-2">
                                                    <Checkbox id={"feat-btn-"+fi} bind:checked={feature.associated_with_button} />
                                                    <Label for={"feat-btn-"+fi} class="text-[10px] uppercase cursor-pointer">Button Linked</Label>
                                                </div>
                                            </div>

                                            <div class="border-t pt-3 mt-2">
                                                <div class="flex justify-between items-center mb-3">
                                                    <span class="text-xs font-bold uppercase text-muted-foreground">Parameters</span>
                                                    <Button variant="outline" size="sm" class="h-7 px-2 text-[10px]" on:click={() => addParam('features', fi)}>
                                                        <Plus class="w-3 h-3 mr-1" /> Add Param
                                                    </Button>
                                                </div>
                                                <div class="space-y-3">
                                                    {#each feature.params || [] as param, pi}
                                                        <div class="bg-muted/30 p-3 rounded-md border border-dashed space-y-3 relative group">
                                                            <div class="grid grid-cols-3 gap-2">
                                                                <div class="space-y-1">
                                                                    <Label class="text-[9px] uppercase opacity-60">Param ID</Label>
                                                                    <Input bind:value={param.id} class="h-7 text-[10px] font-mono" />
                                                                </div>
                                                                <div class="space-y-1">
                                                                    <Label class="text-[9px] uppercase opacity-60">Label</Label>
                                                                    <Input bind:value={param.label} class="h-7 text-[10px]" />
                                                                </div>
                                                                <div class="space-y-1">
                                                                    <Label class="text-[9px] uppercase opacity-60">Type</Label>
                                                                    <select bind:value={param.type} class="w-full h-7 text-[10px] rounded border bg-background px-1">
                                                                        <option value="string">String</option>
                                                                        <option value="number">Number</option>
                                                                        <option value="boolean">Boolean</option>
                                                                        <option value="select">Select</option>
                                                                        <option value="password">Password</option>
                                                                        <option value="hidden">Hidden</option>
                                                                    </select>
                                                                </div>
                                                            </div>
                                                            
                                                            {#if param.type === 'select'}
                                                                <div class="grid grid-cols-2 gap-2">
                                                                    <div class="space-y-1">
                                                                        <Label class="text-[9px] uppercase opacity-60">Source</Label>
                                                                        <Input bind:value={param.source} placeholder="e.g. accounts" class="h-7 text-[10px]" />
                                                                    </div>
                                                                    <div class="space-y-1">
                                                                        <Label class="text-[9px] uppercase opacity-60">Static Options</Label>
                                                                        <Input 
                                                                            value={param.options?.map(o => `${o.value}:${o.label}`).join(', ') || ''} 
                                                                            on:input={(e) => {
                                                                                const val = e.currentTarget.value;
                                                                                param.options = val.split(',').filter(s => s.includes(':')).map(s => {
                                                                                    const [v, l] = s.split(':');
                                                                                    return { value: v.trim(), label: l.trim() };
                                                                                });
                                                                                vendors = [...vendors];
                                                                            }}
                                                                            placeholder="v1:l1, v2:l2" 
                                                                            class="h-7 text-[10px]" 
                                                                        />
                                                                    </div>
                                                                </div>
                                                            {/if}

                                                            <div class="space-y-1">
                                                                <Label class="text-[9px] uppercase opacity-60">Config Template</Label>
                                                                <Input bind:value={param.config_template} placeholder="e.g. key[[key_index]]: [[value]]" class="h-7 text-[10px] font-mono" />
                                                            </div>

                                                            <button 
                                                                class="absolute -top-2 -right-2 opacity-0 group-hover:opacity-100 bg-destructive text-white rounded-full p-1 shadow-sm transition-opacity"
                                                                on:click={() => removeParam('features', fi, pi)}
                                                            >
                                                                <X class="w-3 h-3" />
                                                            </button>
                                                        </div>
                                                    {/each}
                                                </div>
                                            </div>
                                        </Card.Content>
                                    </Card.Root>
                                {/each}
                            </div>
                        </div>

                        <!-- Accounts Column -->
                        <div class="space-y-4">
                            <div class="flex justify-between items-center">
                                <h3 class="text-xl font-bold flex items-center gap-2">
                                    <FileCode class="w-5 h-5" />
                                    Account Parameters
                                </h3>
                                <div class="flex gap-2">
                                    <Button variant="outline" size="sm" on:click={() => addFeature('accounts')}>
                                        <Plus class="w-4 h-4 mr-1" /> Add
                                    </Button>
                                    <Button size="sm" on:click={() => saveVendorData('accounts')} disabled={saving}>
                                        <Save class="w-4 h-4 mr-1" /> Save
                                    </Button>
                                </div>
                            </div>

                            <div class="space-y-3">
                                {#each selectedVendor.accounts || [] as feature, fi}
                                    <Card.Root>
                                        <Card.Header class="p-4 pb-2 border-b bg-muted/10">
                                            <div class="flex justify-between items-center">
                                                <div class="space-y-1">
                                                    <h4 class="text-sm font-bold text-primary">{feature.name || "Unnamed Group"}</h4>
                                                    <p class="text-[10px] font-mono opacity-60">ID: {feature.id}</p>
                                                </div>
                                                <Button variant="ghost" size="sm" class="text-destructive h-8 w-8 p-0" on:click={() => removeFeature('accounts', fi)}>
                                                    <Trash2 class="h-4 w-4" />
                                                </Button>
                                            </div>
                                        </Card.Header>
                                        <Card.Content class="p-4 space-y-4">
                                            <div class="grid grid-cols-2 gap-3">
                                                <div class="space-y-1">
                                                    <Label class="text-[10px] uppercase opacity-70">Group ID</Label>
                                                    <Input bind:value={feature.id} placeholder="e.g. basic" class="h-8 font-mono text-xs" />
                                                </div>
                                                <div class="space-y-1">
                                                    <Label class="text-[10px] uppercase opacity-70">Display Name</Label>
                                                    <Input bind:value={feature.name} placeholder="Group Name" class="h-8 text-xs" />
                                                </div>
                                            </div>

                                            <div class="border-t pt-3 mt-2">
                                                <div class="flex justify-between items-center mb-3">
                                                    <span class="text-xs font-bold uppercase text-muted-foreground">Parameters</span>
                                                    <Button variant="outline" size="sm" class="h-7 px-2 text-[10px]" on:click={() => addParam('accounts', fi)}>
                                                        <Plus class="w-3 h-3 mr-1" /> Add Param
                                                    </Button>
                                                </div>
                                                <div class="space-y-3">
                                                    {#each feature.params || [] as param, pi}
                                                        <div class="bg-muted/30 p-3 rounded-md border border-dashed space-y-3 relative group">
                                                            <div class="grid grid-cols-3 gap-2">
                                                                <div class="space-y-1">
                                                                    <Label class="text-[9px] uppercase opacity-60">Param ID</Label>
                                                                    <Input bind:value={param.id} class="h-7 text-[10px] font-mono" />
                                                                </div>
                                                                <div class="space-y-1">
                                                                    <Label class="text-[9px] uppercase opacity-60">Label</Label>
                                                                    <Input bind:value={param.label} class="h-7 text-[10px]" />
                                                                </div>
                                                                <div class="space-y-1">
                                                                    <Label class="text-[9px] uppercase opacity-60">Type</Label>
                                                                    <select bind:value={param.type} class="w-full h-7 text-[10px] rounded border bg-background px-1">
                                                                        <option value="string">String</option>
                                                                        <option value="number">Number</option>
                                                                        <option value="boolean">Boolean</option>
                                                                        <option value="select">Select</option>
                                                                        <option value="password">Password</option>
                                                                        <option value="hidden">Hidden</option>
                                                                    </select>
                                                                </div>
                                                            </div>
                                                            
                                                            {#if param.type === 'select'}
                                                                <div class="grid grid-cols-2 gap-2">
                                                                    <div class="space-y-1">
                                                                        <Label class="text-[9px] uppercase opacity-60">Source</Label>
                                                                        <Input bind:value={param.source} placeholder="Source" class="h-7 text-[10px]" />
                                                                    </div>
                                                                    <div class="space-y-1">
                                                                        <Label class="text-[9px] uppercase opacity-60">Static Options</Label>
                                                                        <Input 
                                                                            value={param.options?.map(o => `${o.value}:${o.label}`).join(', ') || ''} 
                                                                            on:input={(e) => {
                                                                                const val = e.currentTarget.value;
                                                                                param.options = val.split(',').filter(s => s.includes(':')).map(s => {
                                                                                    const [v, l] = s.split(':');
                                                                                    return { value: v.trim(), label: l.trim() };
                                                                                });
                                                                                vendors = [...vendors];
                                                                            }}
                                                                            placeholder="v1:l1, v2:l2" 
                                                                            class="h-7 text-[10px]" 
                                                                        />
                                                                    </div>
                                                                </div>
                                                            {/if}

                                                            <div class="space-y-1">
                                                                <Label class="text-[9px] uppercase opacity-60">Config Template</Label>
                                                                <Input bind:value={param.config_template} placeholder="Config Template" class="h-7 text-[10px] font-mono" />
                                                            </div>

                                                            <button 
                                                                class="absolute -top-2 -right-2 opacity-0 group-hover:opacity-100 bg-destructive text-white rounded-full p-1 shadow-sm transition-opacity"
                                                                on:click={() => removeParam('accounts', fi, pi)}
                                                            >
                                                                <X class="w-3 h-3" />
                                                            </button>
                                                        </div>
                                                    {/each}
                                                </div>
                                            </div>
                                        </Card.Content>
                                    </Card.Root>
                                {/each}
                            </div>
                        </div>

                        <div class="space-y-4 lg:col-span-2 border-t pt-6">
                            <div class="flex justify-between items-center gap-4">
                                <div class="flex items-center gap-2 flex-1">
                                    <FileCode class="w-5 h-5" />
                                    <h3 class="text-xl font-bold whitespace-nowrap">Templates</h3>
                                    <select 
                                        bind:value={selectedTemplateFile}
                                        class="flex-1 max-w-md h-9 px-3 py-1 rounded-md border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                    >
                                        {#each vendorTemplates as file}
                                            <option value={file}>
                                                {file} {file === selectedVendor.phone_config_template ? '(Main)' : ''}
                                            </option>
                                        {/each}
                                    </select>
                                </div>
                                <Button size="sm" on:click={saveTemplate} disabled={saving || !selectedTemplateFile}>
                                    <Save class="w-4 h-4 mr-1" /> Save Template
                                </Button>
                            </div>
                            <textarea 
                                bind:value={templateContent}
                                class="w-full h-[500px] p-4 font-mono text-sm border rounded-md bg-muted/20 focus:outline-none focus:ring-2 focus:ring-ring"
                                placeholder={selectedTemplateFile ? "Loading template..." : "Select a template file to edit"}
                                disabled={!selectedTemplateFile}
                            ></textarea>
                        </div>
                    </div>
                {:else}
                    <div class="flex flex-col items-center justify-center p-12 opacity-50">
                        <FileCode class="w-12 h-12 mb-4" />
                        <p>No vendors found or selected.</p>
                    </div>
                {/if}
            </Tabs.Content>
        </Tabs.Root>
    {/if}
</div>
