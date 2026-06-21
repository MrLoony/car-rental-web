import { on, qsa } from "./utils.js";

const toastDurations = {
    success: 5000,
    info: 6000,
    warning: 8000,
    error: 10000,
};

const maxVisibleToasts = 4;
const dismissAnimationMs = 180;
let toastRegion;
let initialized = false;

export function initToasts() {
    toastRegion = document.querySelector("[data-toast-region]");
    if (!toastRegion || initialized) {
        return;
    }

    initialized = true;
    toastRegion.setAttribute("role", "status");

    document.addEventListener("app:toast", (event) => {
        showToast(event.detail || {});
    });

    qsa("[data-server-flash]").forEach((flash) => {
        const toast = showToast({
            type: flash.dataset.toastType,
            message: decodeServerFlashMessage(flash.dataset.toastMessage),
        });

        if (toast) {
            flash.hidden = true;
        }
    });
}

export function showToast({ type = "info", message = "", duration } = {}) {
    if (!toastRegion) {
        toastRegion = document.querySelector("[data-toast-region]");
    }
    if (!toastRegion) {
        return null;
    }

    const normalizedType = normalizeType(type);
    const normalizedMessage = String(message || "").trim();
    if (!normalizedMessage) {
        return null;
    }

    while (toastRegion.children.length >= maxVisibleToasts) {
        dismissToast(toastRegion.firstElementChild, { immediate: true });
    }

    const toast = buildToast(normalizedType, normalizedMessage);
    const state = {
        duration: Number.isFinite(duration) ? duration : toastDurations[normalizedType],
        remaining: Number.isFinite(duration) ? duration : toastDurations[normalizedType],
        startedAt: 0,
        timeoutID: 0,
        isPaused: false,
    };

    toastRegion.append(toast);
    startToastTimer(toast, state);

    on(toast, "mouseenter", () => pauseToast(toast, state));
    on(toast, "mouseleave", () => resumeToast(toast, state));
    on(toast, "focusin", () => pauseToast(toast, state));
    on(toast, "focusout", () => resumeToast(toast, state));

    return toast;
}

export function dismissToast(toast, options = {}) {
    if (!toast) {
        return;
    }

    window.clearTimeout(Number(toast.dataset.timeoutID || 0));

    if (options.immediate) {
        toast.remove();
        return;
    }

    toast.classList.add("toast-dismissing");
    window.setTimeout(() => toast.remove(), dismissAnimationMs);
}

function buildToast(type, message) {
    const toast = document.createElement("div");
    toast.className = `toast toast-${type}`;
    toast.dataset.toast = "";

    const body = document.createElement("div");
    body.className = "min-w-0 flex-1";

    const label = document.createElement("p");
    label.className = "text-xs font-semibold uppercase tracking-wide";
    label.textContent = toastLabel(type);

    const text = document.createElement("p");
    text.className = "mt-1 text-sm leading-5 text-slate-700 dark:text-slate-200";
    text.textContent = message;

    const closeButton = document.createElement("button");
    closeButton.className = "toast-close";
    closeButton.type = "button";
    closeButton.setAttribute("aria-label", "Dismiss notification");
    closeButton.textContent = "Close";
    on(closeButton, "click", () => dismissToast(toast));

    const progress = document.createElement("div");
    progress.className = "toast-progress";
    progress.dataset.toastProgress = "";

    body.append(label, text);
    toast.append(body, closeButton, progress);

    return toast;
}

function startToastTimer(toast, state) {
    state.startedAt = Date.now();
    setProgress(toast, state.remaining / state.duration);
    state.timeoutID = window.setTimeout(() => dismissToast(toast), state.remaining);
    toast.dataset.timeoutID = String(state.timeoutID);
    animateProgress(toast, state);
}

function pauseToast(toast, state) {
    if (state.isPaused) {
        return;
    }

    state.isPaused = true;
    window.clearTimeout(state.timeoutID);
    state.remaining = Math.max(0, state.remaining - (Date.now() - state.startedAt));
    setProgress(toast, state.remaining / state.duration);
}

function resumeToast(toast, state) {
    if (!state.isPaused || state.remaining <= 0) {
        return;
    }

    state.isPaused = false;
    startToastTimer(toast, state);
}

function animateProgress(toast, state) {
    if (state.isPaused || !toast.isConnected) {
        return;
    }

    const elapsed = Date.now() - state.startedAt;
    const ratio = Math.max(0, (state.remaining - elapsed) / state.duration);
    setProgress(toast, ratio);

    if (ratio > 0) {
        window.requestAnimationFrame(() => animateProgress(toast, state));
    }
}

function setProgress(toast, ratio) {
    const progress = toast.querySelector("[data-toast-progress]");
    if (!progress) {
        return;
    }

    progress.style.transform = `scaleX(${Math.max(0, Math.min(1, ratio))})`;
}

function normalizeType(type) {
    return Object.hasOwn(toastDurations, type) ? type : "info";
}

function toastLabel(type) {
    if (type === "success") {
        return "Success";
    }
    if (type === "warning") {
        return "Warning";
    }
    if (type === "error") {
        return "Error";
    }

    return "Info";
}

function decodeServerFlashMessage(message = "") {
    try {
        return decodeURIComponent(message.replace(/\+/g, " "));
    } catch {
        return message;
    }
}
