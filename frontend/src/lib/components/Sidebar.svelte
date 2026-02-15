<script>
    import { page } from "$app/stores";
    import { t } from "svelte-i18n";
    import { Settings, Home, Phone, Activity } from "lucide-svelte";

    import { APP_VERSION } from "$lib/version";

    const menuItems = [
        { href: "/", label: "menu.home", icon: Home },
        { href: "/phones", label: "menu.phones", icon: Phone },
        { href: "/system", label: "menu.system", icon: Settings },
        { href: "/debug", label: "menu.debug", icon: Activity },
    ];
</script>

<aside
    class="w-64 bg-gray-100 dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 hidden md:flex flex-col"
>
    <div class="p-6">
        <h1 class="text-xl font-bold text-gray-900 dark:text-gray-100">
            Provisioning
        </h1>
    </div>
    <nav class="flex-1 px-4 space-y-2">
        {#each menuItems as item}
            <a
                href={item.href}
                class="flex items-center gap-3 px-4 py-2 rounded-md transition-colors {$page
                    .url.pathname === item.href
                    ? 'bg-primary text-primary-foreground'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-800'}"
            >
                <svelte:component this={item.icon} size={20} />
                <span>{$t(item.label)}</span>
            </a>
        {/each}
    </nav>
    <div class="p-4 border-t border-gray-200 dark:border-gray-800">
        <p class="text-xs text-center text-gray-500">v{APP_VERSION}</p>
    </div>
</aside>
