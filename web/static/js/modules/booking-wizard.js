import { calculateBillingDays, parseLocalDateTime } from "./booking-preview.js";
import { debounce, formatCurrency, on, qs, qsa } from "./utils.js";

const draftFields = ["pickup_at", "return_at", "customer_name", "customer_email", "customer_phone", "message"];
const saveDelayMs = 250;

export function initBookingWizard() {
    const form = getWizardForm();
    if (!form) {
        return;
    }

    const context = createWizardContext(form);
    if (!context) {
        return;
    }

    restoreDraft(context);
    initSteps(context);
    bindWizardEvents(context);
    updateReview(context);
    saveDraft(context);
}

function getWizardForm() {
    return qs("[data-booking-wizard-form]");
}

function createWizardContext(form) {
    const steps = qsa("[data-wizard-step]", form);
    if (!steps.length) {
        return null;
    }

    return {
        form,
        steps,
        currentStep: 0,
        draftKey: `carRentalBookingDraft:${form.dataset.carSlug || "unknown"}`,
        carName: form.dataset.carName || "Selected vehicle",
        pricePerDay: Number.parseFloat(form.dataset.pricePerDay || "0"),
        pickupInput: qs("[name='pickup_at']", form),
        returnInput: qs("[name='return_at']", form),
        nameInput: qs("[name='customer_name']", form),
        emailInput: qs("[name='customer_email']", form),
        phoneInput: qs("[name='customer_phone']", form),
        messageInput: qs("[name='message']", form),
        progress: qs("[data-wizard-progress]"),
        progressItems: qsa("[data-wizard-progress-item]"),
        backButton: qs("[data-wizard-back]", form),
        nextButton: qs("[data-wizard-next]", form),
        submitButton: qs("[data-wizard-submit]", form),
        hint: qs("[data-wizard-hint]", form),
        draftHint: qs("[data-booking-draft-hint]", form),
        review: {
            vehicle: qs("[data-review-vehicle]", form),
            pickup: qs("[data-review-pickup]", form),
            return: qs("[data-review-return]", form),
            duration: qs("[data-review-duration]", form),
            billingDays: qs("[data-review-billing-days]", form),
            total: qs("[data-review-total]", form),
            name: qs("[data-review-name]", form),
            email: qs("[data-review-email]", form),
            phone: qs("[data-review-phone]", form),
            message: qs("[data-review-message]", form),
        },
    };
}

function initSteps(context) {
    context.form.classList.add("booking-wizard-active");
    context.progress?.classList.remove("hidden");
    context.nextButton?.classList.remove("hidden");

    context.progressItems.forEach((item) => {
        on(item, "click", () => {
            const nextIndex = Number.parseInt(item.dataset.stepIndex || "0", 10);
            goToStep(context, nextIndex);
        });
    });

    goToStep(context, 0);
}

function bindWizardEvents(context) {
    const debouncedSave = debounce(() => {
        saveDraft(context);
        updateReview(context);
    }, saveDelayMs);

    draftFields.forEach((fieldName) => {
        const field = qs(`[name='${fieldName}']`, context.form);
        if (!field) {
            return;
        }

        on(field, "input", debouncedSave);
        on(field, "change", debouncedSave);
    });

    qsa("[data-booking-duration-button], [data-use-window]", context.form).forEach((button) => {
        on(button, "click", () => {
            window.setTimeout(() => {
                saveDraft(context);
                updateReview(context);
            }, 0);
        });
    });

    on(context.nextButton, "click", () => nextStep(context));
    on(context.backButton, "click", () => previousStep(context));
    on(context.form, "submit", () => clearDraft(context));
}

function goToStep(context, index) {
    const nextIndex = Math.max(0, Math.min(index, context.steps.length - 1));
    context.currentStep = nextIndex;

    context.steps.forEach((step, stepIndex) => {
        const active = stepIndex === nextIndex;
        step.classList.toggle("booking-wizard-step-active", active);
        step.hidden = !active;
    });

    updateProgress(context);
    updateReview(context);
    updateWizardControls(context);
    const heading = context.steps[nextIndex]?.querySelector("h3");
    if (heading) {
        heading.tabIndex = -1;
        heading.focus();
    }
}

function nextStep(context) {
    const warnings = validateCurrentStepSoft(context);
    if (warnings.length) {
        showWizardHint(context, warnings.join(" "));
    } else {
        showWizardHint(context, "");
    }

    goToStep(context, context.currentStep + 1);
}

function previousStep(context) {
    showWizardHint(context, "");
    goToStep(context, context.currentStep - 1);
}

function updateProgress(context) {
    context.progressItems.forEach((item, index) => {
        const active = index === context.currentStep;
        const complete = index < context.currentStep;

        item.classList.toggle("booking-wizard-progress-item-active", active);
        item.classList.toggle("booking-wizard-progress-item-complete", complete);
        item.setAttribute("aria-current", active ? "step" : "false");
    });
}

function updateWizardControls(context) {
    const isFirstStep = context.currentStep === 0;
    const isLastStep = context.currentStep === context.steps.length - 1;

    context.backButton?.classList.toggle("hidden", isFirstStep);
    context.nextButton?.classList.toggle("hidden", isLastStep);
    context.submitButton?.classList.toggle("hidden", !isLastStep);
}

