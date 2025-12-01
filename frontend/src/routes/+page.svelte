<script>
    import { t } from "svelte-i18n";
    import * as Card from "$lib/components/ui/card";
    import * as Table from "$lib/components/ui/table";
    import PhoneForm from "$lib/components/phones/PhoneForm.svelte";
    import { onMount } from "svelte";
    import { Activity, Server, Smartphone } from "lucide-svelte";

    let stats = [];
    let totalPhones = 0;
    let vendors = [];
    let domains = [];
    let pivotData = {}; // domain -> vendor -> count
    let domainTotals = {}; // domain -> total

    onMount(async () => {
        try {
            const res = await fetch("/api/system/stats");
            if (res.ok) {
                const data = await res.json();
                stats = data.stats || [];
                totalPhones = data.total_phones;
                processStats();
            }
        } catch (e) {
            console.error("Failed to load stats", e);
        }
    });

    function processStats() {
        const vSet = new Set();
        const dSet = new Set();
        pivotData = {};
        domainTotals = {};

        stats.forEach((s) => {
            vSet.add(s.vendor);
            dSet.add(s.domain);

            if (!pivotData[s.domain]) pivotData[s.domain] = {};
            pivotData[s.domain][s.vendor] = s.count;

            domainTotals[s.domain] = (domainTotals[s.domain] || 0) + s.count;
        });

        vendors = Array.from(vSet).sort();
        domains = Array.from(dSet).sort();
    }
</script>

<div class="p-6 space-y-6">
    <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
        {$t("menu.home")}
    </h1>

    <div class="grid gap-6 md:grid-cols-2">
        <PhoneForm mode="create" />

        <div class="space-y-6">
            <!-- Summary Cards -->
            <div class="grid grid-cols-2 gap-4">
                <Card.Root>
                    <Card.Header
                        class="flex flex-row items-center justify-between space-y-0 pb-2"
                    >
                        <Card.Title class="text-sm font-medium"
                            >Total Phones</Card.Title
                        >
                        <Smartphone class="h-4 w-4 text-muted-foreground" />
                    </Card.Header>
                    <Card.Content>
                        <div class="text-2xl font-bold">{totalPhones}</div>
                    </Card.Content>
                </Card.Root>
                <Card.Root>
                    <Card.Header
                        class="flex flex-row items-center justify-between space-y-0 pb-2"
                    >
                        <Card.Title class="text-sm font-medium"
                            >Active Domains</Card.Title
                        >
                        <Server class="h-4 w-4 text-muted-foreground" />
                    </Card.Header>
                    <Card.Content>
                        <div class="text-2xl font-bold">{domains.length}</div>
                    </Card.Content>
                </Card.Root>
            </div>

            <!-- Stats Table -->
            <Card.Root>
                <Card.Header>
                    <Card.Title>{$t("system.overview")}</Card.Title>
                    <Card.Description
                        >{$t("system.overview_description")}</Card.Description
                    >
                </Card.Header>
                <Card.Content>
                    {#if domains.length > 0}
                        <Table.Root>
                            <Table.Header>
                                <Table.Row>
                                    <Table.Head>{$t("phone.domain")}</Table.Head>
                                    {#each vendors as vendor}
                                        <Table.Head class="capitalize"
                                            >{vendor}</Table.Head
                                        >
                                    {/each}
                                    <Table.Head class="text-right font-bold"
                                        >{$t("common.total")}</Table.Head
                                    >
                                </Table.Row>
                            </Table.Header>
                            <Table.Body>
                                {#each domains as domain}
                                    <Table.Row>
                                        <Table.Cell class="font-medium"
                                            >{domain}</Table.Cell
                                        >
                                        {#each vendors as vendor}
                                            <Table.Cell>
                                                {(pivotData[domain] &&
                                                    pivotData[domain][
                                                        vendor
                                                    ]) ||
                                                    0}
                                            </Table.Cell>
                                        {/each}
                                        <Table.Cell
                                            class="text-right font-bold"
                                        >
                                            {domainTotals[domain] || 0}
                                        </Table.Cell>
                                    </Table.Row>
                                {/each}
                            </Table.Body>
                        </Table.Root>
                    {:else}
                        <div class="text-center text-muted-foreground py-8">
                            No data available
                        </div>
                    {/if}
                </Card.Content>
            </Card.Root>
        </div>
    </div>
</div>
