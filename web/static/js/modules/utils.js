export function qs(selector, root = document) {
    return root.querySelector(selector);
}

export function qsa(selector, root = document) {
    return Array.from(root.querySelectorAll(selector));
}

export function on(element, event, handler) {
    if (!element) {
        return;
    }

    element.addEventListener(event, handler);
}

export function debounce(fn, delay) {
    let timeoutID;

    return function debounced(...args) {
        window.clearTimeout(timeoutID);
        timeoutID = window.setTimeout(() => fn.apply(this, args), delay);
    };
}

export function formatCurrency(value) {
    const amount = Number.parseFloat(value);
    if (Number.isNaN(amount)) {
        return "$0.00";
    }

    return `$${amount.toFixed(2)}`;
}

export function parseNumber(value) {
    const parsed = Number.parseFloat(value);
    return Number.isNaN(parsed) ? 0 : parsed;
}

export function submitForm(form) {
    if (typeof form.requestSubmit === "function") {
        form.requestSubmit();
        return;
    }

    form.submit();
}
