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

    function initAdminFilters() {
        const forms = document.querySelectorAll("[data-admin-filter-form]");
        if (!forms.length) {
            return;
        }

        forms.forEach((form) => {
            const searchInput = form.querySelector("[data-admin-filter-search]");
            const filterSelects = form.querySelectorAll("[data-admin-filter-select]");
            const submitButton = form.querySelector("[data-admin-filter-submit]");
            const status = form.querySelector("[data-admin-filter-status]");
            let lastSubmittedState = new URLSearchParams(new FormData(form)).toString();

            function showStatus() {
                if (status) {
                    status.classList.remove("hidden");
                }

                if (submitButton) {
                    submitButton.disabled = true;
                }
            }

            function submitForm() {
                showStatus();

                if (typeof form.requestSubmit === "function") {
                    form.requestSubmit();
                    return;
                }

                form.submit();
            }

            function submitIfChanged() {
                const currentState = new URLSearchParams(new FormData(form)).toString();
                if (currentState === lastSubmittedState) {
                    return;
                }

                lastSubmittedState = currentState;
                submitForm();
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

    function initImagePreview() {
        const imageInputs = document.querySelectorAll("[data-image-url-input]");
        if (!imageInputs.length) {
            return;
        }

        imageInputs.forEach((input) => {
            const form = input.closest("form");
            const preview = form ? form.querySelector("[data-image-preview]") : null;
            const previewImage = preview ? preview.querySelector("[data-image-preview-img]") : null;
            const previewMessage = preview ? preview.querySelector("[data-image-preview-message]") : null;

            if (!preview || !previewImage || !previewMessage) {
                return;
            }

            function showMessage(message, isError = false) {
                previewImage.classList.add("hidden");
                previewMessage.textContent = message;
                previewMessage.classList.remove("hidden", "text-red-600", "text-slate-500");
                previewMessage.classList.add("flex", isError ? "text-red-600" : "text-slate-500");
            }

            function showImage(src) {
                previewImage.onload = () => {
                    previewImage.classList.remove("hidden");
                    previewMessage.classList.add("hidden");
                    previewMessage.classList.remove("flex");
                };

                previewImage.onerror = () => {
                    showMessage("Image could not be loaded. Check the URL or use a different image.", true);
                };

                previewImage.src = src;
            }

            function updatePreview() {
                const value = input.value.trim();
                if (!value) {
                    showMessage("Enter an image URL to preview it here.");
                    return;
                }

                showImage(value);
            }

            input.addEventListener("input", updatePreview);
            updatePreview();
        });
    }

    function initFallbackImages() {
        const placeholderSrc = "/static/images/car-placeholder.svg";
        const images = document.querySelectorAll("[data-fallback-image]");

        images.forEach((image) => {
            image.addEventListener("error", () => {
                if (image.getAttribute("src") === placeholderSrc) {
                    return;
                }

                image.src = placeholderSrc;
                image.alt = "Image unavailable";
            });
        });
    }

    document.addEventListener("DOMContentLoaded", () => {
        initCatalogFilters();
        initBookingPreview();
        initAdminStatusConfirm();
        initAdminFilters();
        initImagePreview();
        initFallbackImages();
    });
})();
