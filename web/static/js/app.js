import { initAdminActions } from "./modules/admin-actions.js";
import { initAdminTables } from "./modules/admin-tables.js";
import { initBookingPreview } from "./modules/booking-preview.js";
import { initBookingWizard } from "./modules/booking-wizard.js";
import { initCatalogFilters } from "./modules/catalog-filters.js";
import { initCarDetail } from "./modules/car-detail.js";
import { initDashboard } from "./modules/dashboard.js";
import { initFavorites } from "./modules/favorites.js";
import { initFlash } from "./modules/flash.js";
import { initFormHelpers } from "./modules/form-helpers.js";
import { initImagePreview } from "./modules/image-preview.js";
import { initTheme } from "./modules/theme.js";
import { initToasts } from "./modules/toast.js";

document.addEventListener("DOMContentLoaded", () => {
    initToasts();
    initTheme();
    initFavorites();
    initCatalogFilters();
    initCarDetail();
    initDashboard();
    initBookingPreview();
    initBookingWizard();
    initImagePreview();
    initAdminActions();
    initAdminTables();
    initFlash();
    initFormHelpers();
});
