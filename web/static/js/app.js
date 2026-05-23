(() => {
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

    function initBookingPreview() {
        const form = document.querySelector("[data-booking-form]");
        if (!form) {
            return;
        }

        const pickupInput = form.querySelector("[data-pickup-input]");
        const returnInput = form.querySelector("[data-return-input]");
        const billingDaysOutput = form.querySelector("[data-billing-days]");
        const estimatedTotalOutput = form.querySelector("[data-estimated-total]");
        const warning = form.querySelector("[data-booking-warning]");
        const pricePerDay = Number.parseFloat(form.dataset.pricePerDay);

        if (!pickupInput || !returnInput || !billingDaysOutput || !estimatedTotalOutput || !warning || Number.isNaN(pricePerDay)) {
            return;
        }

        function setNeutralPreview() {
            billingDaysOutput.textContent = "--";
            estimatedTotalOutput.textContent = "--";
            warning.textContent = "Select pickup and return times to see an estimate.";
            warning.classList.remove("text-red-600");
            warning.classList.add("text-slate-600");
        }

        function setWarning(message) {
            billingDaysOutput.textContent = "--";
            estimatedTotalOutput.textContent = "--";
            warning.textContent = message;
            warning.classList.remove("text-slate-600");
            warning.classList.add("text-red-600");
        }

        function updatePreview() {
            if (!pickupInput.value || !returnInput.value) {
                setNeutralPreview();
                return;
            }

            const pickupDate = new Date(pickupInput.value);
            const returnDate = new Date(returnInput.value);

            if (Number.isNaN(pickupDate.getTime()) || Number.isNaN(returnDate.getTime())) {
                setWarning("Enter valid pickup and return times.");
                return;
            }

            if (returnDate <= pickupDate) {
                setWarning("Return time must be after pickup time.");
                return;
            }

            const billingDays = calculateBillingDays(pickupDate, returnDate);
            const estimatedTotal = billingDays * pricePerDay;

            billingDaysOutput.textContent = billingDays === 1 ? "1 day" : `${billingDays} days`;
            estimatedTotalOutput.textContent = formatCurrency(estimatedTotal);
            warning.textContent = "Estimate updates as you adjust pickup and return times.";
            warning.classList.remove("text-red-600");
            warning.classList.add("text-slate-600");
        }

        pickupInput.addEventListener("input", updatePreview);
        pickupInput.addEventListener("change", updatePreview);
        returnInput.addEventListener("input", updatePreview);
        returnInput.addEventListener("change", updatePreview);
        updatePreview();
    }

    function calculateBillingDays(pickupDate, returnDate) {
        const durationMs = returnDate - pickupDate;
        const durationHours = durationMs / 1000 / 60 / 60;
        const billingDays = Math.ceil(durationHours / 24);

        return Math.max(1, billingDays);
    }

    function formatCurrency(amount) {
        return `$${amount.toFixed(2)}`;
    }

    function initAdminStatusConfirm() {
        const form = document.querySelector("[data-admin-status-form]");
        if (!form) {
            return;
        }

        form.addEventListener("submit", (event) => {
            if (!window.confirm("Update booking status?")) {
                event.preventDefault();
            }
        });
    }

    document.addEventListener("DOMContentLoaded", () => {
        initCatalogFilters();
        initBookingPreview();
        initAdminStatusConfirm();
    });
})();
