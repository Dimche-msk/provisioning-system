<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { t } from "svelte-i18n";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card/index";
    import { toast } from "svelte-sonner";
    import LineEditor from "./LineEditor.svelte";
    import { Settings } from "lucide-svelte";
    import type { Phone, DeviceModel, Vendor } from "$lib/types";
    import { formatMacInput } from "$lib/utils";

    export let phone: Phone = {
        domain: "",
        vendor: "",
        model_id: "",
        mac_address: "",
        phone_number: "",
        ip_address: "",
        description: "",
        lines: [],
        expansion_module_model: "",
        type: "phone",
    };
    export let mode = "create"; // create | edit

    const dispatch = createEventDispatcher();

    let domains: string[] = [];
    let vendors: Vendor[] = [];
    let models: DeviceModel[] = [];
    let loading = false;
    const formId = Math.random().toString(36).slice(2);
    let showLineEditor = false;

    // Computed
    $: selectedModel = models.find((m) => m.id === phone.model_id);
    $: maxLines = selectedModel?.max_account_lines || 0;
    $: canConfigureLines = maxLines > 0;

    $: maxSoftKeys = selectedModel?.own_soft_keys || 0;
    $: maxHardKeys = selectedModel?.own_hard_keys || 0;

    // Auto-format MAC address reactively
    $: if (phone.mac_address) {
        const formatted = formatMacInput(phone.mac_address);
        if (formatted !== phone.mac_address) {
            phone.mac_address = formatted;
        }
    }

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

                // Force reactivity for vendor select
                if (phone.vendor) {
                    const currentVendor = phone.vendor;
                    // Check if vendor exists in list
                    if (vendors.find((v) => v.id === currentVendor)) {
                        // Trigger update
                        phone.vendor = currentVendor;
                    }
                }

                if (mode === "create" && vendors.length > 0 && !phone.vendor) {
                    phone.vendor = vendors[0].id;
                    await loadModels(phone.vendor);
                }
            }
        } catch (e) {
            console.error("Failed to load vendors", e);
        }
    }

    async function loadModels(vendorId: string) {
        try {
            const res = await fetch(`/api/models?vendor=${vendorId}`);
            if (res.ok) {
                const data = await res.json();
                models = (data.models || [])
                    .filter((m: DeviceModel) => m.type !== "expansion-module")
                    .sort((a: DeviceModel, b: DeviceModel) =>
                        a.name.localeCompare(b.name),
                    );

                console.log("Loaded models:", models);

                // Validate current model_id
                if (phone.model_id) {
                    const found = models.find((m) => m.id === phone.model_id);
                    if (!found) {
                        console.warn(
                            `Model ${phone.model_id} not found in list. Resetting.`,
                        );
                        phone.model_id = "";
                    } else {
                        // Force UI update to ensure correct option is selected
                        const current = phone.model_id;
                        phone.model_id = "";
                        setTimeout(() => {
                            phone.model_id = current;
                        }, 0);
                    }
                }

                if (models.length > 0 && !phone.model_id) {
                    phone.model_id = models[0].id;
                    console.log(
                        "loadModels - Defaulting to first model",
                        phone.model_id,
                    );
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

    function onModelChange(e: Event) {
        const target = e.target as HTMLSelectElement;
        console.log("Model changed to:", target.value);
        phone.model_id = target.value;
    }

    async function save() {
        // Validation for expansion modules count
        if (phone.expansion_module_model && selectedModel) {
            const count = parseInt(String(phone.expansion_modules_count), 10);
            const maxAllowed = selectedModel.maximum_expansion_modules || 0;
            if (count > maxAllowed) {
                toast.error($t("phone.error_max_exp_modules", { values: { max: maxAllowed } }) || `Maximum expansion modules allowed: ${maxAllowed}`);
                return;
            }
        }

        // Ensure numeric types
        if (phone.expansion_modules_count) {
            phone.expansion_modules_count = parseInt(
                String(phone.expansion_modules_count),
                10,
            );
        }

        console.log("Saving phone:", JSON.parse(JSON.stringify(phone)));
        loading = true;
        try {
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
                        description: "",
                        lines: [],
                    };
                }
            } else {
                const text = await res.text();
                toast.error(text || "Failed to save phone");
            }
        } catch (e: any) {
            toast.error("Error: " + e.message);
        } finally {
            loading = false;
        }
    }

    function handleLinesSave(e: CustomEvent) {
        phone.lines = e.detail;
        showLineEditor = false;
    }
    function onPhoneNumberChange() {
        if (
            mode === "create" &&
            phone.phone_number &&
            (!phone.lines || phone.lines.length === 0)
        ) {
            phone.lines = [
                {
                    type: "Line",
                    account_number: 1,
                    panel_number: 0,
                    key_number: 1, // Default to 1 for first account key?
                    additional_info: JSON.stringify({
                        line_number: "1",
                        display_name: phone.phone_number,
                        user_name: phone.phone_number,
                        auth_name: phone.phone_number,
                        password: "",
                        screen_name: phone.phone_number,
                    }),
                },
            ];
            toast.info("Line 1 automatically created");
        }
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title>
            {mode === "create"
                ? $t("phone.add_title") || "Add Phone Configuration"
                : ($t("phone.edit_title") || "Edit Phone Configuration") + (phone.phone_number ? ` : ${phone.phone_number}` : "")}
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
                <Label for="domain-{formId}"
                    >{$t("phone.domain") || "Domain"}</Label
                >
                <select
                    id="domain-{formId}"
                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    bind:value={phone.domain}
                >
                    {#each domains as d}
                        <option value={d}>{d}</option>
                    {/each}
                </select>
            </div>
            <div class="space-y-2">
                <Label for="vendor-{formId}"
                    >{$t("phone.vendor") || "Vendor"}</Label
                >
                <select
                    id="vendor-{formId}"
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

        <div class="grid grid-cols-3 gap-4">
            <div class="space-y-2">
                <Label for="model-{formId}"
                    >{$t("phone.model") || "Model"}</Label
                >
                <select
                    id="model-{formId}"
                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    value={phone.model_id}
                    on:change={onModelChange}
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

            <div class="space-y-2">
                <Label for="exp_module-{formId}"
                    >{$t("phone.expansion_module") || "Expansion Module"}</Label
                >
                <select
                    id="exp_module-{formId}"
                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    bind:value={phone.expansion_module_model}
                    disabled={(selectedModel?.supported_expansion_modules?.length ||
                        0) === 0}
                >
                    <option value="">None</option>
                    {#each selectedModel?.supported_expansion_modules || [] as m}
                        <option value={m}>{m}</option>
                    {/each}
                </select>
            </div>

            <div class="space-y-2">
                <Label for="exp_count-{formId}"
                    >{$t("phone.exp_count") || "Count"}</Label
                >
                <Input
                    id="exp_count-{formId}"
                    type="number"
                    min="0"
                    max={selectedModel?.maximum_expansion_modules || 1}
                    bind:value={phone.expansion_modules_count}
                    disabled={!phone.expansion_module_model}
                />
            </div>
        </div>

        <div class="grid grid-cols-3 gap-4">
            <div class="space-y-2">
                <Label for="mac-{formId}">
                    {$t("phone.mac") || "MAC Address"}
                    {#if selectedModel?.type !== "gateway"}
                        <span class="text-red-500">*</span>
                    {/if}
                </Label>
                <Input
                    id="mac-{formId}"
                    bind:value={phone.mac_address}
                    placeholder="00:11:22:33:44:55"
                    class="font-mono uppercase"
                />
            </div>
            <div class="space-y-2">
                <Label for="number-{formId}">
                    {$t("phone.number") || "Phone Number"}
                    {#if selectedModel?.type !== "gateway"}
                        <span class="text-red-500">*</span>
                    {/if}
                </Label>
                <Input
                    id="number-{formId}"
                    bind:value={phone.phone_number}
                    placeholder="101"
                    on:blur={onPhoneNumberChange}
                />
            </div>
            <div class="space-y-2">
                <Label for="ip_address-{formId}"
                    >{$t("phone.ip_address") || "IP Address"}</Label
                >
                <Input
                    id="ip_address-{formId}"
                    bind:value={phone.ip_address}
                    placeholder="192.168.1.100"
                />
            </div>
        </div>

        <div class="space-y-2">
            <Label for="description-{formId}"
                >{$t("common.description") || "Description"}</Label
            >
            <Input id="description-{formId}" bind:value={phone.description} />
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
    {maxSoftKeys}
    {maxHardKeys}
    image={selectedModel?.image}
    model={selectedModel}
    {phone}
    on:save={handleLinesSave}
    on:close={() => (showLineEditor = false)}
/>
