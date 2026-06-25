import { debounce, on, qs, qsa, submitForm } from "./utils.js";

export function initCatalogFilters() {
    initPublicCatalogFilters();
    initAdminFilters();
}

function initPublicCatalogFilters() {
    const form = qs("[data-catalog-filter-form]") || qs("[data-filter-form]");
    if (!form) {
        return;
    }

    const searchInput = qs("[data-debounce-search]", form) || qs("[data-filter-search]", form);
    const filterSelects = qsa("[data-auto-submit]", form).length ? qsa("[data-auto-submit]", form) : qsa("[data-filter-select]", form);
    const submitButton = qs("[data-filter-submit]", form);
    const status = qs("[data-filter-status]", form);
    const resetLinks = qsa("[data-catalog-reset]");
    let lastSubmittedState = new URLSearchParams(new FormData(form)).toString();

    function showStatus() {
        form.classList.add("catalog-loading");
        form.setAttribute("aria-busy", "true");

        if (status) {
            status.classList.remove("hidden");
        }

        if (submitButton) {
            submitButton.disabled = true;
        }
    }

    function submitIfChanged() {
        const currentState = new URLSearchParams(new FormData(form)).toString();
        if (currentState === lastSubmittedState) {
            return;
        }

        lastSubmittedState = currentState;
        showStatus();
        submitForm(form);
    }

    const debouncedSubmit = debounce(submitIfChanged, 500);

    filterSelects.forEach((select) => {
        select.addEventListener("change", () => {
            updateActiveFilterCount(form);
            submitIfChanged();
        });
    });

    if (searchInput) {
        searchInput.addEventListener("input", () => {
            updateActiveFilterCount(form);
            debouncedSubmit();
        });
    }

    initCatalogFilterToggle();
    updateActiveFilterCount(form);
    resetLinks.forEach((link) => {
        on(link, "click", (event) => {
            event.preventDefault();
            window.location.assign("/cars");
        });
    });
    form.addEventListener("submit", showStatus);
}

function initCatalogFilterToggle() {
    const toggle = qs("[data-catalog-filter-toggle]");
    const panel = qs("[data-catalog-filter-panel]");
    if (!toggle || !panel) {
        return;
    }

    const mobileMedia = window.matchMedia("(max-width: 767px)");

    function applyMobileState() {
        if (mobileMedia.matches) {
            toggle.classList.remove("hidden");
            panel.hidden = true;
            toggle.setAttribute("aria-expanded", "false");
            return;
        }

        toggle.classList.add("hidden");
        panel.hidden = false;
        toggle.setAttribute("aria-expanded", "true");
    }

    on(toggle, "click", () => {
        const isExpanded = toggle.getAttribute("aria-expanded") === "true";
        panel.hidden = isExpanded;
        toggle.setAttribute("aria-expanded", isExpanded ? "false" : "true");
    });

    applyMobileState();
    if (typeof mobileMedia.addEventListener === "function") {
        mobileMedia.addEventListener("change", applyMobileState);
    }
}

function updateActiveFilterCount(form) {
    const output = qs("[data-catalog-active-filter-count]");
    if (!output) {
        return;
    }

    const formData = new FormData(form);
    let count = 0;

    if ((formData.get("search") || "").toString().trim() !== "") {
        count += 1;
    }
    if ((formData.get("category") || "").toString() !== "") {
        count += 1;
    }
    if ((formData.get("fuel") || "").toString() !== "") {
        count += 1;
    }
    if ((formData.get("transmission") || "").toString() !== "") {
        count += 1;
    }
    if ((formData.get("sort") || "").toString() !== "" && formData.get("sort") !== "newest") {
        count += 1;
    }
    if (new URLSearchParams(window.location.search).has("favorites")) {
        count += 1;
    }

    output.textContent = String(count);
}

function initAdminFilters() {
    const forms = qsa("[data-admin-filter-form]");
    if (!forms.length) {
        return;
    }

    forms.forEach((form) => {
        const searchInput = qs("[data-admin-filter-search]", form);
        const filterSelects = qsa("[data-admin-filter-select]", form);
        const submitButton = qs("[data-admin-filter-submit]", form);
        const status = qs("[data-admin-filter-status]", form);
        let lastSubmittedState = new URLSearchParams(new FormData(form)).toString();

        function showStatus() {
            if (status) {
                status.classList.remove("hidden");
            }

            if (submitButton) {
                submitButton.disabled = true;
            }
        }

        function submitIfChanged() {
            const currentState = new URLSearchParams(new FormData(form)).toString();
            if (currentState === lastSubmittedState) {
                return;
            }

            lastSubmittedState = currentState;
            showStatus();
            submitForm(form);
        }

        const debouncedSubmit = debounce(submitIfChanged, 500);

        filterSelects.forEach((select) => {
            select.addEventListener("change", submitIfChanged);
        });

        if (searchInput) {
            searchInput.addEventListener("input", debouncedSubmit);
        }

        form.addEventListener("submit", showStatus);
    });
}
