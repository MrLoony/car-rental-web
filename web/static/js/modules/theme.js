import { on, qsa } from "./utils.js";

const storageKey = "carRentalTheme";
const validModes = ["system", "light", "dark"];
const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
let currentMode = "system";

export function initTheme() {
    currentMode = getThemePreference();
    applyTheme(currentMode);
    updateThemeControls();

    qsa("[data-theme-toggle]").forEach((button) => {
        on(button, "click", () => {
            const nextMode = cycleTheme();
            announceThemeChange(nextMode);
        });
    });

    on(mediaQuery, "change", () => {
        if (currentMode === "system") {
            applyTheme("system");
            updateThemeControls();
        }
    });
}

export function getThemePreference() {
    try {
        const storedMode = window.localStorage.getItem(storageKey);
        return validModes.includes(storedMode) ? storedMode : "system";
    } catch {
        return "system";
    }
}

export function setThemePreference(mode) {
    const nextMode = validModes.includes(mode) ? mode : "system";
    currentMode = nextMode;

    try {
        window.localStorage.setItem(storageKey, nextMode);
    } catch {
        // Theme still applies for this page even if storage is unavailable.
    }

    applyTheme(nextMode);
    updateThemeControls();
    return nextMode;
}

export function getResolvedTheme() {
    if (currentMode === "dark") {
        return "dark";
    }
    if (currentMode === "light") {
        return "light";
    }

    return mediaQuery.matches ? "dark" : "light";
}

export function applyTheme(mode = currentMode) {
    currentMode = validModes.includes(mode) ? mode : "system";
    const resolvedTheme = getResolvedTheme();
    const root = document.documentElement;

    root.classList.toggle("dark", resolvedTheme === "dark");
    root.dataset.themeMode = currentMode;
    root.dataset.themeResolved = resolvedTheme;

    return resolvedTheme;
}

export function cycleTheme() {
    const index = validModes.indexOf(currentMode);
    const nextMode = validModes[(index + 1) % validModes.length] || "system";
    return setThemePreference(nextMode);
}

function updateThemeControls() {
    const label = themeLabel(currentMode);
    const resolvedTheme = getResolvedTheme();

    qsa("[data-theme-toggle]").forEach((button) => {
        button.dataset.themeMode = currentMode;
        button.dataset.themeResolved = resolvedTheme;
        button.setAttribute("aria-label", `Color theme: ${label}. Activate to change theme.`);
        button.title = `Color theme: ${label}`;
    });

    qsa("[data-theme-label]").forEach((target) => {
        target.textContent = label;
    });
}

function announceThemeChange(mode) {
    document.dispatchEvent(new CustomEvent("app:toast", {
        detail: {
            type: "info",
            message: `Theme set to ${themeLabel(mode)}.`,
        },
    }));
}

function themeLabel(mode) {
    if (mode === "light") {
        return "Light";
    }
    if (mode === "dark") {
        return "Dark";
    }

    return "System";
}
