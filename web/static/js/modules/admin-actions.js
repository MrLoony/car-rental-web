import { qsa } from "./utils.js";

export function initAdminActions() {
    initConfirmActions();
}

function initConfirmActions() {
    const forms = qsa("[data-confirm-action], [data-admin-status-form]");

    forms.forEach((form) => {
        form.addEventListener("submit", (event) => {
            const message = form.dataset.confirmMessage || "Continue with this admin action?";
            if (!window.confirm(message)) {
                event.preventDefault();
            }
        });
    });
}
