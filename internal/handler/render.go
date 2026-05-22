package handler

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/model"
)

type TemplateData struct {
	Title   string
	AppName string
	Cars    []model.Car
	Car     model.Car
}

func render(w http.ResponseWriter, page string, data TemplateData) error {
	tmpl, err := template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/"+page,
	)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = buf.WriteTo(w)
	return err
}
