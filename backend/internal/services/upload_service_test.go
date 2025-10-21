package services

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"myanimelogs/internal/common"
)

func TestUploadService_validateImage(t *testing.T) {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	service := NewUploadService("./test_uploads", 5*1024*1024, logger)

	tests := []struct {
		name        string
		fileSize    int64
		contentType string
		filename    string
		wantErr     bool
	}{
		{
			name:        "Valid JPEG image",
			fileSize:    1024,
			contentType: "image/jpeg",
			filename:    "test.jpg",
			wantErr:     false,
		},
		{
			name:        "Valid PNG image",
			fileSize:    2048,
			contentType: "image/png",
			filename:    "test.png",
			wantErr:     false,
		},
		{
			name:        "File too large",
			fileSize:    10 * 1024 * 1024,
			contentType: "image/jpeg",
			filename:    "test.jpg",
			wantErr:     true,
		},
		{
			name:        "Invalid content type",
			fileSize:    1024,
			contentType: "application/pdf",
			filename:    "test.pdf",
			wantErr:     true,
		},
		{
			name:        "Invalid extension",
			fileSize:    1024,
			contentType: "image/jpeg",
			filename:    "test.txt",
			wantErr:     true,
		},
		{
			name:        "Empty file",
			fileSize:    0,
			contentType: "image/jpeg",
			filename:    "test.jpg",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock file header
			fileHeader := createMockFileHeader(tt.filename, tt.contentType, tt.fileSize)

			err := service.validateImage(fileHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadService_generateFilename(t *testing.T) {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	service := NewUploadService("./test_uploads", 5*1024*1024, logger)

	tests := []struct {
		name             string
		userID           int
		originalFilename string
		wantPrefix       string
		wantExt          string
	}{
		{
			name:             "JPEG file",
			userID:           123,
			originalFilename: "profile.jpg",
			wantPrefix:       "user_123_",
			wantExt:          ".jpg",
		},
		{
			name:             "PNG file",
			userID:           456,
			originalFilename: "avatar.PNG",
			wantPrefix:       "user_456_",
			wantExt:          ".png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := service.generateFilename(tt.userID, tt.originalFilename)

			if len(filename) == 0 {
				t.Error("generateFilename() returned empty string")
			}

			// Check prefix
			if len(filename) < len(tt.wantPrefix) || filename[:len(tt.wantPrefix)] != tt.wantPrefix {
				t.Errorf("generateFilename() prefix = %v, want %v", filename[:len(tt.wantPrefix)], tt.wantPrefix)
			}

			// Check extension
			ext := filepath.Ext(filename)
			if ext != tt.wantExt {
				t.Errorf("generateFilename() extension = %v, want %v", ext, tt.wantExt)
			}
		})
	}
}

func TestUploadService_generateURL(t *testing.T) {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	service := NewUploadService("./test_uploads", 5*1024*1024, logger)

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "Generate URL",
			filename: "user_123_1234567890.jpg",
			want:     "/uploads/icons/user_123_1234567890.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.generateURL(tt.filename)
			if got != tt.want {
				t.Errorf("generateURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUploadService_DeleteUserIcon(t *testing.T) {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	uploadDir := "./test_uploads_delete"
	service := NewUploadService(uploadDir, 5*1024*1024, logger)

	// Create test directory
	os.MkdirAll(uploadDir, 0755)
	defer os.RemoveAll(uploadDir)

	tests := []struct {
		name       string
		iconURL    string
		createFile bool
		wantErr    bool
	}{
		{
			name:       "Delete existing file",
			iconURL:    "/uploads/icons/test_file.jpg",
			createFile: true,
			wantErr:    false,
		},
		{
			name:       "Delete non-existing file",
			iconURL:    "/uploads/icons/non_existing.jpg",
			createFile: false,
			wantErr:    false, // Should not error
		},
		{
			name:       "Empty URL",
			iconURL:    "",
			createFile: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createFile && tt.iconURL != "" {
				filename := filepath.Base(tt.iconURL)
				filepath := filepath.Join(uploadDir, filename)
				os.WriteFile(filepath, []byte("test"), 0644)
			}

			err := service.DeleteUserIcon(tt.iconURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteUserIcon() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to create a mock file header
func createMockFileHeader(filename, contentType string, size int64) *multipart.FileHeader {
	// Create a buffer with mock image data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add proper magic numbers for valid images
	var content []byte
	if contentType == "image/jpeg" {
		content = []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG magic number
	} else if contentType == "image/png" {
		// PNG magic number (8 bytes)
		content = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	} else if contentType == "image/gif" {
		content = []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61} // GIF89a
	} else if contentType == "image/webp" {
		// WebP magic number
		content = []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50}
	} else {
		content = []byte("not an image")
	}

	// Pad to desired size
	if int64(len(content)) < size {
		padding := make([]byte, size-int64(len(content)))
		content = append(content, padding...)
	}

	part, _ := writer.CreateFormFile("file", filename)
	io.Copy(part, bytes.NewReader(content))
	writer.Close()

	// Parse the multipart form
	reader := multipart.NewReader(&buf, writer.Boundary())
	form, _ := reader.ReadForm(size + 1024)

	fileHeader := form.File["file"][0]
	fileHeader.Header.Set("Content-Type", contentType)
	fileHeader.Size = size

	return fileHeader
}
