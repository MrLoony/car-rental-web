function debounce(fn, delay) {
    let timeoutID;

    return function debounced(...args) {
        window.clearTimeout(timeoutID);
        timeoutID = window.setTimeout(() => fn.apply(this, args), delay);
    };
}

function initCatalogFilters() {
    const form = document.querySelector("[data-filter-form]");
    if (!form) {
        return;
    }

    const searchInput = form.querySelector("[data-filter-search]");
    const filterSelects = form.querySelectorAll("[data-filter-select]");
    const submitButton = form.querySelector("[data-filter-submit]");
    const status = form.querySelector("[data-filter-status]");
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
        form.requestSubmit();
    }

    const debouncedSubmit = debounce(submitIfChanged, 500);

    filterSelects.forEach((select) => {
        select.addEventListener("change", submitIfChanged);
    });

    if (searchInput) {
        searchInput.addEventListener("input", debouncedSubmit);
    }

    form.addEventListener("submit", showStatus);
}

document.addEventListener("DOMContentLoaded", initCatalogFilters);
