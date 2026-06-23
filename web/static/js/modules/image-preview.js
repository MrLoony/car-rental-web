import { formatCurrency, on, parseNumber, qs, qsa } from "./utils.js";

const placeholderSrc = "/static/images/car-placeholder.svg";

export function initImagePreview() {
    initAdminCarFormHelpers();
    initFallbackImages();
}

function initAdminCarFormHelpers() {
    const forms = new Set();

    qsa("[data-slug-input], [data-price-input]").forEach((field) => {
        const form = field.closest("form");
        if (form) {
            forms.add(form);
        }
    });

    forms.forEach((form) => {
        initSlugPreview(form);
        initPricePreview(form);
    });
}

function initSlugPreview(form) {
    const slugInput = qs("[data-slug-input]", form);
    const slugPreview = qs("[data-slug-preview]", form);
    const slugSources = qsa("[data-slug-source]", form);

    if (!slugInput || !slugPreview) {
        return;
    }

    let slugEdited = slugInput.value.trim() !== "";

    function updatePreview() {
        const slug = slugInput.value.trim() || suggestedSlug(slugSources) || "example-slug";
        slugPreview.textContent = `/cars/${slug}`;
    }

    on(slugInput, "input", () => {
        slugEdited = slugInput.value.trim() !== "";
        updatePreview();
    });

    slugSources.forEach((source) => {
        on(source, "input", () => {
            if (!slugEdited) {
                updatePreview();
            }
        });
    });

    updatePreview();
}

function suggestedSlug(inputs) {
    return inputs
        .map((input) => input.value.trim())
        .filter(Boolean)
        .join(" ")
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "");
}

function initPricePreview(form) {
    const priceInput = qs("[data-price-input]", form);
    const output = qs("[data-price-preview]", form);
    if (!priceInput || !output) {
        return;
    }

    function updatePrice() {
        const price = parseNumber(priceInput.value);
        output.textContent = price > 0 ? `Displayed as ${formatCurrency(price)} / day.` : "Shown as a daily rental price.";
    }

    on(priceInput, "input", updatePrice);
    updatePrice();
}

function initFallbackImages() {
    qsa("[data-fallback-image]").forEach((image) => {
        image.addEventListener("error", () => {
            if (image.getAttribute("src") === placeholderSrc) {
                return;
            }

            image.src = placeholderSrc;
            image.alt = "Image unavailable";
        });
    });
}
