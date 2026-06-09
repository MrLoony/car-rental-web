package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxCarImageUploadSize = 5 << 20
const carImageUploadDir = "web/static/uploads/cars"
const carImagePublicPath = "/static/uploads/cars"
const carImageSniffBytes = 512

var allowedCarImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
}

func validateCarImageUpload(header *multipart.FileHeader, file multipart.File) error {
	if header == nil {
		return nil
	}

	if header.Size == 0 {
		return errors.New("Uploaded image cannot be empty.")
	}

	if header.Size > maxCarImageUploadSize {
		return errors.New("Uploaded image must be 5 MB or smaller.")
	}

	ext := carImageExtension(header.Filename)
	if _, ok := allowedCarImageExtensions[ext]; !ok {
		return errors.New("Uploaded image must be a JPEG, PNG, or WebP file.")
	}

	if file == nil {
		return errors.New("Uploaded image must be a JPEG, PNG, or WebP file.")
	}

	contentType, err := sniffCarImageContentType(file)
	if err != nil {
		return err
	}

	if !carImageExtensionMatchesContentType(ext, contentType) {
		return errors.New("Uploaded image must be a JPEG, PNG, or WebP file.")
	}

	return nil
}

func saveCarImageUpload(file multipart.File, header *multipart.FileHeader, carSlug string) (string, error) {
	defer file.Close()

	if err := validateCarImageUpload(header, file); err != nil {
		return "", err
	}

	if err := os.MkdirAll(carImageUploadDir, 0o755); err != nil {
		return "", fmt.Errorf("create car image upload directory: %w", err)
	}

	filename := generateCarImageFilename(carSlug, time.Now().UnixNano(), header.Filename)
	destinationPath, err := carImageDestinationPath(filename)
	if err != nil {
		return "", err
	}

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

	ext := carImageExtension(originalFilename)
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

		if (char == '-' || char == ' ' || char == '_') && !previousHyphen {
			builder.WriteByte('-')
			previousHyphen = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func sniffCarImageContentType(file multipart.File) (string, error) {
	buffer := make([]byte, carImageSniffBytes)
	bytesRead, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read uploaded image: %w", err)
	}

	if bytesRead == 0 {
		return "", errors.New("Uploaded image cannot be empty.")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("reset uploaded image: %w", err)
	}

	sniffedBytes := buffer[:bytesRead]
	if isWebP(sniffedBytes) {
		return "image/webp", nil
	}

	return http.DetectContentType(sniffedBytes), nil
}

func carImageExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

func carImageExtensionMatchesContentType(ext string, contentType string) bool {
	switch ext {
	case ".jpg", ".jpeg":
		return contentType == "image/jpeg"
	case ".png":
		return contentType == "image/png"
	case ".webp":
		return contentType == "image/webp"
	default:
		return false
	}
}

func isWebP(data []byte) bool {
	return len(data) >= 12 &&
		string(data[0:4]) == "RIFF" &&
		string(data[8:12]) == "WEBP"
}

func carImageDestinationPath(filename string) (string, error) {
	destinationPath := filepath.Join(carImageUploadDir, filename)

	uploadDirAbs, err := filepath.Abs(carImageUploadDir)
	if err != nil {
		return "", fmt.Errorf("resolve car image upload directory: %w", err)
	}

	destinationAbs, err := filepath.Abs(destinationPath)
	if err != nil {
		return "", fmt.Errorf("resolve car image upload path: %w", err)
	}

	relativePath, err := filepath.Rel(uploadDirAbs, destinationAbs)
	if err != nil {
		return "", fmt.Errorf("validate car image upload path: %w", err)
	}

	if relativePath == "." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) || relativePath == ".." || filepath.IsAbs(relativePath) {
		return "", errors.New("invalid car image upload path")
	}

	return destinationPath, nil
}
