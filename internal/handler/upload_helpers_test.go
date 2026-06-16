package handler

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"
)

func TestValidateCarImageUploadValidJPEG(t *testing.T) {
	header, file := testCarImageUpload("car.jpg", jpegBytes())

	if err := validateCarImageUpload(header, file); err != nil {
		t.Fatalf("validateCarImageUpload() error = %v, want nil", err)
	}
}

func TestValidateCarImageUploadValidPNG(t *testing.T) {
	header, file := testCarImageUpload("car.png", pngBytes())

	if err := validateCarImageUpload(header, file); err != nil {
		t.Fatalf("validateCarImageUpload() error = %v, want nil", err)
	}
}

func TestValidateCarImageUploadValidWebP(t *testing.T) {
	header, file := testCarImageUpload("car.webp", webpBytes())

	if err := validateCarImageUpload(header, file); err != nil {
		t.Fatalf("validateCarImageUpload() error = %v, want nil", err)
	}
}

func TestValidateCarImageUploadRejectsUnsupportedExtensions(t *testing.T) {
	tests := []string{
		"car.gif",
		"car.svg",
		"car.php",
		"car.exe",
		"car",
		"car.jpg.php",
	}

	for _, filename := range tests {
		t.Run(filename, func(t *testing.T) {
			header, file := testCarImageUpload(filename, jpegBytes())

			if err := validateCarImageUpload(header, file); err == nil {
				t.Fatal("validateCarImageUpload() error = nil, want unsupported file error")
			}
		})
	}
}

func TestValidateCarImageUploadRejectsMismatchedExtensionAndContent(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  []byte
	}{
		{
			name:     "png extension with jpeg content",
			filename: "car.png",
			content:  jpegBytes(),
		},
		{
			name:     "jpg extension with png content",
			filename: "car.jpg",
			content:  pngBytes(),
		},
		{
			name:     "webp extension with png content",
			filename: "car.webp",
			content:  pngBytes(),
		},
		{
			name:     "jpg extension with html content",
			filename: "car.jpg",
			content:  []byte("<!doctype html><script>alert(1)</script>"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, file := testCarImageUpload(tt.filename, tt.content)

			if err := validateCarImageUpload(header, file); err == nil {
				t.Fatal("validateCarImageUpload() error = nil, want content mismatch error")
			}
		})
	}
}

func TestValidateCarImageUploadRejectsZeroByteFile(t *testing.T) {
	header, file := testCarImageUpload("car.jpg", nil)

	if err := validateCarImageUpload(header, file); err == nil {
		t.Fatal("validateCarImageUpload() error = nil, want empty file error")
	}
}

func TestValidateCarImageUploadRejectsOversizedFile(t *testing.T) {
	header, file := testCarImageUpload("car.jpg", jpegBytes())
	header.Size = maxCarImageUploadSize + 1

	err := validateCarImageUpload(header, file)
	if err == nil {
		t.Fatal("validateCarImageUpload() error = nil, want size error")
	}
	if err.Error() != "Image file is too large. Upload a JPEG, PNG, or WebP image up to 5 MB." {
		t.Fatalf("validateCarImageUpload() error = %q", err.Error())
	}
}

func TestValidateCarImageUploadAllowsNilHeader(t *testing.T) {
	if err := validateCarImageUpload(nil, nil); err != nil {
		t.Fatalf("validateCarImageUpload() error = %v, want nil", err)
	}
}

func TestValidateCarImageUploadResetsFilePosition(t *testing.T) {
	content := jpegBytes()
	header, file := testCarImageUpload("car.jpg", content)

	if err := validateCarImageUpload(header, file); err != nil {
		t.Fatalf("validateCarImageUpload() error = %v, want nil", err)
	}

	remaining, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if !bytes.Equal(remaining, content) {
		t.Fatal("validateCarImageUpload() did not reset file position")
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

func TestGenerateCarImageFilenameIgnoresDangerousOriginalPath(t *testing.T) {
	got := generateCarImageFilename("Toyota Corolla", 1719321123123456789, "../../evil.php.jpg")
	want := "toyota-corolla-1719321123123456789.jpg"

	if got != want {
		t.Fatalf("generateCarImageFilename() = %q, want %q", got, want)
	}

	if strings.Contains(got, "/") || strings.Contains(got, "\\") || strings.Contains(got, "evil") || strings.Contains(got, "php") {
		t.Fatalf("generateCarImageFilename() = %q contains unsafe original filename content", got)
	}
}

func TestCarImageDestinationPathRejectsTraversal(t *testing.T) {
	if _, err := carImageDestinationPath("../car.jpg"); err == nil {
		t.Fatal("carImageDestinationPath() error = nil, want traversal error")
	}
}

type testMultipartFile struct {
	*bytes.Reader
}

func (f testMultipartFile) Close() error {
	return nil
}

func testCarImageUpload(filename string, content []byte) (*multipart.FileHeader, multipart.File) {
	return &multipart.FileHeader{
		Filename: filename,
		Size:     int64(len(content)),
	}, testMultipartFile{Reader: bytes.NewReader(content)}
}

func jpegBytes() []byte {
	return []byte{
		0xff, 0xd8, 0xff, 0xe0,
		0x00, 0x10,
		'J', 'F', 'I', 'F', 0x00,
		0x01, 0x01, 0x01,
		0x00, 0x48, 0x00, 0x48,
		0x00, 0x00,
		0xff, 0xd9,
	}
}

func pngBytes() []byte {
	return []byte{
		0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d,
		'I', 'H', 'D', 'R',
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00,
	}
}

func webpBytes() []byte {
	return []byte{
		'R', 'I', 'F', 'F',
		0x10, 0x00, 0x00, 0x00,
		'W', 'E', 'B', 'P',
		'V', 'P', '8', ' ',
		0x04, 0x00, 0x00, 0x00,
	}
}
