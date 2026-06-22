package handler

import (
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/model"
)

const recentBookingsDashboardLimit = 10

func (h *Handler) AdminIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dashboardRange := model.NormalizeDashboardRange(r.URL.Query().Get("range"))

		stats, err := h.bookingService.GetBookingStats(r.Context(), dashboardRange)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}
		revenueStats, err := h.bookingService.GetRevenueStats(r.Context(), dashboardRange)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}
		recentBookings, err := h.bookingService.GetRecentBookings(r.Context(), recentBookingsDashboardLimit, dashboardRange)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		data := TemplateData{
			Title:                  "Admin",
			AppName:                h.appName,
			BookingStats:           stats,
			RevenueStats:           revenueStats,
			RecentBookings:         recentBookings,
			SelectedDashboardRange: dashboardRange,
			DashboardRangeOptions:  dashboardRangeOptions(dashboardRange),
		}

		if err := h.render(w, r, "admin/index.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func dashboardRangeOptions(selected model.DashboardRange) []model.DashboardRangeOption {
	return []model.DashboardRangeOption{
		{
			Value:  model.DashboardRangeAll,
			Label:  "All Time",
			URL:    "/admin?range=all",
			Active: selected == model.DashboardRangeAll,
		},
		{
			Value:  model.DashboardRangeLast30Days,
			Label:  "Last 30 Days",
			URL:    "/admin?range=30d",
			Active: selected == model.DashboardRangeLast30Days,
		},
		{
			Value:  model.DashboardRangeThisMonth,
			Label:  "This Month",
			URL:    "/admin?range=month",
			Active: selected == model.DashboardRangeThisMonth,
		},
	}
}

func (h *Handler) AdminCleanupPrefills() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.bookingPrefillService.CleanupExpiredPrefills(r.Context()); err != nil {
			h.renderServerError(w, r, err)
			return
		}

		h.redirectWithFlash(w, r, "/admin", model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Expired prefill tokens cleaned successfully.",
		})
	}
}
