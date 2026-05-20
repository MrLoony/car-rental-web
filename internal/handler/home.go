package handler

import "net/http"

func Home(appName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			Title:   "Home",
			AppName: appName,
		}

		if err := render(w, "home.html", data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}
