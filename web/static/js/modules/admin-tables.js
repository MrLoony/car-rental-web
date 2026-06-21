import { debounce, on, qs, qsa } from "./utils.js";

const copiedClass = "copy-button-copied";
const activeRowClass = "admin-table-row-active";
const copiedResetMs = 1400;

export function initAdminTables() {
    qsa("[data-admin-table]").forEach((tableRoot) => {
        initVisibleRowFilter(tableRoot);
        initRowHighlighting(tableRoot);
        initCopyButtons(tableRoot);
        updateVisibleRowCount(tableRoot);
    });
}

function initVisibleRowFilter(tableRoot) {
    const filterInput = qs("[data-admin-table-filter]", tableRoot);
    if (!filterInput) {
        return;
    }

    const applyFilter = debounce(() => {
        const query = filterInput.value.trim().toLowerCase();
        const rows = qsa("[data-admin-table-row]", tableRoot);

        rows.forEach((row) => {
            const isVisible = query === "" || row.textContent.toLowerCase().includes(query);
            row.hidden = !isVisible;
        });

        updateVisibleRowCount(tableRoot);
    }, 120);

    on(filterInput, "input", applyFilter);
}

function initRowHighlighting(tableRoot) {
    qsa("[data-admin-table-row]", tableRoot).forEach((row) => {
        on(row, "mouseenter", () => row.classList.add(activeRowClass));
        on(row, "mouseleave", () => row.classList.remove(activeRowClass));
        on(row, "focusin", () => row.classList.add(activeRowClass));
        on(row, "focusout", () => row.classList.remove(activeRowClass));
    });
}

function initCopyButtons(tableRoot) {
    qsa("[data-copy-value]", tableRoot).forEach((button) => {
        on(button, "click", async () => {
            const value = button.dataset.copyValue || "";
            if (!value) {
                return;
            }

            const copied = await copyText(value);
            setCopyButtonState(button, copied);
        });
    });
}

function updateVisibleRowCount(tableRoot) {
    const output = qs("[data-admin-row-count]", tableRoot);
    const emptyState = qs("[data-admin-table-empty]", tableRoot);
    const rows = qsa("[data-admin-table-row]", tableRoot);

    if (!output || !rows.length) {
        return;
    }

    const visibleRows = rows.filter((row) => !row.hidden).length;
    output.textContent = `Showing ${visibleRows} ${visibleRows === 1 ? "row" : "rows"} on this page`;

    if (emptyState) {
        emptyState.classList.toggle("hidden", visibleRows > 0);
    }
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
    const textarea = document.createElement("textarea");
    textarea.value = value;
    textarea.setAttribute("readonly", "");
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.append(textarea);
    textarea.select();

    let copied = false;
    try {
        copied = document.execCommand("copy");
    } catch {
        copied = false;
    }

    textarea.remove();
    return copied;
}

function setCopyButtonState(button, copied) {
    const originalText = button.dataset.originalText || button.textContent;
    button.dataset.originalText = originalText;
    button.textContent = copied ? "Copied" : "Copy failed";
    button.classList.toggle(copiedClass, copied);

    window.setTimeout(() => {
        button.textContent = originalText;
        button.classList.remove(copiedClass);
    }, copiedResetMs);
}
