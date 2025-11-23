<script>
    import { onMount, onDestroy, tick } from "svelte";
    import { t } from "svelte-i18n";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import * as Table from "$lib/components/ui/table";
    import * as Card from "$lib/components/ui/card";
    import {
        Trash2,
        Search,
        Pause,
        Play,
        Filter,
        ChevronDown,
        ChevronUp,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import { Badge } from "$lib/components/ui/badge";

    let logs = [];
    let eventSource;
    let isPaused = false;
    let searchQuery = "";
    let isFiltersOpen = false;

    // Filters
    let filterIP = "";
    let filterCode = "";
    let filterFile = "";

    let tableContainer;

    // Computed
    $: activeFiltersCount = [filterIP, filterCode, filterFile].filter(
        Boolean,
    ).length;
    $: hasActiveFilters = activeFiltersCount > 0;

    onMount(() => {
        connect();
    });

    onDestroy(() => {
        if (eventSource) {
            eventSource.close();
        }
    });

    function connect() {
        eventSource = new EventSource("/api/debug/logs");

        eventSource.onmessage = async (event) => {
            if (event.data === "connected") {
                console.log("Connected to log stream");
                return;
            }
            if (isPaused) return;

            try {
                const logEntry = JSON.parse(event.data);
                // Prepend new logs (newest at top)
                logs = [logEntry, ...logs].slice(0, 1000);

                // If we want auto-scroll to top (which happens automatically with prepend if scrolled to top)
                // If we wanted append (newest at bottom), we would scroll to bottom here.
                // Since user asked for "show latest", newest at top is best.
            } catch (e) {
                console.error("Failed to parse log entry", e);
            }
        };

        eventSource.onerror = (err) => {
            console.error("EventSource failed:", err);
            eventSource.close();
            setTimeout(connect, 5000);
        };
    }

    function clearLogs() {
        logs = [];
    }

    function togglePause() {
        isPaused = !isPaused;
    }

    function toggleFilters() {
        isFiltersOpen = !isFiltersOpen;
    }

    $: filteredLogs = logs.filter((log) => {
        if (searchQuery) {
            const q = searchQuery.toLowerCase();
            const match =
                log.source_ip.toLowerCase().includes(q) ||
                log.requested_file.toLowerCase().includes(q) ||
                log.status_code.toString().includes(q);
            if (!match) return false;
        }
        if (filterIP && !log.source_ip.includes(filterIP)) return false;
        if (filterCode && !log.status_code.toString().includes(filterCode))
            return false;
        if (
            filterFile &&
            !log.requested_file.toLowerCase().includes(filterFile.toLowerCase())
        )
            return false;
        return true;
    });
</script>

<div class="h-full flex flex-col p-6 space-y-4 overflow-hidden">
    <!-- Header -->
    <div class="flex justify-between items-center shrink-0">
        <h1 class="text-3xl font-bold tracking-tight">
            {$t("menu.debug") || "Debug Logs"}
        </h1>
        <div class="flex gap-2">
            <Button variant="outline" on:click={togglePause}>
                {#if isPaused}
                    <Play class="mr-2 h-4 w-4" /> Resume
                {:else}
                    <Pause class="mr-2 h-4 w-4" /> Pause
                {/if}
            </Button>
            <Button variant="destructive" on:click={clearLogs}>
                <Trash2 class="mr-2 h-4 w-4" /> Clear
            </Button>
        </div>
    </div>

    <!-- Controls & Filters -->
    <Card.Root class="shrink-0">
        <Card.Content class="p-4 space-y-4">
            <div class="flex gap-4">
                <div class="relative flex-1">
                    <Search
                        class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground"
                    />
                    <Input
                        placeholder="Search all..."
                        class="pl-8"
                        bind:value={searchQuery}
                    />
                </div>
                <Button
                    variant="outline"
                    on:click={toggleFilters}
                    class="relative"
                >
                    <Filter class="mr-2 h-4 w-4" />
                    Filters
                    {#if hasActiveFilters}
                        <span class="absolute -top-1 -right-1 flex h-3 w-3">
                            <span
                                class="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"
                            ></span>
                            <span
                                class="relative inline-flex rounded-full h-3 w-3 bg-primary"
                            ></span>
                        </span>
                    {/if}
                    {#if isFiltersOpen}
                        <ChevronUp class="ml-2 h-4 w-4" />
                    {:else}
                        <ChevronDown class="ml-2 h-4 w-4" />
                    {/if}
                </Button>
            </div>

            {#if isFiltersOpen}
                <div
                    class="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2 border-t"
                >
                    <div class="space-y-1">
                        <span class="text-xs font-medium text-muted-foreground"
                            >Source IP</span
                        >
                        <Input
                            placeholder="Filter by IP"
                            bind:value={filterIP}
                            class="h-8"
                        />
                    </div>
                    <div class="space-y-1">
                        <span class="text-xs font-medium text-muted-foreground"
                            >Status Code</span
                        >
                        <Input
                            placeholder="Filter by Code"
                            bind:value={filterCode}
                            class="h-8"
                        />
                    </div>
                    <div class="space-y-1">
                        <span class="text-xs font-medium text-muted-foreground"
                            >Requested File</span
                        >
                        <Input
                            placeholder="Filter by File"
                            bind:value={filterFile}
                            class="h-8"
                        />
                    </div>
                </div>
            {/if}
        </Card.Content>
    </Card.Root>

    <!-- Logs Table -->
    <div
        class="flex-1 border rounded-md overflow-hidden flex flex-col bg-background"
    >
        <div class="flex-1 overflow-y-auto" bind:this={tableContainer}>
            <Table.Root>
                <Table.Header class="sticky top-0 bg-background z-10 shadow-sm">
                    <Table.Row>
                        <Table.Head class="w-[180px]">Time</Table.Head>
                        <Table.Head class="w-[150px]">Source IP</Table.Head>
                        <Table.Head class="w-[100px]">Status</Table.Head>
                        <Table.Head class="w-[100px]">Method</Table.Head>
                        <Table.Head>Requested File</Table.Head>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {#each filteredLogs as log (log)}
                        <Table.Row>
                            <Table.Cell
                                class="font-mono text-xs whitespace-nowrap"
                            >
                                {new Date(log.time).toLocaleString()}
                            </Table.Cell>
                            <Table.Cell class="font-mono"
                                >{log.source_ip}</Table.Cell
                            >
                            <Table.Cell>
                                <span
                                    class:text-green-600={log.status_code >=
                                        200 && log.status_code < 300}
                                    class:text-yellow-600={log.status_code >=
                                        300 && log.status_code < 400}
                                    class:text-red-600={log.status_code >= 400}
                                    class="font-bold"
                                >
                                    {log.status_code}
                                </span>
                            </Table.Cell>
                            <Table.Cell>{log.method}</Table.Cell>
                            <Table.Cell class="font-mono text-sm break-all"
                                >{log.requested_file}</Table.Cell
                            >
                        </Table.Row>
                    {/each}
                    {#if filteredLogs.length === 0}
                        <Table.Row>
                            <Table.Cell
                                colspan="5"
                                class="text-center text-muted-foreground py-8"
                            >
                                No logs received yet...
                            </Table.Cell>
                        </Table.Row>
                    {/if}
                </Table.Body>
            </Table.Root>
        </div>
    </div>
</div>
