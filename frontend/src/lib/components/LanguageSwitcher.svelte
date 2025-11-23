<script>
    import { locale } from "svelte-i18n";
    import { Button } from "$lib/components/ui/button";

    function toggleLang() {
        const newLang = $locale === "ru" ? "en" : "ru";
        locale.set(newLang);
        localStorage.setItem("locale", newLang);
    }
</script>

<Button variant="ghost" on:click={toggleLang}>
    {$locale === "ru" ? "EN" : "RU"}
</Button>
