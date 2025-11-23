<script>
    import { createEventDispatcher, onMount } from "svelte";
    import { t } from "svelte-i18n";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card/index";
    import { toast } from "svelte-sonner";
    import LineEditor from "./LineEditor.svelte";
    import { Settings } from "lucide-svelte";

    export let phone = {
        domain: "",
        vendor: "",
        model_id: "",
        mac_address: "",
        phone_number: "",
        ip_address: "",
        caller_id: "",
        account_settings: "{}",
        description: "",
        lines: [],
        expansion_module_model: "",
    };
    export let mode = "create"; // create | edit

    const dispatch = createEventDispatcher();

    let domains = [];
    let vendors = [];
    let models = [];
    let loading = false;
    let showLineEditor = false;

    // Computed
    $: selectedModel = models.find((m) => m.id === phone.model_id);
    $: maxLines = selectedModel?.max_additional_lines || 0;
    $: canConfigureLines = maxLines > 0;

    onMount(async () => {
        await Promise.all([loadDomains(), loadVendors()]);
        if (phone.vendor) {
            await loadModels(phone.vendor);
        }
    });

    async function loadDomains() {
        try {
            const res = await fetch("/api/domains");
            if (res.ok) {
                const data = await res.json();
                domains = data.domains || [];
                if (mode === "create" && domains.length > 0 && !phone.domain) {
                    phone.domain = domains[0];
                }
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
                if (mode === "create" && vendors.length > 0 && !phone.vendor) {
                    phone.vendor = vendors[0].id;
                    await loadModels(phone.vendor);
                }
            }
        } catch (e) {
            console.error("Failed to load vendors", e);
        }
    }

    async function loadModels(vendorId) {
        try {
            const res = await fetch(`/api/models?vendor=${vendorId}`);
            if (res.ok) {
                const data = await res.json();
                models = data.models || [];
                if (mode === "create" && models.length > 0 && !phone.model_id) {
                    phone.model_id = models[0].id;
                    console.log("loadModels - mode -> Create", phone.model_id);
                }
            }
        } catch (e) {
            console.error("Failed to load models", e);
            models = [];
        }
    }

    function onVendorChange() {
        phone.model_id = "";
        if (phone.vendor) {
            loadModels(phone.vendor);
        } else {
            models = [];
        }
    }

    async function save() {
        loading = true;
        try {
            // Validate JSON
            try {
                JSON.parse(phone.account_settings);
            } catch (e) {
                toast.error(
                    $t("phone.invalid_json") || "Invalid JSON settings",
                );
                loading = false;
                return;
            }

            const url =
                mode === "create" ? "/api/phones" : `/api/phones/${phone.id}`;
            const method = mode === "create" ? "POST" : "PUT";

            const res = await fetch(url, {
                method: method,
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(phone),
            });

            if (res.ok) {
                const savedPhone = await res.json();
                toast.success(
                    mode === "create"
                        ? $t("phone.create_success") ||
                              "Phone created successfully"
                        : $t("phone.update_success") ||
                              "Phone updated successfully",
                );
                dispatch("save", savedPhone);
                if (mode === "create") {
                    // Reset form
                    phone = {
                        ...phone,
                        mac_address: "",
                        phone_number: "",
                        ip_address: "",
                        caller_id: "",
                        account_settings: "{}",
                        description: "",
                        lines: [],
                    };
                }
            } else {
                const text = await res.text();
                toast.error(text || "Failed to save phone");
            }
        } catch (e) {
            toast.error("Error: " + e.message);
        } finally {
            loading = false;
        }
    }

    function handleLinesSave(e) {
        phone.lines = e.detail;
        showLineEditor = false;
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title>
            {mode === "create"
                ? $t("phone.add_title") || "Add Phone Configuration"
                : $t("phone.edit_title") || "Edit Phone Configuration"}
        </Card.Title>
        <Card.Description>
            {mode === "create"
                ? $t("phone.add_desc") ||
                  "Manually add a new phone configuration"
                : $t("phone.edit_desc") || "Edit existing phone configuration"}
        </Card.Description>
    </Card.Header>
    <Card.Content class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
                <Label for="domain">{$t("phone.domain") || "Domain"}</Label>
                <select
                    id="domain"
                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    bind:value={phone.domain}
                >
                    {#each domains as d}
                        <option value={d}>{d}</option>
                    {/each}
                </select>
            </div>
            <div class="space-y-2">
                <Label for="vendor">{$t("phone.vendor") || "Vendor"}</Label>
                <select
                    id="vendor"
                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    bind:value={phone.vendor}
                    on:change={onVendorChange}
                >
                    {#each vendors as v}
                        <option value={v.id}>{v.name}</option>
                    {/each}
                </select>
            </div>
        </div>

        <div class="space-y-2">
            <Label for="model">{$t("phone.model") || "Model"}</Label>
            <select
                id="model"
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                bind:value={phone.model_id}
                disabled={models.length === 0}
            >
                {#if models.length === 0}
                    <option value="">No models available</option>
                {/if}
                {#each models as m}
                    <option value={m.id}>{m.name}</option>
                {/each}
            </select>
        </div>

        <!-- Expansion Module Selection (if supported) -->
        {#if selectedModel?.supported_expansion_modules?.length > 0}
            <div class="space-y-2">
                <Label for="exp_module"
                    >{$t("phone.expansion_module") || "Expansion Module"}</Label
                >
                <select
                    id="exp_module"
                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    bind:value={phone.expansion_module_model}
                >
                    <option value="">None</option>
                    {#each selectedModel.supported_expansion_modules as m}
                        <option value={m}>{m}</option>
                    {/each}
                </select>
            </div>
        {/if}

        <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
                <Label for="mac">
                    {$t("phone.mac") || "MAC Address"}
                    {#if selectedModel?.type !== "gateway"}
                        <span class="text-red-500">*</span>
                    {/if}
                </Label>
                <Input
                    id="mac"
                    bind:value={phone.mac_address}
                    placeholder="00:11:22:33:44:55"
                />
            </div>
            <div class="space-y-2">
                <Label for="number">
                    {$t("phone.number") || "Phone Number"}
                    {#if selectedModel?.type !== "gateway"}
                        <span class="text-red-500">*</span>
                    {/if}
                </Label>
                <Input
                    id="number"
                    bind:value={phone.phone_number}
                    placeholder="101"
                />
            </div>
        </div>

        <div class="space-y-2">
            <Label for="ip_address"
                >{$t("phone.ip_address") || "IP Address"}</Label
            >
            <Input
                id="ip_address"
                bind:value={phone.ip_address}
                placeholder="192.168.1.100"
            />
        </div>

        <div class="space-y-2">
            <Label for="callerid">{$t("phone.caller_id") || "Caller ID"}</Label>
            <Input
                id="callerid"
                bind:value={phone.caller_id}
                placeholder="John Doe"
            />
        </div>

        <div class="space-y-2">
            <Label for="settings"
                >{$t("phone.settings") || "Account Settings (JSON)"}</Label
            >
            <textarea
                id="settings"
                class="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                bind:value={phone.account_settings}
            ></textarea>
        </div>

        <div class="space-y-2">
            <Label for="description"
                >{$t("common.description") || "Description"}</Label
            >
            <Input id="description" bind:value={phone.description} />
        </div>

        <div class="flex gap-4">
            <Button class="flex-1" on:click={save} disabled={loading}>
                {loading
                    ? $t("common.saving") || "Saving..."
                    : $t("common.save") || "Save"}
            </Button>

            {#if canConfigureLines}
                <Button
                    variant="outline"
                    on:click={() => (showLineEditor = true)}
                >
                    <Settings class="mr-2 h-4 w-4" />
                    {$t("lines.configure") || "Configure Lines"}
                    {#if phone.lines?.length > 0}
                        ({phone.lines.length})
                    {/if}
                </Button>
            {/if}
        </div>
    </Card.Content>
</Card.Root>

<LineEditor
    bind:open={showLineEditor}
    lines={phone.lines || []}
    {maxLines}
    on:save={handleLinesSave}
    on:close={() => (showLineEditor = false)}
/>
