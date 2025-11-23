<script>
    import { t } from "svelte-i18n";
    import * as Card from "$lib/components/ui/card";
    import PhoneForm from "$lib/components/phones/PhoneForm.svelte";
</script>

<div class="p-6 space-y-6">
    <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
        {$t("menu.home")}
    </h1>

    <div class="grid gap-6 md:grid-cols-2">
        <PhoneForm mode="create" />

        <Card.Root>
            <Card.Header>
                <Card.Title>{$t("welcome")}</Card.Title>
                <Card.Description>{$t("welcome_message")}</Card.Description>
            </Card.Header>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    Тут будет онформация о количестве телефонов по локациям,
                    вендорам и моделям
                </p>
            </Card.Content>
        </Card.Root>
    </div>
</div>
