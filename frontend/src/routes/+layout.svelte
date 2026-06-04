<script lang="ts">
  import "../lib/i18n";
  import "../app.css";
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import Sidebar from "$lib/components/Sidebar.svelte";
  import { page } from "$app/stores"; // Keep this import as it's used in the markup
  import Header from "$lib/components/Header.svelte"; // Keep this import as it's used in the markup
  import { theme } from "$lib/stores/theme";
  import { waitLocale, t } from "svelte-i18n";
  import { Toaster } from "svelte-sonner";

  import { onDestroy } from "svelte";

  let isLoading = true;
  let authCheckInterval: any;
  let remainingSeconds = 0;
  let showWarning = false;
  let uiTickInterval: any;

  function startUiTick() {
    if (uiTickInterval) return;
    uiTickInterval = setInterval(() => {
      if (remainingSeconds > 0) {
        remainingSeconds -= 1;
      } else {
        stopUiTick();
      }
    }, 1000);
  }

  function stopUiTick() {
    if (uiTickInterval) {
      clearInterval(uiTickInterval);
      uiTickInterval = null;
    }
  }

  async function performAuthCheck() {
    try {
      const res = await fetch("/api/check-auth");
      if (res.ok) {
        const data = await res.json();
        const rem = data.remaining_seconds;
        if (rem > 0 && rem < 30) {
          remainingSeconds = rem;
          showWarning = true;
          startUiTick();
        } else {
          showWarning = false;
          stopUiTick();
        }
      }
    } catch (e) {
      // Ignore network errors
    }
  }

  function startAuthCheck() {
    if (authCheckInterval) return;
    performAuthCheck();
    authCheckInterval = setInterval(performAuthCheck, 10000); // Check every 10 seconds
  }

  function stopAuthCheck() {
    if (authCheckInterval) {
      clearInterval(authCheckInterval);
      authCheckInterval = null;
    }
    showWarning = false;
    stopUiTick();
  }

  async function resetSessionTimer() {
    try {
      const res = await fetch("/api/check-auth", { method: "POST" });
      if (res.ok) {
        const data = await res.json();
        remainingSeconds = data.remaining_seconds;
        showWarning = false;
        stopUiTick();
      }
    } catch (e) {
      // Ignore errors
    }
  }

  $: {
    if (typeof window !== "undefined") {
      if ($page.url.pathname === "/login") {
        stopAuthCheck();
      } else {
        startAuthCheck();
      }
    }
  }

  onDestroy(() => {
    stopAuthCheck();
  });

  onMount(async () => {
    theme.init();
    await waitLocale();
    isLoading = false;

    // Set up global fetch interceptor once
    if (typeof window !== "undefined" && !(window as any).__fetchIntercepted__) {
      (window as any).__fetchIntercepted__ = true;
      const originalFetch = window.fetch;
      window.fetch = async (...args) => {
        try {
          const response = await originalFetch(...args);
          if (response.status === 401 && $page.url.pathname !== "/login") {
            goto("/login");
          }
          return response;
        } catch (error) {
          throw error;
        }
      };
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

    {#if showWarning}
      <div class="fixed bottom-4 right-4 z-50 max-w-sm w-full bg-white/95 dark:bg-gray-900/95 backdrop-blur border border-amber-300 dark:border-amber-700 rounded-lg shadow-xl p-4 transition-all duration-300 ease-in-out">
        <div class="flex items-start">
          <div class="flex-shrink-0">
            <!-- Warning Icon -->
            <svg class="h-6 w-6 text-amber-500 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div class="ml-3 w-0 flex-1 pt-0.5">
            <p class="text-sm font-semibold text-gray-900 dark:text-gray-100">
              {$t("auth.session_timeout_warning_title", {default: "Сессия скоро завершится"})}
            </p>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {$t("auth.session_timeout_warning_desc", {
                values: { seconds: remainingSeconds },
                default: "Вы будете автоматически выведены из системы через {seconds} сек. из-за неактивности."
              })}
            </p>
            <div class="mt-3 flex space-x-3">
              <button
                type="button"
                class="inline-flex items-center px-3 py-2 border border-transparent text-xs font-semibold rounded-md shadow-sm text-white bg-amber-500 hover:bg-amber-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 transition-colors"
                on:click={resetSessionTimer}
              >
                {$t("auth.session_extend", {default: "Продолжить работу"})}
              </button>
            </div>
          </div>
        </div>
      </div>
    {/if}
  {/if}
{/if}
