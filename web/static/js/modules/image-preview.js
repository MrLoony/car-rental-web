import { formatCurrency, on, parseNumber, qs, qsa } from "./utils.js";

const placeholderSrc = "/static/images/car-placeholder.svg";
const maxUploadBytes = 5 * 1024 * 1024;
const allowedImageTypes = new Set(["image/jpeg", "image/png", "image/webp"]);

export function initImagePreview() {
    initAdminImageForms();
    initFallbackImages();
}

function initAdminImageForms() {
    qsa("[data-image-preview]").forEach((preview) => {
        const form = preview.closest("form");
        if (!form) {
            return;
        }

        const context = createImageContext(form, preview);
        if (!context) {
            return;
        }

        initImageURLPreview(context);
        initUploadPreview(context);
        initSlugPreview(form);
        initPricePreview(form);
    });
}

function createImageContext(form, preview) {
    const previewImage = qs("[data-image-preview-img]", preview);
    const previewMessage = qs("[data-image-preview-message]", preview);

    if (!previewImage || !previewMessage) {
        return null;
    }

    return {
        form,
        preview,
        previewImage,
        previewMessage,
        previewSource: qs("[data-image-preview-source]", preview),
        imageInput: qs("[data-image-url-input]", form),
        uploadZone: qs("[data-upload-zone]", form),
        uploadInput: qs("[data-upload-input]", form),
        uploadFilename: qs("[data-upload-filename]", form),
        uploadHint: qs("[data-upload-hint]", form),
        objectURL: null,
    };
}

function initImageURLPreview(context) {
    if (!context.imageInput) {
        return;
    }

    on(context.imageInput, "input", () => {
        if (context.uploadInput?.files?.length) {
            setPreviewSource(context, "Local selected file");
            return;
        }

        updateURLPreview(context);
    });

    updateURLPreview(context);
}

function updateURLPreview(context) {
    const value = context.imageInput?.value.trim() || "";
    if (!value) {
        showPreviewMessage(context, "No image selected.");
        setPreviewSource(context, "No image selected");
        return;
    }

    showPreviewImage(context, value, "URL image");
}

function initUploadPreview(context) {
    if (!context.uploadInput) {
        return;
    }

    on(context.uploadInput, "change", () => {
        updateUploadPreview(context);
    });

    if (!context.uploadZone) {
        return;
    }

    on(context.uploadZone, "click", (event) => {
        if (event.target === context.uploadInput) {
            return;
        }

        context.uploadInput.click();
    });

    ["dragenter", "dragover"].forEach((eventName) => {
        on(context.uploadZone, eventName, (event) => {
            event.preventDefault();
            context.uploadZone.classList.add("upload-dropzone-active");
        });
    });

    ["dragleave", "drop"].forEach((eventName) => {
        on(context.uploadZone, eventName, () => {
            context.uploadZone.classList.remove("upload-dropzone-active");
        });
    });

    on(context.uploadZone, "drop", (event) => {
        event.preventDefault();
        const files = event.dataTransfer?.files;
        if (!files || !files.length) {
            return;
        }

        context.uploadInput.files = files;
        updateUploadPreview(context);
    });
}

function updateUploadPreview(context) {
    const file = context.uploadInput.files?.[0];
    if (!file) {
        setUploadFilename(context, "");
        updateURLPreview(context);
        return;
    }

    setUploadFilename(context, file.name);
    showUploadHint(context, uploadHintForFile(file));

    if (!allowedImageTypes.has(file.type)) {
        showPreviewMessage(context, "Preview unavailable. Please choose a JPEG, PNG, or WebP image.", true);
        setPreviewSource(context, "Selected file needs review");
        return;
    }

    if (file.size > maxUploadBytes) {
        showPreviewMessage(context, "Preview unavailable. This file is larger than 5 MB.", true);
        setPreviewSource(context, "Selected file needs review");
        return;
    }

    if (context.objectURL) {
        URL.revokeObjectURL(context.objectURL);
    }

    context.objectURL = URL.createObjectURL(file);
    showPreviewImage(context, context.objectURL, "Local selected file");
}

function uploadHintForFile(file) {
    if (!allowedImageTypes.has(file.type)) {
        return "This file does not look like a JPEG, PNG, or WebP image. Server validation will reject invalid uploads.";
    }

    if (file.size > maxUploadBytes) {
        return "This file is larger than 5 MB. Server validation will reject oversized uploads.";
    }

    return "Selected file looks compatible. Server validation will confirm it on submit.";
}

function showUploadHint(context, message) {
    if (!context.uploadHint) {
        return;
    }

    context.uploadHint.textContent = message;
    context.uploadHint.classList.toggle("field-warning", message.includes("reject"));
}

function setUploadFilename(context, filename) {
    if (!context.uploadFilename) {
        return;
    }

    if (!filename) {
        context.uploadFilename.classList.add("hidden");
        context.uploadFilename.textContent = "";
        return;
    }

    context.uploadFilename.textContent = `Selected: ${filename}`;
    context.uploadFilename.classList.remove("hidden");
}

function showPreviewMessage(context, message, isError = false) {
    context.previewImage.classList.add("hidden");
    context.previewMessage.textContent = message;
    context.previewMessage.classList.remove("hidden", "text-red-600", "text-slate-500");
    context.previewMessage.classList.add("flex", isError ? "text-red-600" : "text-slate-500");
}

function showPreviewImage(context, src, sourceLabel) {
    context.previewImage.onload = () => {
        context.previewImage.classList.remove("hidden");
        context.previewMessage.classList.add("hidden");
        context.previewMessage.classList.remove("flex");
        setPreviewSource(context, sourceLabel);
    };

    context.previewImage.onerror = () => {
        showPreviewMessage(context, "Image could not be loaded. Check the URL or use a different image.", true);
        setPreviewSource(context, "Preview unavailable");
    };

    context.previewImage.src = src;
}

function setPreviewSource(context, sourceLabel) {
    if (context.previewSource) {
        context.previewSource.textContent = sourceLabel;
    }
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
