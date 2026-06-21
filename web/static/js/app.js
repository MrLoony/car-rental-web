import { initAdminActions } from "./modules/admin-actions.js";
import { initAdminTables } from "./modules/admin-tables.js";
import { initBookingPreview } from "./modules/booking-preview.js";
import { initCatalogFilters } from "./modules/catalog-filters.js";
import { initCarDetail } from "./modules/car-detail.js";
import { initDashboard } from "./modules/dashboard.js";
import { initFlash } from "./modules/flash.js";
import { initFormHelpers } from "./modules/form-helpers.js";
import { initImagePreview } from "./modules/image-preview.js";

document.addEventListener("DOMContentLoaded", () => {
    initCatalogFilters();
    initCarDetail();
    initDashboard();
    initBookingPreview();
    initImagePreview();
    initAdminActions();
    initAdminTables();
    initFlash();
    initFormHelpers();
});
