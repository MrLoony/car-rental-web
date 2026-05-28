package handler

import (
	"mime/multipart"
	"testing"
)

func TestValidateCarImageUploadAllowedExtensions(t *testing.T) {
	tests := []string{
		"car.jpg",
		"car.jpeg",
		"car.png",
		"car.webp",
		"CAR.JPG",
	}

	for _, filename := range tests {
		t.Run(filename, func(t *testing.T) {
			header := &multipart.FileHeader{
				Filename: filename,
				Size:     maxCarImageUploadSize,
			}

			if err := validateCarImageUpload(header); err != nil {
				t.Fatalf("validateCarImageUpload() error = %v, want nil", err)
			}
		})
	}
}

func TestValidateCarImageUploadRejectsUnsupportedExtension(t *testing.T) {
	header := &multipart.FileHeader{
		Filename: "car.gif",
		Size:     1024,
	}

	if err := validateCarImageUpload(header); err == nil {
		t.Fatal("validateCarImageUpload() error = nil, want unsupported extension error")
	}
}

func TestValidateCarImageUploadRejectsOversizedFile(t *testing.T) {
	header := &multipart.FileHeader{
		Filename: "car.jpg",
		Size:     maxCarImageUploadSize + 1,
	}

	if err := validateCarImageUpload(header); err == nil {
		t.Fatal("validateCarImageUpload() error = nil, want size error")
	}
}

func TestGenerateCarImageFilename(t *testing.T) {
	got := generateCarImageFilename("toyota-corolla", 1719321123123456789, "Original.JPG")
	want := "toyota-corolla-1719321123123456789.jpg"

	if got != want {
		t.Fatalf("generateCarImageFilename() = %q, want %q", got, want)
	}
}

func TestGenerateCarImageFilenameFallback(t *testing.T) {
	got := generateCarImageFilename("", 1719321123123456789, "car.png")
	want := "car-image-1719321123123456789.png"

	if got != want {
		t.Fatalf("generateCarImageFilename() = %q, want %q", got, want)
	}
}
