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
    import CodeMirror from "svelte-codemirror-editor";
    import { yaml } from "@codemirror/lang-yaml";
    import { oneDark } from "@codemirror/theme-one-dark";

    let rawYaml = "";
    let loading = true;
    let saving = false;
    let activeTab = "raw";
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
            const rawRes = await fetch("/api/system/config/raw");

            if (rawRes.ok) {
                rawYaml = await rawRes.text();
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



    async function saveRawYaml() {
        if (!confirm("Save raw YAML? This will apply your exact text and overwrite settings.")) return;

        saving = true;
        try {
            const res = await fetch("/api/system/config", {
                method: "POST",
                headers: { "Content-Type": "text/yaml" },
                body: rawYaml
            });

            if (res.ok) {
                toast.success($t("templates.save_success"));
                await loadData();
            } else {
                let errorMsg = $t("templates.save_error");
                try {
                    const data = await res.json();
                    errorMsg = data.error || errorMsg;
                } catch {
                    errorMsg = await res.text();
                }
                toast.error(errorMsg);
            }
        } catch (e: any) {
            toast.error($t("templates.save_error") + ": " + e.message);
        } finally {
            saving = false;
        }
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
        {#if activeTab === "raw"}
            <Button on:click={saveRawYaml} disabled={saving || loading}>
                <Save class="mr-2 h-4 w-4" />
                {saving ? $t("common.saving") : $t("templates.save_button")}
            </Button>
        {/if}
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
    {:else}
        <Tabs.Root bind:value={activeTab} class="w-full">
            <Tabs.List class="grid w-full grid-cols-2 max-w-xl">
                <Tabs.Trigger value="raw">{$t("templates.system_config")}</Tabs.Trigger>
                <Tabs.Trigger value="functions">{$t("templates.function_settings")}</Tabs.Trigger>
            </Tabs.List>

            <Tabs.Content value="raw" class="pt-6 space-y-4">
                <div class="flex justify-between items-center mb-2">
                    <h2 class="text-2xl font-bold">{$t("templates.system_config")}</h2>
                </div>
                <Alert variant="destructive">
                    <AlertCircle class="h-4 w-4" />
                    <AlertTitle>{$t("common.notice") || "Notice"}</AlertTitle>
                    <AlertDescription>
                        {$t("templates.yaml_disclaimer")}
                    </AlertDescription>
                </Alert>
                <div class="border rounded-md overflow-hidden bg-[#282c34]">
                    <CodeMirror bind:value={rawYaml} lang={yaml()} theme={oneDark} styles={{ "&": { minHeight: "600px", fontSize: "14px", backgroundColor: "#282c34" } }} />
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
