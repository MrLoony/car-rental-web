import { on, qs, qsa } from "./utils.js";

const copiedResetMs = 1400;

export function initCarDetail() {
    const root = qs("[data-car-detail]");
    if (!root) {
        return;
    }

    initCopyPageLink(root);
    initCarImageLightbox(root);
}

function initCopyPageLink(root) {
    qsa("[data-copy-page-url]", root).forEach((button) => {
        on(button, "click", async () => {
            const copied = await copyText(window.location.href);
            const originalText = button.dataset.originalText || button.textContent;

            button.dataset.originalText = originalText;
            button.textContent = copied ? "Copied" : "Copy failed";

            document.dispatchEvent(new CustomEvent("app:toast", {
                detail: {
                    type: copied ? "success" : "warning",
                    message: copied ? "Link copied to clipboard." : "Copy failed. Use the address bar to copy this link.",
                },
            }));

            window.setTimeout(() => {
                button.textContent = originalText;
            }, copiedResetMs);
        });
    });
}

function initCarImageLightbox(root) {
    const trigger = qs("[data-lightbox-trigger]", root);
    const lightbox = qs("[data-image-lightbox]", root);
    const closeButton = qs("[data-lightbox-close]", root);

    if (!trigger || !lightbox || !closeButton) {
        return;
    }

    let previousFocus = null;

    function openLightbox() {
        previousFocus = document.activeElement;
        lightbox.hidden = false;
        lightbox.classList.add("image-lightbox-open");
        lightbox.setAttribute("aria-hidden", "false");
        document.body.classList.add("overflow-hidden");
        closeButton.focus();
    }

    function closeLightbox() {
        lightbox.hidden = true;
        lightbox.classList.remove("image-lightbox-open");
        lightbox.setAttribute("aria-hidden", "true");
        document.body.classList.remove("overflow-hidden");
        if (previousFocus && typeof previousFocus.focus === "function") {
            previousFocus.focus();
        } else {
            trigger.focus();
        }
    }

    on(trigger, "click", openLightbox);
    on(closeButton, "click", closeLightbox);
    on(lightbox, "click", (event) => {
        if (event.target === lightbox) {
            closeLightbox();
        }
    });
    on(document, "keydown", (event) => {
        if (event.key === "Escape" && !lightbox.hidden) {
            closeLightbox();
        }
    });
}

async function copyText(value) {
    if (navigator.clipboard && typeof navigator.clipboard.writeText === "function") {
        try {
            await navigator.clipboard.writeText(value);
            return true;
        } catch {
            return fallbackCopyText(value);
        }
    }

    return fallbackCopyText(value);
}

function fallbackCopyText(value) {
    const input = document.createElement("input");
    input.value = value;
    input.setAttribute("readonly", "");
    input.style.position = "fixed";
    input.style.opacity = "0";
    document.body.append(input);
    input.select();

    let copied = false;
    try {
        copied = document.execCommand("copy");
    } catch {
        copied = false;
    }

    input.remove();
    return copied;
}
