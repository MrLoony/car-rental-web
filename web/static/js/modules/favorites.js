import { on, qs, qsa } from "./utils.js";

const storageKey = "carRentalFavorites";
let favorites = new Set();

export function initFavorites() {
    favorites = new Set(readFavoritesFromStorage());
    initFavoriteButtons();
    initFavoritesFilter();
    renderFavoriteState();

    on(window, "storage", (event) => {
        if (event.key !== storageKey) {
            return;
        }

        favorites = new Set(readFavoritesFromStorage());
        renderFavoriteState();
        if (isFavoritesModeActive()) {
            syncFavoritesURL();
        }
    });
}

export function addFavorite(carSlug) {
    const slug = normalizeSlug(carSlug);
    if (!slug) {
        return getFavorites();
    }

    favorites.add(slug);
    persistFavorites();
    renderFavoriteState();
    return getFavorites();
}

export function removeFavorite(carSlug) {
    const slug = normalizeSlug(carSlug);
    if (!slug) {
        return getFavorites();
    }

    favorites.delete(slug);
    persistFavorites();
    renderFavoriteState();
    return getFavorites();
}

export function toggleFavorite(carSlug) {
    const slug = normalizeSlug(carSlug);
    if (!slug) {
        return false;
    }

    const nextIsFavorite = !favorites.has(slug);
    if (nextIsFavorite) {
        favorites.add(slug);
    } else {
        favorites.delete(slug);
    }

    persistFavorites();
    renderFavoriteState();
    return nextIsFavorite;
}

export function isFavorite(carSlug) {
    return favorites.has(normalizeSlug(carSlug));
}

export function getFavorites() {
    return Array.from(favorites).sort();
}

export function getFavoritesCount() {
    return favorites.size;
}

export function renderFavoriteState() {
    renderFavoriteButtons();
    renderFavoritesCounter();
    renderFavoritesFilter();
}

function initFavoriteButtons() {
    qsa("[data-favorite-toggle]").forEach((button) => {
        on(button, "click", () => {
            const slug = button.dataset.carSlug;
            const carName = button.dataset.carName || "vehicle";
            const isNowFavorite = toggleFavorite(slug);

            showFavoritesToast(isNowFavorite, carName);
            if (isFavoritesModeActive()) {
                syncFavoritesURL();
            }
        });
    });
}

function initFavoritesFilter() {
    qsa("[data-favorites-only-toggle]").forEach((toggle) => {
        toggle.checked = isFavoritesModeActive();

        on(toggle, "change", () => {
            if (toggle.checked) {
                navigateToFavoritesMode();
            } else {
                navigateFromFavoritesMode();
            }
        });
    });
}

function renderFavoriteButtons() {
    qsa("[data-favorite-toggle]").forEach((button) => {
        const slug = button.dataset.carSlug;
        const carName = button.dataset.carName || "vehicle";
        const active = isFavorite(slug);
        const icon = qs("[data-favorite-icon]", button);
        const label = qs("[data-favorite-label]", button);

        button.classList.toggle("favorite-button-active", active);
        button.setAttribute("aria-pressed", active ? "true" : "false");
        button.setAttribute("aria-label", `${active ? "Remove" : "Add"} ${carName} ${active ? "from" : "to"} favorites`);

        if (icon) {
            icon.textContent = active ? "♥" : "♡";
        }
        if (label) {
            label.textContent = active ? "Saved" : "Save";
        }
    });
}

function renderFavoritesCounter() {
    const count = getFavoritesCount();

    qsa("[data-favorites-counter]").forEach((counter) => {
        counter.hidden = false;
        counter.classList.toggle("favorites-counter-active", count > 0);
        counter.setAttribute("aria-label", `${count} saved ${count === 1 ? "vehicle" : "vehicles"}`);
    });

    qsa("[data-favorites-count]").forEach((target) => {
        target.textContent = String(count);
    });
}

function renderFavoritesFilter() {
    const favoritesOnly = isFavoritesModeActive();

    qsa("[data-favorites-only-toggle]").forEach((toggle) => {
        toggle.checked = favoritesOnly;
    });
}

function readFavoritesFromStorage() {
    try {
        const parsed = JSON.parse(window.localStorage.getItem(storageKey) || "[]");
        if (!Array.isArray(parsed)) {
            return [];
        }

        return Array.from(new Set(parsed.map(normalizeSlug).filter(Boolean)));
    } catch {
        return [];
    }
}

function persistFavorites() {
    try {
        window.localStorage.setItem(storageKey, JSON.stringify(getFavorites()));
    } catch {
        // Keep in-memory state for the current page if storage is unavailable.
    }
}

function isFavoritesModeActive() {
    return new URLSearchParams(window.location.search).has("favorites");
}

function navigateToFavoritesMode() {
    const url = new URL(window.location.href);
    const slugs = getFavorites();

    url.searchParams.delete("page");
    url.searchParams.set("favorites", slugs.join(","));
    window.location.assign(url.toString());
}

function navigateFromFavoritesMode() {
    const url = new URL(window.location.href);

    url.searchParams.delete("favorites");
    url.searchParams.delete("page");
    window.location.assign(url.toString());
}

function syncFavoritesURL() {
    const current = new URLSearchParams(window.location.search).get("favorites") || "";
    const next = getFavorites().join(",");
    if (current === next) {
        return;
    }

    const url = new URL(window.location.href);
    url.searchParams.delete("page");
    url.searchParams.set("favorites", next);
    window.location.assign(url.toString());
}

function normalizeSlug(value) {
    return String(value || "").trim().toLowerCase();
}

function showFavoritesToast(added, carName) {
    document.dispatchEvent(new CustomEvent("app:toast", {
        detail: {
            type: added ? "success" : "info",
            message: `${added ? "Added" : "Removed"} ${carName} ${added ? "to" : "from"} favorites.`,
        },
    }));
}
