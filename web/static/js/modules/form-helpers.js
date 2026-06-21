import { on, qsa } from "./utils.js";

export function initFormHelpers() {
    initSubmitOnceForms();
    initDirtyForms();
}

function initSubmitOnceForms() {
    qsa("[data-submit-once]").forEach((form) => {
        form.addEventListener("submit", () => {
            qsa("[type='submit']", form).forEach((button) => {
                button.disabled = true;
                if (button.dataset.submitLabel) {
                    button.dataset.originalText = button.textContent.trim();
                    button.textContent = button.dataset.submitLabel;
                }
            });
        });
    });
}

function initDirtyForms() {
    qsa("[data-dirty-form]").forEach((form) => {
        let isDirty = false;
        let isSubmitting = false;

        const markDirty = () => {
            if (!isSubmitting) {
                isDirty = true;
            }
        };

        qsa("input, select, textarea", form).forEach((field) => {
            if (field.type === "hidden") {
                return;
            }

            on(field, "input", markDirty);
            on(field, "change", markDirty);
        });

        on(form, "submit", () => {
            isSubmitting = true;
            isDirty = false;
        });

        window.addEventListener("beforeunload", (event) => {
            if (!isDirty || isSubmitting) {
                return;
            }

            event.preventDefault();
            event.returnValue = "";
        });
    });
}
