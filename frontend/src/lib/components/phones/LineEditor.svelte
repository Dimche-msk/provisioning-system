<script>
    import { createEventDispatcher } from "svelte";
    import { t } from "svelte-i18n";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Table from "$lib/components/ui/table";
    import { Pencil, Trash2, Plus, Search, Save, X } from "lucide-svelte";
    import { toast } from "svelte-sonner";

    export let lines = [];
    export let maxLines = 0;
    export let open = false;

    const dispatch = createEventDispatcher();

    let searchQuery = "";
    let currentPage = 1;
    let itemsPerPage = 16;

    // Editing state
    let editForm = null;

    // Filtered lines
    $: filteredLines = lines.filter((l) => {
        const q = searchQuery.toLowerCase();
        return (
            (l.account_name || "").toLowerCase().includes(q) ||
            (l.phone_number || "").toLowerCase().includes(q)
        );
    });

    $: totalPages = Math.ceil(filteredLines.length / itemsPerPage);
    $: paginatedLines = filteredLines.slice(
        (currentPage - 1) * itemsPerPage,
        currentPage * itemsPerPage,
    );

    // Refined Edit Logic
    let originalLine = null;

    function edit(line) {
        originalLine = line;
        editForm = { ...line };
    }

    function add() {
        if (maxLines > 0 && lines.length >= maxLines) {
            toast.error($t("lines.max_reached") || "Maximum lines reached");
            return;
        }
        originalLine = null;
        editForm = {
            account_name: "",
            phone_number: "",
            caller_id: "",
            account_settings: "{}",
            description: "",
        };
    }

    function save() {
        if (!editForm.account_name || !editForm.phone_number) {
            toast.error("Name and Number are required");
            return;
        }

        if (originalLine) {
            // Update
            const idx = lines.indexOf(originalLine);
            if (idx !== -1) {
                lines[idx] = { ...editForm };
            }
        } else {
            // Create
            lines = [...lines, { ...editForm }];
        }
        originalLine = null;
        editForm = null;
    }

    function remove(line) {
        lines = lines.filter((l) => l !== line);
    }

    function cancelEdit() {
        originalLine = null;
        editForm = null;
    }

    function close() {
        dispatch("close");
    }

    function saveAll() {
        dispatch("save", lines);
        close();
    }
</script>

{#if open}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    >
        <div
            class="bg-background p-6 rounded-lg shadow-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto border"
        >
            <div class="flex justify-between items-center mb-4">
                <div>
                    <h2 class="text-lg font-semibold">
                        {$t("lines.title") || "Line Configuration"}
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

            <div class="space-y-4">
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
                    <div class="border rounded-md p-4 bg-muted/50 space-y-4">
                        <div class="grid grid-cols-2 gap-4">
                            <div class="space-y-2">
                                <Label
                                    >{$t("lines.account_name") ||
                                        "Account Name"}</Label
                                >
                                <Input
                                    bind:value={editForm.account_name}
                                    placeholder="line.2"
                                />
                            </div>
                            <div class="space-y-2">
                                <Label>{$t("phone.number") || "Number"}</Label>
                                <Input
                                    bind:value={editForm.phone_number}
                                    placeholder="102"
                                />
                            </div>
                            <div class="space-y-2">
                                <Label
                                    >{$t("phone.caller_id") ||
                                        "Caller ID"}</Label
                                >
                                <Input
                                    bind:value={editForm.caller_id}
                                    placeholder="John Doe"
                                />
                            </div>
                            <div class="space-y-2">
                                <Label>{$t("phone.domain") || "Domain"}</Label>
                                <Input
                                    bind:value={editForm.domain}
                                    placeholder="sip.example.com"
                                />
                            </div>
                        </div>
                        <div class="space-y-2">
                            <Label
                                >{$t("phone.settings") ||
                                    "Settings (JSON)"}</Label
                            >
                            <Input bind:value={editForm.account_settings} />
                        </div>
                        <div class="space-y-2">
                            <Label
                                >{$t("common.description") ||
                                    "Description"}</Label
                            >
                            <Input bind:value={editForm.description} />
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
                                <Save class="mr-2 h-4 w-4" />
                                {$t("common.save") || "Save"}
                            </Button>
                        </div>
                    </div>
                {/if}

                <!-- Table -->
                <div class="border rounded-md">
                    <Table.Root>
                        <Table.Header>
                            <Table.Row>
                                <Table.Head
                                    >{$t("lines.account_name") ||
                                        "Name"}</Table.Head
                                >
                                <Table.Head
                                    >{$t("phone.number") ||
                                        "Number"}</Table.Head
                                >
                                <Table.Head
                                    >{$t("phone.caller_id") ||
                                        "Caller ID"}</Table.Head
                                >
                                <Table.Head
                                    >{$t("phone.domain") ||
                                        "Domain"}</Table.Head
                                >
                                <Table.Head class="text-right"
                                    >Actions</Table.Head
                                >
                            </Table.Row>
                        </Table.Header>
                        <Table.Body>
                            {#each paginatedLines as line}
                                <Table.Row>
                                    <Table.Cell>{line.account_name}</Table.Cell>
                                    <Table.Cell>{line.phone_number}</Table.Cell>
                                    <Table.Cell>{line.caller_id}</Table.Cell>
                                    <Table.Cell>{line.domain}</Table.Cell>
                                    <Table.Cell class="text-right">
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            on:click={() => edit(line)}
                                            disabled={!!editForm}
                                        >
                                            <Pencil class="h-4 w-4" />
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            class="text-destructive"
                                            on:click={() => remove(line)}
                                            disabled={!!editForm}
                                        >
                                            <Trash2 class="h-4 w-4" />
                                        </Button>
                                    </Table.Cell>
                                </Table.Row>
                            {/each}
                            {#if paginatedLines.length === 0}
                                <Table.Row>
                                    <Table.Cell
                                        colspan="5"
                                        class="text-center text-muted-foreground"
                                    >
                                        {$t("common.no_results") ||
                                            "No lines found"}
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
                            Previous
                        </Button>
                        <span class="py-2 text-sm"
                            >Page {currentPage} of {totalPages}</span
                        >
                        <Button
                            variant="outline"
                            size="sm"
                            disabled={currentPage === totalPages}
                            on:click={() => currentPage++}
                        >
                            Next
                        </Button>
                    </div>
                {/if}
            </div>

            <div class="flex justify-end gap-2 mt-4">
                <Button variant="outline" on:click={close}
                    >{$t("common.cancel") || "Cancel"}</Button
                >
                <Button on:click={saveAll}
                    >{$t("common.save_all") || "Save All"}</Button
                >
            </div>
        </div>
    </div>
{/if}
