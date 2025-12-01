<script lang="ts">
    import { onMount } from "svelte";
    import { t } from "svelte-i18n";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import * as Table from "$lib/components/ui/table";
    import * as Card from "$lib/components/ui/card";
    import { Badge } from "$lib/components/ui/badge";
    import { toast } from "svelte-sonner";
    import {
        Upload,
        Check,
        AlertTriangle,
        X,
        Loader2,
        ArrowLeft,
    } from "lucide-svelte";
    import * as XLSX from "xlsx";
    import type {
        Phone,
        Vendor,
        DeviceModel,
        DomainSettings,
    } from "$lib/types";

    let files: FileList;
    let processing = false;
    let importLog: any[] = [];
    let stats = {
        total: 0,
        success: 0,
        error: 0,
        skipped: 0,
    };

    // Data caches
    let existingPhones: Phone[] = [];
    let vendors: Vendor[] = [];
    let models: DeviceModel[] = [];
    let domains: string[] = [];
    let domainConfigs: DomainSettings[] = [];

    // Mapped data from Excel
    let parsedRows: any[] = [];
    let validatedRows: any[] = [];

    onMount(async () => {
        await Promise.all([loadPhones(), loadVendors(), loadDomains()]);
        // Models are loaded per vendor, but we might need all models to map by name.
        // Or we can load models for all vendors.
        await loadAllModels();
    });

    async function loadPhones() {
        try {
            const res = await fetch("/api/phones");
            if (res.ok) {
                const data = await res.json();
                existingPhones = data.phones || [];
            }
        } catch (e) {
            console.error("Failed to load phones", e);
        }
    }

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

    async function loadDomains() {
        try {
            const res = await fetch("/api/domains");
            if (res.ok) {
                const data = await res.json();
                domains = data.domains || [];
                domainConfigs = data.detailed_domains || [];
            }
        } catch (e) {
            console.error("Failed to load domains", e);
        }
    }

    async function loadAllModels() {
        // This might be inefficient if there are many vendors, but necessary for mapping by name
        let allModels: DeviceModel[] = [];
        for (const v of vendors) {
            try {
                const res = await fetch(`/api/models?vendor=${v.id}`);
                if (res.ok) {
                    const data = await res.json();
                    allModels = [...allModels, ...(data.models || [])];
                }
            } catch (e) {
                console.error(`Failed to load models for ${v.name}`, e);
            }
        }
        models = allModels;
    }

    function handleFileSelect(e: Event) {
        const target = e.target as HTMLInputElement;
        if (target.files && target.files.length > 0) {
            parseFile(target.files[0]);
        }
    }

    async function parseFile(file: File) {
        processing = true;
        parsedRows = [];
        validatedRows = [];
        importLog = [];
        stats = { total: 0, success: 0, error: 0, skipped: 0 };

        const reader = new FileReader();
        reader.onload = (e) => {
            try {
                const data = new Uint8Array(e.target?.result as ArrayBuffer);
                const workbook = XLSX.read(data, { type: "array" });
                const firstSheetName = workbook.SheetNames[0];
                const worksheet = workbook.Sheets[firstSheetName];
                const jsonData = XLSX.utils.sheet_to_json(worksheet, {
                    header: 1,
                });

                // Assume header is row 0
                const headers = (jsonData[0] as string[]).map((h) =>
                    h.toLowerCase().trim(),
                );
                const rows = jsonData.slice(1);

                // Map columns
                const colMap = {
                    number: headers.findIndex(
                        (h) => h.includes("number") || h.includes("номер"),
                    ),
                    mac: headers.findIndex(
                        (h) => h.includes("mac") || h.includes("мак"),
                    ),
                    user: headers.findIndex(
                        (h) =>
                            h.includes("user") ||
                            h.includes("имя") ||
                            h.includes("name"),
                    ),
                    vendor: headers.findIndex(
                        (h) =>
                            h.includes("vendor") || h.includes("производитель"),
                    ),
                    model: headers.findIndex(
                        (h) => h.includes("model") || h.includes("модель"),
                    ),
                    domain: headers.findIndex(
                        (h) => h.includes("domain") || h.includes("домен"),
                    ),
                    login: headers.findIndex(
                        (h) => h.includes("login") || h.includes("логин"),
                    ),
                    password: headers.findIndex(
                        (h) => h.includes("password") || h.includes("пароль"),
                    ),
                };

                parsedRows = rows.map((row: any, index) => {
                    return {
                        id: index,
                        number: row[colMap.number],
                        mac: row[colMap.mac],
                        user: row[colMap.user],
                        vendor: row[colMap.vendor],
                        model: row[colMap.model],
                        domain: row[colMap.domain],
                        login: row[colMap.login],
                        password: row[colMap.password],
                        status: "pending", // pending, valid, error, conflict
                        message: "",
                        phoneData: null,
                    };
                });

                validateRows();
            } catch (err) {
                toast.error("Failed to parse Excel file");
                console.error(err);
            } finally {
                processing = false;
            }
        };
        reader.readAsArrayBuffer(file);
    }

    function validateRows() {
        validatedRows = parsedRows.map((row) => {
            const errors: string[] = [];

            // Basic Validation
            if (!row.mac) errors.push("MAC is required");
            if (!row.number) errors.push("Number is required");
            if (!row.vendor) errors.push("Vendor is required");
            if (!row.model) errors.push("Model is required");

            // Data Mapping
            let vendorId = "";
            let modelId = "";
            let domain = row.domain;

            // Map Vendor
            if (row.vendor) {
                const v = vendors.find(
                    (v) =>
                        v.name.toLowerCase() ===
                            String(row.vendor).toLowerCase() ||
                        v.id === row.vendor,
                );
                if (v) {
                    vendorId = v.id;
                } else {
                    errors.push(`Vendor '${row.vendor}' not found`);
                }
            }

            // Map Model
            if (row.model && vendorId) {
                const m = models.find(
                    (m) =>
                        m.vendor === vendorId &&
                        (m.name.toLowerCase() ===
                            String(row.model).toLowerCase() ||
                            m.id === row.model),
                );
                if (m) {
                    modelId = m.id;
                } else {
                    errors.push(`Model '${row.model}' not found for vendor`);
                }
            }

            // Map Domain
            if (!domain && domains.length > 0) {
                domain = domains[0]; // Default to first
            } else if (domain && !domains.includes(domain)) {
                // Maybe it's a valid domain but not in our list? Or just warn?
                // We'll assume strict check
                if (
                    !domains.find(
                        (d) => d.toLowerCase() === String(domain).toLowerCase(),
                    )
                ) {
                    errors.push(`Domain '${domain}' not found`);
                }
            }

            // Get Domain Config
            const domainConfig = domainConfigs.find((d) => d.name === domain);
            const domainVars = domainConfig?.variables || {};

            // Check Duplicates
            let conflictPhone: Phone | undefined;
            if (row.mac) {
                conflictPhone = existingPhones.find(
                    (p) =>
                        p.mac_address?.toLowerCase() ===
                        String(row.mac).toLowerCase(),
                );
                if (conflictPhone) {
                    // Conflict
                }
            }
            if (row.number && !conflictPhone) {
                // Check number conflict too?
                // existingPhones doesn't strictly enforce unique number, but usually it's unique per domain.
                // Let's stick to MAC conflict as primary.
            }

            let status = "valid";
            let message = "";

            if (errors.length > 0) {
                status = "error";
                message = errors.join(", ");
            } else if (conflictPhone) {
                status = "conflict";
                message = `Phone with MAC ${row.mac} already exists.`;
            }

            // Construct Phone Object
            const phoneData: Phone = {
                id: conflictPhone?.id, // Undefined if not updating, so omitted in JSON
                mac_address: String(row.mac),
                phone_number: String(row.number),
                description: row.user || "",
                vendor: vendorId,
                model_id: modelId,
                domain: domain,
                ip_address: conflictPhone?.ip_address || "", // Keep existing IP if update
                type: "phone", // Default
                lines: [],
                expansion_module_model: "", // Initialize required field
            };

            // Construct Line
            const line = {
                type: "line",
                number: 1,
                expansion_module_number: 0,
                key_number: 0,
                additional_info: JSON.stringify({
                    display_name: row.user || "",
                    user_name: row.login || row.number, // Fallback to number if login missing
                    auth_name: row.login || row.number,
                    password: row.password || domainVars.sip_password || "",
                    line_number: String(row.number), // Use phone number
                    screen_name: row.user || "", // Use user name
                    registrar1_ip: domainVars.sip_server_ip || "",
                }),
            };
            phoneData.lines = [line];

            return {
                ...row,
                status,
                message,
                phoneData,
                conflictId: conflictPhone?.id,
            };
        });

        stats.total = validatedRows.length;
    }

    async function importRow(row: any, overwrite = false) {
        if (row.status === "error") return;

        // If conflict and not overwrite, skip or ask?
        // This function is called when user clicks "Import" or "Overwrite"

        const method = row.conflictId ? "PUT" : "POST";
        const url = row.conflictId
            ? `/api/phones/${row.conflictId}`
            : "/api/phones";

        try {
            const res = await fetch(url, {
                method: method,
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(row.phoneData),
            });

            if (res.ok) {
                row.status = "success";
                row.message = "Imported successfully";
                stats.success++;
            } else {
                const text = await res.text();
                row.status = "error";
                row.message = "API Error: " + text;
                stats.error++;
            }
        } catch (e: any) {
            row.status = "error";
            row.message = "Network Error: " + e.message;
            stats.error++;
        }

        // Trigger reactivity
        validatedRows = [...validatedRows];
    }

    async function processAll() {
        processing = true;
        stats.success = 0;
        stats.error = 0;

        for (const row of validatedRows) {
            if (row.status === "valid") {
                await importRow(row);
            }
        }
        processing = false;
    }
