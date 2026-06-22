import { on, qsa } from "./utils.js";

export function initPrintSummary() {
    qsa("[data-print-booking-summary]").forEach((button) => {
        button.hidden = false;
        button.classList.remove("hidden");

        on(button, "click", () => {
            window.print();
        });
    });
}
