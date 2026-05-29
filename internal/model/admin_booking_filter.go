package model

import "strings"

const AdminBookingStatusAll = "all"

type AdminBookingFilter struct {
	Search  string
	Status  string
	Page    int
	PerPage int
}

func NormalizeAdminBookingStatus(value string) string {
	status := strings.ToLower(strings.TrimSpace(value))
	if status == AdminBookingStatusAll {
		return AdminBookingStatusAll
	}
	if IsValidBookingStatus(status) {
		return status
	}
	return AdminBookingStatusAll
}
