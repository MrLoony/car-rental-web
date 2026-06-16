package model

type FlashType string

const (
	FlashSuccess FlashType = "success"
	FlashWarning FlashType = "warning"
	FlashError   FlashType = "error"
)

type FlashMessage struct {
	Type    FlashType
	Message string
}
