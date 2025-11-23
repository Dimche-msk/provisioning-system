import { writable } from "svelte/store";
import { browser } from "$app/environment";

type Theme = "light" | "dark";

function createThemeStore() {
    const { subscribe, set, update } = writable<Theme>("light");

    return {
        subscribe,
        init: () => {
            if (!browser) return;

            const savedTheme = localStorage.getItem("theme") as Theme;
            if (savedTheme) {
                set(savedTheme);
                updateClass(savedTheme);
            } else if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
                set("dark");
                updateClass("dark");
            } else {
                set("light");
                updateClass("light");
            }
        },
        toggle: () => {
            update((current) => {
                const newTheme = current === "light" ? "dark" : "light";
                localStorage.setItem("theme", newTheme);
                updateClass(newTheme);
                return newTheme;
            });
        },
        set: (theme: Theme) => {
            localStorage.setItem("theme", theme);
            updateClass(theme);
            set(theme);
        }
    };
}

function updateClass(theme: Theme) {
    if (!browser) return;
    const root = document.documentElement;
    if (theme === "dark") {
        root.classList.add("dark");
    } else {
        root.classList.remove("dark");
    }
}

export const theme = createThemeStore();
