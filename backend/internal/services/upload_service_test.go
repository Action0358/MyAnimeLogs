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
	service := NewUploadService("./test_uploads", "/uploads", 5*1024*1024, logger)

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
	service := NewUploadService("./test_uploads", "/uploads", 5*1024*1024, logger)

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
	service := NewUploadService("./test_uploads/icons", "/uploads", 5*1024*1024, logger)

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
	service := NewUploadService(uploadDir, "/uploads", 5*1024*1024, logger)

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

// Test path traversal protection
func TestUploadService_DeleteUserIcon_PathTraversal(t *testing.T) {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	uploadDir := "./test_uploads_security"
	service := NewUploadService(uploadDir, "/uploads", 5*1024*1024, logger)

	// Create test directory
	os.MkdirAll(uploadDir, 0755)
	defer os.RemoveAll(uploadDir)

	tests := []struct {
		name        string
		iconURL     string
		createFile  bool
		wantErr     bool
		description string
	}{
		{
			name:        "Path traversal with .. (empty after sanitization)",
			iconURL:     "/uploads/icons/..",
			createFile:  false,
			wantErr:     true,
			description: "Should reject empty filename after sanitization",
		},
		{
			name:        "Valid filename (file not exists)",
			iconURL:     "/uploads/icons/user_123_1234567890.jpg",
			createFile:  false,
			wantErr:     false,
			description: "Should not error when file doesn't exist (idempotent)",
		},
		{
			name:        "Valid filename (file exists)",
			iconURL:     "/uploads/icons/user_456_1234567890.jpg",
			createFile:  true,
			wantErr:     false,
			description: "Should successfully delete existing file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create file if needed
			if tt.createFile {
				filename := filepath.Base(tt.iconURL)
				filepath := filepath.Join(uploadDir, filename)
				os.WriteFile(filepath, []byte("test"), 0644)
			}

			err := service.DeleteUserIcon(tt.iconURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteUserIcon() error = %v, wantErr %v (%s)", err, tt.wantErr, tt.description)
			}
		})
	}
}

// Test sanitization prevents path traversal
func TestUploadService_DeleteUserIcon_SanitizationPreventsTraversal(t *testing.T) {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	uploadDir := "./test_uploads_sanitization"
	service := NewUploadService(uploadDir, "/uploads", 5*1024*1024, logger)

	// Create test directory and parent directory file
	os.MkdirAll(uploadDir, 0755)
	defer os.RemoveAll(uploadDir)

	// Create a file in parent directory that should NOT be deleted
	parentFile := filepath.Join(filepath.Dir(uploadDir), "important_file.txt")
	os.WriteFile(parentFile, []byte("important data"), 0644)
	defer os.Remove(parentFile)

	// Try to delete parent file using path traversal
	// After sanitization, this should become "important_file.txt" in uploadDir
	err := service.DeleteUserIcon("/uploads/icons/../important_file.txt")

	// Should not error (file doesn't exist in uploadDir)
	if err != nil {
		t.Errorf("DeleteUserIcon() should not error for non-existent file, got: %v", err)
	}

	// Verify parent file still exists (was not deleted)
	if _, err := os.Stat(parentFile); os.IsNotExist(err) {
		t.Error("Parent directory file was deleted! Path traversal vulnerability exists!")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "Normal filename",
			filename: "test.jpg",
			want:     "test.jpg",
		},
		{
			name:     "Filename with path traversal",
			filename: "../../../etc/passwd",
			want:     "etcpasswd",
		},
		{
			name:     "Filename with backslash",
			filename: "..\\..\\file.txt",
			want:     "file.txt",
		},
		{
			name:     "Filename with NULL byte",
			filename: "file.jpg\x00.txt",
			want:     "file.jpg.txt",
		},
		{
			name:     "Filename with forward slash",
			filename: "path/to/file.jpg",
			want:     "pathtofile.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.filename)
			if got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUploadService_SaveFile_PathTraversal(t *testing.T) {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	uploadDir := "./test_uploads_save_security"
	service := NewUploadService(uploadDir, "/uploads", 5*1024*1024, logger)

	// Create test directory
	os.MkdirAll(uploadDir, 0755)
	defer os.RemoveAll(uploadDir)

	// Create a mock file
	fileHeader := createMockFileHeader("test.jpg", "image/jpeg", 1024)

	tests := []struct {
		name       string
		targetPath string
		wantErr    bool
	}{
		{
			name:       "Valid path",
			targetPath: filepath.Join(uploadDir, "test.jpg"),
			wantErr:    false,
		},
		{
			name:       "Path traversal attempt",
			targetPath: filepath.Join(uploadDir, "../../../etc/passwd"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.saveFile(fileHeader, tt.targetPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("saveFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
