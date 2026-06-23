import { on, qs, qsa } from "./utils.js";

export function initCarGallery() {
    qsa("[data-car-gallery]").forEach(initGallery);
}

function initGallery(gallery) {
    const mainImage = qs("[data-gallery-main-image]", gallery);
    const thumbnails = qsa("[data-gallery-thumbnail]", gallery);
    const previousButton = qs("[data-gallery-prev]", gallery);
    const nextButton = qs("[data-gallery-next]", gallery);
    const counter = qs("[data-gallery-counter]", gallery);

    if (!mainImage || thumbnails.length === 0) {
        return;
    }

    const items = thumbnails.map((thumbnail, index) => ({
        index,
        src: thumbnail.dataset.gallerySrc || "",
        alt: thumbnail.dataset.galleryAlt || mainImage.alt || "Vehicle image",
        thumbnail,
    })).filter((item) => item.src !== "");

    if (items.length === 0) {
        return;
    }

    let activeIndex = clampIndex(activeThumbnailIndex(thumbnails), items.length);
    revealControls(previousButton, nextButton);

    function showImage(index) {
        activeIndex = clampIndex(index, items.length);
        const activeItem = items[activeIndex];

        mainImage.src = activeItem.src;
        mainImage.alt = activeItem.alt;

        items.forEach((item, itemIndex) => {
            const isActive = itemIndex === activeIndex;
            item.thumbnail.classList.toggle("car-gallery-thumbnail-active", isActive);
            item.thumbnail.setAttribute("aria-pressed", String(isActive));
            if (isActive) {
                item.thumbnail.setAttribute("aria-current", "true");
            } else {
                item.thumbnail.removeAttribute("aria-current");
            }
        });

        if (counter) {
            counter.textContent = `${activeIndex + 1} / ${items.length}`;
        }
    }

    thumbnails.forEach((thumbnail, index) => {
        on(thumbnail, "click", () => showImage(index));
        on(thumbnail, "keydown", (event) => {
            if (event.key !== "Enter" && event.key !== " ") {
                return;
            }

            event.preventDefault();
            showImage(index);
        });
    });

    on(previousButton, "click", () => showImage(activeIndex - 1));
    on(nextButton, "click", () => showImage(activeIndex + 1));
    on(gallery, "keydown", (event) => {
        if (event.key === "ArrowLeft") {
            event.preventDefault();
            showImage(activeIndex - 1);
        }
        if (event.key === "ArrowRight") {
            event.preventDefault();
            showImage(activeIndex + 1);
        }
    });

    showImage(activeIndex);
}

function revealControls(...controls) {
    controls.forEach((control) => {
        if (control) {
            control.hidden = false;
        }
    });
}

function activeThumbnailIndex(thumbnails) {
    const index = thumbnails.findIndex((thumbnail) => thumbnail.getAttribute("aria-current") === "true");
    return index < 0 ? 0 : index;
}

function clampIndex(index, length) {
    if (length <= 0) {
        return 0;
    }

    if (index < 0) {
        return length - 1;
    }

    if (index >= length) {
        return 0;
    }

    return index;
}
