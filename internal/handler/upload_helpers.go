package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxCarImageUploadSize = 5 << 20
const carImageUploadDir = "web/static/uploads/cars"
const carImagePublicPath = "/static/uploads/cars"

var allowedCarImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
}

func validateCarImageUpload(header *multipart.FileHeader) error {
	if header == nil {
		return nil
	}

	if header.Size > maxCarImageUploadSize {
		return fmt.Errorf("car image upload exceeds 5 MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if _, ok := allowedCarImageExtensions[ext]; !ok {
		return fmt.Errorf("unsupported car image extension")
	}

	return nil
}

func saveCarImageUpload(file multipart.File, header *multipart.FileHeader, carSlug string) (string, error) {
	defer file.Close()

	if err := validateCarImageUpload(header); err != nil {
		return "", err
	}

	if err := os.MkdirAll(carImageUploadDir, 0o755); err != nil {
		return "", fmt.Errorf("create car image upload directory: %w", err)
	}

	filename := generateCarImageFilename(carSlug, time.Now().UnixNano(), header.Filename)
	destinationPath := filepath.Join(carImageUploadDir, filename)

	destination, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("create car image upload file: %w", err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, file); err != nil {
		return "", fmt.Errorf("save car image upload: %w", err)
	}

	return carImagePublicPath + "/" + filename, nil
}

func generateCarImageFilename(carSlug string, timestamp int64, originalFilename string) string {
	base := sanitizeCarImageFilenameBase(carSlug)
	if base == "" {
		base = "car-image"
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	return fmt.Sprintf("%s-%d%s", base, timestamp, ext)
}

func sanitizeCarImageFilenameBase(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))

	var builder strings.Builder
	previousHyphen := false
	for _, char := range value {
		isAllowed := (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
		if isAllowed {
			builder.WriteRune(char)
			previousHyphen = false
			continue
		}

		if char == '-' && !previousHyphen {
			builder.WriteRune(char)
			previousHyphen = true
		}
	}

	return strings.Trim(builder.String(), "-")
}
