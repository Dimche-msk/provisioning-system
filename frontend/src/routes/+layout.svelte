<script>
  import "../lib/i18n";
  import "../app.css";
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import Sidebar from "$lib/components/Sidebar.svelte";
  import { page } from "$app/stores"; // Keep this import as it's used in the markup
  import Header from "$lib/components/Header.svelte"; // Keep this import as it's used in the markup
  import { theme } from "$lib/stores/theme";
  import { waitLocale } from "svelte-i18n";
  import { Toaster } from "svelte-sonner";

  let isLoading = true;

  onMount(async () => {
    theme.init();
    await waitLocale();
    isLoading = false;
    // Не проверяем авторизацию, если мы уже на странице логина
    if ($page.url.pathname === "/login") {
      return;
    }

    try {

      const res = await fetch("/api/check-auth");
      if (!res.ok) {
        goto("/login");
      }
    } catch (e) {
      goto("/login");
    }
  });
</script>

<Toaster position="top-center" richColors closeButton />

{#if !isLoading}
  {#if $page.url.pathname === "/login"}
    <slot />
  {:else}
    <div class="flex h-screen overflow-hidden bg-gray-50 dark:bg-gray-950">
      <Sidebar />
      <div class="flex-1 flex flex-col overflow-hidden">
        <Header />
        <main class="flex-1 overflow-y-auto">
          <slot />
        </main>
      </div>
    </div>
  {/if}
{/if}
