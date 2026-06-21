import { formatCurrency, on, parseNumber, qs, qsa } from "./utils.js";

const highlightDurationMs = 1200;

export function initBookingPreview() {
    const form = qs("[data-booking-form]");
    if (!form) {
        return;
    }

    const context = createBookingContext(form);
    if (!context) {
        return;
    }

    initDurationButtons(context);
    initDateWarnings(context);
    initSuggestedWindowButtons(context);
    updateBookingSummary(context);
}

function createBookingContext(form) {
    const pickupInput = qs("[data-pickup-input]", form);
    const returnInput = qs("[data-return-input]", form);
    const pricePerDay = parseNumber(form.dataset.pricePerDay);

    if (!pickupInput || !returnInput || pricePerDay <= 0) {
        return null;
    }

    return {
        form,
        pickupInput,
        returnInput,
        pricePerDay,
        summary: qs("[data-booking-summary]", form),
        formCard: qs("[data-booking-form-card]", form),
        pickupOutput: qs("[data-summary-pickup]", form),
        returnOutput: qs("[data-summary-return]", form),
        durationOutput: qs("[data-summary-duration]", form),
        billingDaysOutput: qs("[data-billing-days]", form),
        estimatedTotalOutput: qs("[data-estimated-total]", form),
        warning: qs("[data-booking-warning]", form),
        dateWarning: qs("[data-date-warning]", form),
        clientHint: qs("[data-booking-client-hint]", form),
        durationButtons: qsa("[data-booking-duration-button]", form),
        suggestedWindows: qsa("[data-suggested-window]", form),
    };
}

function initDurationButtons(context) {
    context.durationButtons.forEach((button) => {
        on(button, "click", () => {
            const pickupDate = parseLocalDateTime(context.pickupInput.value);
            const durationHours = parseNumber(button.dataset.durationHours);

            if (!pickupDate) {
                showClientHint(context, "Select a pickup time before choosing a quick duration.");
                return;
            }
            if (durationHours <= 0) {
                return;
            }

            const returnDate = new Date(pickupDate.getTime() + durationHours * 60 * 60 * 1000);
            context.returnInput.value = formatDateTimeForInput(returnDate);
            setActiveDurationButton(context, button);
            showClientHint(context, `${button.textContent.trim()} duration applied.`);
            updateBookingSummary(context);
        });
    });
}

function initDateWarnings(context) {
    const handleDateChange = () => {
        clearActiveDurationButtons(context);
        updateBookingSummary(context);
    };

    on(context.pickupInput, "input", handleDateChange);
    on(context.pickupInput, "change", handleDateChange);
    on(context.returnInput, "input", handleDateChange);
    on(context.returnInput, "change", handleDateChange);
}

function initSuggestedWindowButtons(context) {
    context.suggestedWindows.forEach((windowCard) => {
        const button = qs("[data-use-window]", windowCard);
        if (!button) {
            return;
        }

        on(button, "click", () => {
            const pickupValue = windowCard.dataset.windowPickup;
            const returnValue = windowCard.dataset.windowReturn;
            if (!pickupValue || !returnValue) {
                return;
            }

            context.pickupInput.value = pickupValue;
            context.returnInput.value = returnValue;
            clearActiveDurationButtons(context);
            updateBookingSummary(context);
            showClientHint(context, "Suggested availability window applied.");
            highlightBookingArea(context);
        });
    });
}

function updateBookingSummary(context) {
    const pickupDate = parseLocalDateTime(context.pickupInput.value);
    const returnDate = parseLocalDateTime(context.returnInput.value);

    setText(context.pickupOutput, pickupDate ? formatReadableDateTime(pickupDate) : "Not selected");
    setText(context.returnOutput, returnDate ? formatReadableDateTime(returnDate) : "Not selected");

    if (!pickupDate || !returnDate) {
        setNeutralSummary(context);
        showDateWarning(context, "");
        return;
    }

    if (returnDate < pickupDate) {
        setInvalidSummary(context, "Return time is before pickup time.");
        showDateWarning(context, "Return time is before pickup time. Please choose a later return time.");
        return;
    }

    if (returnDate.getTime() === pickupDate.getTime()) {
        setInvalidSummary(context, "Return time must be after pickup time.");
        showDateWarning(context, "Pickup and return times are the same. Choose a later return time.");
        return;
    }

    const durationHours = calculateDurationHours(pickupDate, returnDate);
    const billingDays = calculateBillingDays(pickupDate, returnDate);
    const estimatedTotal = billingDays * context.pricePerDay;

    setText(context.durationOutput, formatDuration(durationHours));
    setText(context.billingDaysOutput, billingDays === 1 ? "1 day" : `${billingDays} days`);
    setText(context.estimatedTotalOutput, formatCurrency(estimatedTotal));
    setText(context.warning, "Estimate updates as you adjust pickup and return times.");
    context.warning?.classList.remove("text-red-600");
    context.warning?.classList.add("text-slate-600");
    showDateWarning(context, "");
}

function setNeutralSummary(context) {
    setText(context.durationOutput, "--");
    setText(context.billingDaysOutput, "--");
    setText(context.estimatedTotalOutput, "--");
    setText(context.warning, "Select pickup and return times to see an estimate.");
    context.warning?.classList.remove("text-red-600");
    context.warning?.classList.add("text-slate-600");
}

function setInvalidSummary(context, message) {
    setText(context.durationOutput, "--");
    setText(context.billingDaysOutput, "--");
    setText(context.estimatedTotalOutput, "--");
    setText(context.warning, message);
    context.warning?.classList.remove("text-slate-600");
    context.warning?.classList.add("text-red-600");
}

function showDateWarning(context, message) {
    if (!context.dateWarning) {
        return;
    }

    if (!message) {
        context.dateWarning.classList.add("hidden");
        context.dateWarning.textContent = "";
        return;
    }

    context.dateWarning.textContent = message;
    context.dateWarning.classList.remove("hidden");
}

function showClientHint(context, message) {
    if (!context.clientHint) {
        return;
    }

    context.clientHint.textContent = message;
    context.clientHint.classList.remove("hidden");
}

function highlightBookingArea(context) {
    const target = context.summary || context.formCard || context.form;
    target.scrollIntoView({ behavior: "smooth", block: "center" });

    [context.summary, context.formCard].forEach((element) => {
        if (!element) {
            return;
        }

        element.classList.add("booking-highlight");
        window.setTimeout(() => {
            element.classList.remove("booking-highlight");
        }, highlightDurationMs);
    });
}

function setActiveDurationButton(context, activeButton) {
    context.durationButtons.forEach((button) => {
        button.classList.toggle("duration-button-active", button === activeButton);
    });
}

function clearActiveDurationButtons(context) {
    context.durationButtons.forEach((button) => {
        button.classList.remove("duration-button-active");
    });
}

function setText(element, value) {
    if (element) {
        element.textContent = value;
    }
}

export function parseLocalDateTime(value) {
    if (!value) {
        return null;
    }

    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? null : date;
}

export function formatDateTimeForInput(date) {
    const year = date.getFullYear();
    const month = pad(date.getMonth() + 1);
    const day = pad(date.getDate());
    const hours = pad(date.getHours());
    const minutes = pad(date.getMinutes());

    return `${year}-${month}-${day}T${hours}:${minutes}`;
}

export function calculateBillingDays(pickupDate, returnDate) {
    const durationHours = calculateDurationHours(pickupDate, returnDate);
    const billingDays = Math.ceil(durationHours / 24);

    return Math.max(1, billingDays);
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

function pad(value) {
    return String(value).padStart(2, "0");
}
