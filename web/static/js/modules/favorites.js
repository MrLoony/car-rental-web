import { on, qs, qsa } from "./utils.js";

const storageKey = "carRentalFavorites";
const favoritesFilterKey = "carRentalFavoritesOnly";
let favorites = new Set();

export function initFavorites() {
    favorites = new Set(readFavoritesFromStorage());
    initFavoriteButtons();
    initFavoritesFilter();
    renderFavoriteState();

    on(window, "storage", (event) => {
        if (event.key !== storageKey && event.key !== favoritesFilterKey) {
            return;
        }

        favorites = new Set(readFavoritesFromStorage());
        renderFavoriteState();
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
        });
    });
}

function initFavoritesFilter() {
    qsa("[data-favorites-only-toggle]").forEach((toggle) => {
        toggle.checked = readFavoritesOnlyPreference();

        on(toggle, "change", () => {
            persistFavoritesOnlyPreference(toggle.checked);
            renderFavoriteState();
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
    const favoritesOnly = readFavoritesOnlyPreference();
    const cards = qsa("[data-favorite-card]");
    const emptyState = qs("[data-favorites-empty-state]");
    let visibleCount = 0;

    qsa("[data-favorites-only-toggle]").forEach((toggle) => {
        toggle.checked = favoritesOnly;
    });

    cards.forEach((card) => {
        const visible = !favoritesOnly || isFavorite(card.dataset.carSlug);
        card.hidden = !visible;
        if (visible) {
            visibleCount += 1;
        }
    });

    if (emptyState) {
        emptyState.classList.toggle("hidden", !favoritesOnly || visibleCount > 0);
    }
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

function readFavoritesOnlyPreference() {
    try {
        return window.localStorage.getItem(favoritesFilterKey) === "true";
    } catch {
        return false;
    }
}

function persistFavoritesOnlyPreference(enabled) {
    try {
        window.localStorage.setItem(favoritesFilterKey, enabled ? "true" : "false");
    } catch {
        // Filtering still applies for the current page even if storage is unavailable.
    }
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
