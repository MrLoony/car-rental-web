package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
	"github.com/MrLoony/car-rental-web/internal/service"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) AdminCarsIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		availability := model.NormalizeAdminCarAvailability(r.URL.Query().Get("availability"))
		page := parsePositiveInt(r.URL.Query().Get("page"), model.DefaultPage)
		filter := model.AdminCarFilter{
			Search:       strings.TrimSpace(r.URL.Query().Get("search")),
			Availability: availability,
			Page:         page,
			PerPage:      model.DefaultPerPage,
		}

		cars, pagination, err := h.carService.ListCarsForAdminPage(r.Context(), filter)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		data := TemplateData{
			Title:                    "Cars",
			AppName:                  h.appName,
			AdminCars:                cars,
			AdminCarFilter:           filter,
			HasActiveAdminCarFilters: hasActiveAdminCarFilters(filter),
			Pagination:               pagination,
		}
		if pagination.HasPrevious {
			data.PaginationPreviousURL = paginationURL(r, pagination.PreviousPage)
		}
		if pagination.HasNext {
			data.PaginationNextURL = paginationURL(r, pagination.NextPage)
		}

		if err := h.render(w, r, "admin/cars/index.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func hasActiveAdminCarFilters(filter model.AdminCarFilter) bool {
	return filter.Search != "" ||
		model.NormalizeAdminCarAvailability(filter.Availability) != model.AdminCarAvailabilityAll
}

func (h *Handler) AdminCarsNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categories, err := h.categoryService.ListCategories(r.Context())
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		data := TemplateData{
			Title:      "New car",
			AppName:    h.appName,
			CarForm:    model.NewCarForm(),
			Categories: categories,
		}

		if err := h.render(w, r, "admin/cars/new.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func (h *Handler) AdminCarsCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form := parseCarForm(r)

		id, form, err := h.carService.CreateCar(r.Context(), form)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		if form.HasErrors() {
			h.renderAdminCarForm(w, r, "admin/cars/new.html", "New car", model.Car{}, form, http.StatusUnprocessableEntity)
			return
		}

		h.redirectWithFlash(w, r, adminCarEditURL(id)+"#vehicle-gallery", model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Car created as unavailable. Add gallery images, then enable it in the public catalog.",
		})
	}
}

func (h *Handler) AdminCarsEdit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseCarID(w, r)
		if !ok {
			return
		}

		car, err := h.carService.GetCarByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrCarNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		categories, err := h.categoryService.ListCategories(r.Context())
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		carImages, err := h.carService.GetCarImages(r.Context(), car.ID)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		data := TemplateData{
			Title:      "Edit " + car.Brand + " " + car.Model,
			AppName:    h.appName,
			AdminCar:   car,
			CarImages:  carImages,
			CarForm:    carToForm(car),
			Categories: categories,
		}

		if err := h.render(w, r, "admin/cars/edit.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func (h *Handler) AdminCarsUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseCarID(w, r)
		if !ok {
			return
		}

		car, err := h.carService.GetCarByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrCarNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		form := parseCarForm(r)

		form, err = h.carService.UpdateCar(r.Context(), id, form)
		if err != nil {
			if errors.Is(err, repository.ErrCarNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		if form.HasErrors() {
			h.renderAdminCarForm(w, r, "admin/cars/edit.html", "Edit "+car.Brand+" "+car.Model, car, form, http.StatusUnprocessableEntity)
			return
		}

		h.redirectWithFlash(w, r, "/admin/cars/"+strconv.FormatInt(id, 10), model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Car updated successfully.",
		})
	}
}

