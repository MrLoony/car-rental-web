package model

// CarForm holds raw admin car form input.
type CarForm struct {
	CategoryID string
	Brand      string
	Model      string
	Slug       string
	Year       string

	PricePerDay  string
	Transmission string
	FuelType     string
	Seats        string
	ImageURL     string

	IsAvailable bool
	Errors      map[string]string
}

func NewCarForm() CarForm {
	return CarForm{
		IsAvailable: true,
		Errors:      make(map[string]string),
	}
}

func (f CarForm) HasErrors() bool {
	return len(f.Errors) > 0
}
