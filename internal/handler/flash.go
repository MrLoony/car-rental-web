package handler

import (
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/gorilla/sessions"
)

const (
	flashTypeSessionKey    = "flash_type"
	flashMessageSessionKey = "flash_message"
)

func SetFlash(session *sessions.Session, flash model.FlashMessage) {
	session.Values[flashTypeSessionKey] = string(flash.Type)
	session.Values[flashMessageSessionKey] = flash.Message
}

func GetFlash(session *sessions.Session) (*model.FlashMessage, bool) {
	flashType, typeOK := session.Values[flashTypeSessionKey].(string)
	message, messageOK := session.Values[flashMessageSessionKey].(string)

	delete(session.Values, flashTypeSessionKey)
	delete(session.Values, flashMessageSessionKey)

	if !typeOK || !messageOK || flashType == "" || message == "" {
		return nil, false
	}

	return &model.FlashMessage{
		Type:    model.FlashType(flashType),
		Message: message,
	}, true
}

func (h *Handler) setFlash(w http.ResponseWriter, r *http.Request, flash model.FlashMessage) error {
	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}

	SetFlash(session, flash)
	return session.Save(r, w)
}

func (h *Handler) clearAdminSessionWithFlash(w http.ResponseWriter, r *http.Request, flash model.FlashMessage) error {
	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}

	clearAdminSessionValues(session.Values)
	SetFlash(session, flash)
	return session.Save(r, w)
}

func (h *Handler) popFlash(w http.ResponseWriter, r *http.Request) (*model.FlashMessage, error) {
	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		return nil, err
	}

	flash, ok := GetFlash(session)
	if !ok {
		return nil, nil
	}

	if err := session.Save(r, w); err != nil {
		return nil, err
	}

	return flash, nil
}

func (h *Handler) redirectWithFlash(w http.ResponseWriter, r *http.Request, url string, flash model.FlashMessage) {
	if err := h.setFlash(w, r, flash); err != nil {
		h.renderServerError(w, r, err)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}