func (h *Handler) AdminCarGalleryCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseCarID(w, r)
		if !ok {
			return
		}

		car, err := h.carService.GetCarByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrCarNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		galleryURL := adminCarGalleryURL(id)

		if err := parseOptionalCarMultipartForm(r); err != nil {
			h.redirectWithFlash(w, r, galleryURL, model.FlashMessage{
				Type:    model.FlashError,
				Message: "The uploaded gallery image could not be processed.",
			})
			return
		}

		imageURL := strings.TrimSpace(r.FormValue("gallery_image_url"))
		uploadedImageURLs, uploaded, err := saveOptionalGalleryImageUploads(r, car.Slug)
		if err != nil {
			h.redirectWithFlash(w, r, galleryURL, model.FlashMessage{
				Type:    model.FlashError,
				Message: err.Error(),
			})
			return
		}

		altText := r.FormValue("gallery_alt_text")
		images := []model.CarImage{}
		if uploaded {
			for _, uploadedImageURL := range uploadedImageURLs {
				images = append(images, model.CarImage{
					CarID:    id,
					ImageURL: uploadedImageURL,
					AltText:  altText,
				})
			}
		} else if imageURL != "" {
			images = append(images, model.CarImage{
				CarID:    id,
				ImageURL: imageURL,
				AltText:  altText,
			})
		}

		_, err = h.carService.AddCarImages(r.Context(), images)
		if err != nil {
			if uploaded {
				removeUploadedCarImages(uploadedImageURLs)
			}
			if errors.Is(err, service.ErrInvalidCarImage) {
				h.redirectWithFlash(w, r, galleryURL, model.FlashMessage{
					Type:    model.FlashError,
					Message: "Image URL or image file is required.",
				})
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		h.redirectWithFlash(w, r, galleryURL, model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: galleryImagesAddedMessage(len(images)),
		})
	}
}

func (h *Handler) AdminCarGallerySetPrimary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseCarID(w, r)
		if !ok {
			return
		}
		imageID, ok := parseCarImageID(w, r)
		if !ok {
			return
		}

		err := h.carService.SetPrimaryCarImage(r.Context(), id, imageID)
		if err != nil {
			if errors.Is(err, repository.ErrCarImageNotFound) {
				h.renderNotFound(w, r)
				return
			}
			if errors.Is(err, service.ErrInvalidCarImage) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		h.redirectWithFlash(w, r, adminCarGalleryURL(id), model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Primary image updated.",
		})
	}
}

func (h *Handler) AdminCarGalleryDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseCarID(w, r)
		if !ok {
			return
		}
		imageID, ok := parseCarImageID(w, r)
		if !ok {
			return
		}

		err := h.carService.DeleteCarImage(r.Context(), id, imageID)
		if err != nil {
			if errors.Is(err, repository.ErrCarImageNotFound) || errors.Is(err, service.ErrCarImageNotFound) {
				h.renderNotFound(w, r)
				return
			}
			if errors.Is(err, service.ErrInvalidCarImage) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		h.redirectWithFlash(w, r, adminCarGalleryURL(id), model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Gallery image deleted.",
		})
	}
}

func (h *Handler) AdminCarAvailabilityUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseCarID(w, r)
		if !ok {
			return
		}

		isAvailable := r.FormValue("is_available") != ""
		err := h.carService.UpdateCarAvailability(r.Context(), id, isAvailable)
		if err != nil {
			if errors.Is(err, repository.ErrCarNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		message := "Car is now unavailable."
		if isAvailable {
			message = "Car is now available."
		}

		h.redirectWithFlash(w, r, "/admin/cars/"+strconv.FormatInt(id, 10), model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: message,
		})
	}
}

func (h *Handler) AdminCarArchive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseCarID(w, r)
		if !ok {
			return
		}

		err := h.carService.ArchiveCar(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrCarNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		h.redirectWithFlash(w, r, "/admin/cars/"+strconv.FormatInt(id, 10), model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Car archived successfully.",
		})
	}
}

func (h *Handler) AdminCarUnarchive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseCarID(w, r)
		if !ok {
			return
		}

		err := h.carService.UnarchiveCar(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrCarNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		h.redirectWithFlash(w, r, "/admin/cars/"+strconv.FormatInt(id, 10), model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Car restored successfully.",
		})
	}
}

func (h *Handler) AdminCarsShow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseCarID(w, r)
		if !ok {
			return
		}

		car, err := h.carService.GetCarByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrCarNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		carImages, err := h.carService.GetCarImages(r.Context(), car.ID)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		data := TemplateData{
			Title:     car.Brand + " " + car.Model,
			AppName:   h.appName,
			AdminCar:  car,
			CarImages: carImages,
		}

		if err := h.render(w, r, "admin/cars/show.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func parseCarID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return 0, false
	}

	return id, true
}

func parseCarImageID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "imageID"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return 0, false
	}

	return id, true
}

func adminCarEditURL(id int64) string {
	return "/admin/cars/" + strconv.FormatInt(id, 10) + "/edit"
}

