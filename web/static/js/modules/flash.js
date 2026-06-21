import { qsa } from "./utils.js";

export function initFlash() {
    qsa("[data-flash-message]").forEach((flash) => {
        flash.setAttribute("tabindex", "-1");
    });
}
