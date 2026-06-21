import { on, parseNumber, qs, qsa } from "./utils.js";

const collapsedStoragePrefix = "dashboard-section-collapsed:";
const activeFilterClass = "dashboard-filter-chip-active";

export function initDashboard() {
    initDashboardSections();
    initRecentActivityFilter();
    initMetricHighlights();
}

function initDashboardSections() {
    qsa("[data-dashboard-section]").forEach((section) => {
        const sectionName = section.dataset.dashboardSection;
        const panel = qs("[data-dashboard-panel]", section);
        const toggle = qs("[data-dashboard-toggle]", section);

        if (!sectionName || !panel || !toggle) {
            return;
        }

        const isCollapsed = getStoredCollapsedState(sectionName);
        setSectionCollapsed(panel, toggle, isCollapsed);

        on(toggle, "click", () => {
            const nextCollapsed = !panel.hidden;
            setSectionCollapsed(panel, toggle, nextCollapsed);
            setStoredCollapsedState(sectionName, nextCollapsed);
        });
    });
}

function initRecentActivityFilter() {
    const filters = qsa("[data-dashboard-filter]");
    const rows = qsa("[data-booking-status]");
    const emptyState = qs("[data-dashboard-filter-empty]");

    if (!filters.length || !rows.length) {
        return;
    }

    filters.forEach((filter) => {
        on(filter, "click", () => {
            const status = filter.dataset.dashboardFilter || "all";
            let visibleCount = 0;

            filters.forEach((item) => {
                const isActive = item === filter;
                item.classList.toggle(activeFilterClass, isActive);
                item.setAttribute("aria-pressed", isActive ? "true" : "false");
            });

            rows.forEach((row) => {
                const rowStatus = row.dataset.bookingStatus || "";
                const isVisible = status === "all" || rowStatus === status;
                row.hidden = !isVisible;
                if (isVisible) {
                    visibleCount += 1;
                }
            });

            if (emptyState) {
                emptyState.classList.toggle("hidden", visibleCount > 0);
            }
        });
    });
}

function initMetricHighlights() {
    qsa("[data-metric-card]").forEach((card) => {
        const metricName = card.dataset.metricCard;
        const metricValue = parseNumber(card.dataset.metricValue);

        if (metricValue <= 0) {
            return;
        }

        if (metricName === "pending") {
            card.classList.add("metric-card-attention");
        } else if (metricName === "cancelled") {
            card.classList.add("metric-card-caution");
        } else if (metricName === "completed") {
            card.classList.add("metric-card-positive");
        }
    });
}

function setSectionCollapsed(panel, toggle, isCollapsed) {
    panel.hidden = isCollapsed;
    toggle.setAttribute("aria-expanded", isCollapsed ? "false" : "true");

    const label = qs("[data-dashboard-toggle-label]", toggle);
    if (label) {
        label.textContent = isCollapsed ? "Expand" : "Collapse";
    }
}

function getStoredCollapsedState(sectionName) {
    try {
        return window.localStorage.getItem(`${collapsedStoragePrefix}${sectionName}`) === "true";
    } catch {
        return false;
    }
}

function setStoredCollapsedState(sectionName, isCollapsed) {
    try {
        window.localStorage.setItem(`${collapsedStoragePrefix}${sectionName}`, String(isCollapsed));
    } catch {
        // Ignore storage failures so dashboard controls still work in private or restricted modes.
    }
}
