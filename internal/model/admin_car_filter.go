package model

import "strings"

const (
	AdminCarAvailabilityAll         = "all"
	AdminCarAvailabilityAvailable   = "available"
	AdminCarAvailabilityUnavailable = "unavailable"
)

type AdminCarFilter struct {
	Search       string
	Availability string
	Page         int
	PerPage      int
}

func NormalizeAdminCarAvailability(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case AdminCarAvailabilityAvailable:
		return AdminCarAvailabilityAvailable
	case AdminCarAvailabilityUnavailable:
		return AdminCarAvailabilityUnavailable
	default:
		return AdminCarAvailabilityAll
	}
}