function updateReview(context) {
    const pickupDate = parseLocalDateTime(context.pickupInput?.value);
    const returnDate = parseLocalDateTime(context.returnInput?.value);
    const hasValidDates = pickupDate && returnDate && returnDate > pickupDate;
    const billingDays = hasValidDates ? calculateBillingDays(pickupDate, returnDate) : 0;
    const estimatedTotal = billingDays * context.pricePerDay;

    setText(context.review.vehicle, context.carName);
    setText(context.review.pickup, pickupDate ? formatReadableDateTime(pickupDate) : "Not selected");
    setText(context.review.return, returnDate ? formatReadableDateTime(returnDate) : "Not selected");
    setText(context.review.duration, hasValidDates ? formatDuration(calculateDurationHours(pickupDate, returnDate)) : "--");
    setText(context.review.billingDays, billingDays ? `${billingDays} ${billingDays === 1 ? "day" : "days"}` : "--");
    setText(context.review.total, billingDays ? formatCurrency(estimatedTotal) : "--");
    setText(context.review.name, valueOrFallback(context.nameInput?.value, "Not entered"));
    setText(context.review.email, valueOrFallback(context.emailInput?.value, "Not entered"));
    setText(context.review.phone, valueOrFallback(context.phoneInput?.value, "Not entered"));
    setText(context.review.message, valueOrFallback(context.messageInput?.value, "No message"));
}

function validateCurrentStepSoft(context) {
    if (context.currentStep === 0) {
        return validateDateStep(context);
    }
    if (context.currentStep === 1) {
        return validateContactStep(context);
    }

    return [];
}

function validateDateStep(context) {
    const warnings = [];
    const pickupDate = parseLocalDateTime(context.pickupInput?.value);
    const returnDate = parseLocalDateTime(context.returnInput?.value);

    if (!pickupDate) {
        warnings.push("Pickup time is missing.");
    }
    if (!returnDate) {
        warnings.push("Return time is missing.");
    }
    if (pickupDate && returnDate && returnDate <= pickupDate) {
        warnings.push("Return time should be after pickup time.");
    }

    return warnings;
}

function validateContactStep(context) {
    const warnings = [];

    if (!context.nameInput?.value.trim()) {
        warnings.push("Name is missing.");
    }
    if (!context.emailInput?.value.trim()) {
        warnings.push("Email is missing.");
    }
    if (!context.phoneInput?.value.trim()) {
        warnings.push("Phone is missing.");
    }

    return warnings;
}

function saveDraft(context) {
    const draft = collectFormData(context.form);

    try {
        window.sessionStorage.setItem(context.draftKey, JSON.stringify(draft));
    } catch {
        // Draft saving is a convenience feature; the SSR form still works.
    }
}

function restoreDraft(context) {
    let draft;

    try {
        draft = JSON.parse(window.sessionStorage.getItem(context.draftKey) || "null");
    } catch {
        draft = null;
    }

    if (!draft || typeof draft !== "object") {
        return;
    }

    let restored = false;
    draftFields.forEach((fieldName) => {
        const field = qs(`[name='${fieldName}']`, context.form);
        const draftValue = draft[fieldName];
        if (!field || field.value || typeof draftValue !== "string" || !draftValue) {
            return;
        }

        field.value = draftValue;
        field.dispatchEvent(new Event("input", { bubbles: true }));
        restored = true;
    });

    if (restored) {
        showDraftHint(context, "Draft restored from this browser session.");
        document.dispatchEvent(new CustomEvent("app:toast", {
            detail: {
                type: "info",
                message: "Booking draft restored.",
            },
        }));
    }
}

function clearDraft(context) {
    try {
        window.sessionStorage.removeItem(context.draftKey);
    } catch {
        // Ignore storage failures.
    }
}

function collectFormData(form) {
    const data = {};

    draftFields.forEach((fieldName) => {
        data[fieldName] = qs(`[name='${fieldName}']`, form)?.value || "";
    });

    return data;
}

function showWizardHint(context, message) {
    if (!context.hint) {
        return;
    }

    context.hint.textContent = message;
    context.hint.classList.toggle("hidden", !message);
}

function showDraftHint(context, message) {
    if (!context.draftHint) {
        return;
    }

    context.draftHint.textContent = message;
    context.draftHint.classList.remove("hidden");
}

function setText(element, value) {
    if (element) {
        element.textContent = value;
    }
}

function valueOrFallback(value, fallback) {
    const trimmed = String(value || "").trim();
    return trimmed || fallback;
}

function calculateDurationHours(pickupDate, returnDate) {
    return (returnDate - pickupDate) / 1000 / 60 / 60;
}

function formatDuration(hours) {
    const roundedHours = Math.round(hours * 10) / 10;
    const days = Math.floor(hours / 24);
    const remainingHours = Math.round(hours - days * 24);

    if (days >= 1 && remainingHours > 0) {
        return `${days}d ${remainingHours}h`;
    }
    if (days >= 1) {
        return days === 1 ? "1 day" : `${days} days`;
    }

    return `${roundedHours} hours`;
}

function formatReadableDateTime(date) {
    return new Intl.DateTimeFormat(undefined, {
        month: "short",
        day: "2-digit",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
    }).format(date);
}