</script>

<div class="h-full flex flex-col p-6 space-y-4 overflow-hidden">
    <div class="flex justify-between items-center shrink-0">
        <div class="flex items-center gap-4">
            <Button variant="ghost" size="icon" href="/phones">
                <ArrowLeft class="h-4 w-4" />
            </Button>
            <h1 class="text-3xl font-bold tracking-tight">
                {$t("phones.import") || "Import Phones"}
            </h1>
        </div>
    </div>

    <Card.Root>
        <Card.Content class="p-6 space-y-4">
            <div class="flex items-center gap-4">
                <Input
                    type="file"
                    accept=".xlsx, .xls"
                    on:change={handleFileSelect}
                    disabled={processing}
                />
                {#if validatedRows.length > 0}
                    <Button on:click={processAll} disabled={processing}>
                        {#if processing}
                            <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                        {/if}
                        Import Valid Rows
                    </Button>
                {/if}
            </div>

            {#if validatedRows.length > 0}
                <div class="flex gap-4 text-sm">
                    <Badge variant="outline">Total: {stats.total}</Badge>
                    <Badge variant="default" class="bg-green-500"
                        >Success: {stats.success}</Badge
                    >
                    <Badge variant="destructive">Error: {stats.error}</Badge>
                </div>
            {/if}
        </Card.Content>
    </Card.Root>

    {#if validatedRows.length > 0}
        <div
            class="flex-1 border rounded-md overflow-hidden flex flex-col bg-background"
        >
            <div class="flex-1 overflow-y-auto">
                <Table.Root>
                    <Table.Header class="sticky top-0 bg-background z-10">
                        <Table.Row>
                            <Table.Head>Status</Table.Head>
                            <Table.Head>MAC</Table.Head>
                            <Table.Head>Number</Table.Head>
                            <Table.Head>Vendor/Model</Table.Head>
                            <Table.Head>User</Table.Head>
                            <Table.Head>Message</Table.Head>
                            <Table.Head class="text-right">Actions</Table.Head>
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {#each validatedRows as row}
                            <Table.Row>
                                <Table.Cell>
                                    {#if row.status === "success"}
                                        <Check class="h-4 w-4 text-green-500" />
                                    {:else if row.status === "error"}
                                        <X class="h-4 w-4 text-red-500" />
                                    {:else if row.status === "conflict"}
                                        <AlertTriangle
                                            class="h-4 w-4 text-yellow-500"
                                        />
                                    {:else}
                                        <Badge variant="secondary">Ready</Badge>
                                    {/if}
                                </Table.Cell>
                                <Table.Cell>{row.mac}</Table.Cell>
                                <Table.Cell>{row.number}</Table.Cell>
                                <Table.Cell
                                    >{row.vendor} / {row.model}</Table.Cell
                                >
                                <Table.Cell>{row.user}</Table.Cell>
                                <Table.Cell
                                    class="text-sm text-muted-foreground"
                                >
                                    {row.message}
                                </Table.Cell>
                                <Table.Cell class="text-right">
                                    {#if row.status === "conflict"}
                                        <Button
                                            size="sm"
                                            variant="outline"
                                            on:click={() =>
                                                importRow(row, true)}
                                        >
                                            Overwrite
                                        </Button>
                                    {:else if row.status === "valid"}
                                        <Button
                                            size="sm"
                                            variant="ghost"
                                            on:click={() => importRow(row)}
                                        >
                                            Import
                                        </Button>
                                    {/if}
                                </Table.Cell>
                            </Table.Row>
                        {/each}
                    </Table.Body>
                </Table.Root>
            </div>
        </div>
    {/if}
</div>