func adminCarGalleryURL(id int64) string {
	return adminCarEditURL(id) + "#vehicle-gallery"
}

func galleryImagesAddedMessage(count int) string {
	if count == 1 {
		return "Gallery image added."
	}

	return strconv.Itoa(count) + " gallery images added."
}

func parseCarForm(r *http.Request) model.CarForm {
	return model.CarForm{
		CategoryID:   r.FormValue("category_id"),
		Brand:        r.FormValue("brand"),
		Model:        r.FormValue("model"),
		Slug:         r.FormValue("slug"),
		Year:         r.FormValue("year"),
		PricePerDay:  r.FormValue("price_per_day"),
		Transmission: r.FormValue("transmission"),
		FuelType:     r.FormValue("fuel_type"),
		Seats:        r.FormValue("seats"),
		IsAvailable:  r.FormValue("is_available") != "",
	}
}

func parseOptionalCarMultipartForm(r *http.Request) error {
	err := r.ParseMultipartForm(maxCarImageUploadSize)
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return err
	}

	return nil
}

func saveOptionalGalleryImageUploads(r *http.Request, carSlug string) ([]string, bool, error) {
	return saveOptionalCarImageUploadFields(r, "gallery_image_files", carSlug)
}

func saveOptionalCarImageUploadFields(r *http.Request, fieldName, carSlug string) ([]string, bool, error) {
	headers := uploadedFileHeaders(r, fieldName)
	if len(headers) == 0 && fieldName == "gallery_image_files" {
		headers = uploadedFileHeaders(r, "gallery_image_file")
	}
	if len(headers) == 0 {
		return nil, false, nil
	}

	imageURLs := make([]string, 0, len(headers))
	for _, header := range headers {
		file, err := header.Open()
		if err != nil {
			return nil, false, err
		}

		imageURL, err := saveCarImageUpload(file, header, carSlug)
		if err != nil {
			removeUploadedCarImages(imageURLs)
			return nil, false, err
		}

		imageURLs = append(imageURLs, imageURL)
	}

	return imageURLs, true, nil
}

func removeUploadedCarImages(imageURLs []string) {
	for _, imageURL := range imageURLs {
		if strings.HasPrefix(imageURL, "/static/uploads/cars/") {
			_ = os.Remove("web/static" + strings.TrimPrefix(imageURL, "/static"))
		}
	}
}

func uploadedFileHeaders(r *http.Request, fieldName string) []*multipart.FileHeader {
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil
	}

	headers := r.MultipartForm.File[fieldName]
	files := make([]*multipart.FileHeader, 0, len(headers))
	for _, header := range headers {
		if header != nil && header.Filename != "" && header.Size > 0 {
			files = append(files, header)
		}
	}

	return files
}

func addCarFormError(form model.CarForm, field, message string) model.CarForm {
	if form.Errors == nil {
		form.Errors = make(map[string]string)
	}

	form.Errors[field] = message
	return form
}

func (h *Handler) renderAdminCarForm(w http.ResponseWriter, r *http.Request, page, title string, car model.Car, form model.CarForm, status int) {
	categories, err := h.categoryService.ListCategories(r.Context())
	if err != nil {
		h.renderServerError(w, r, err)
		return
	}

	var carImages []model.CarImage
	if car.ID > 0 {
		carImages, err = h.carService.GetCarImages(r.Context(), car.ID)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}
	}

	data := TemplateData{
		Title:      title,
		AppName:    h.appName,
		AdminCar:   car,
		CarImages:  carImages,
		CarForm:    form,
		Categories: categories,
	}

	if err := h.renderWithStatus(w, r, page, data, status); err != nil {
		h.renderServerError(w, r, err)
	}
}

func carToForm(car model.Car) model.CarForm {
	form := model.NewCarForm()
	form.CategoryID = strconv.FormatInt(car.CategoryID, 10)
	form.Brand = car.Brand
	form.Model = car.Model
	form.Slug = car.Slug
	form.Year = strconv.Itoa(car.Year)
	form.PricePerDay = strconv.FormatFloat(car.PricePerDay, 'f', 2, 64)
	form.Transmission = car.Transmission
	form.FuelType = car.FuelType
	form.Seats = strconv.Itoa(car.Seats)
	form.IsAvailable = car.IsAvailable

	return form
}
