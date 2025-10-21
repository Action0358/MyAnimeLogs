package services

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

	"myanimelogs/internal/common"
)

type UploadService struct {
	uploadDir string
	maxSize   int64
	logger    *common.Logger
}

func NewUploadService(uploadDir string, maxSize int64, logger *common.Logger) *UploadService {
	// アップロードディレクトリが存在しない場合は作成
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("Failed to create upload directory",
			"error", err.Error(),
			"dir", uploadDir)
	}

	return &UploadService{
		uploadDir: uploadDir,
		maxSize:   maxSize,
		logger:    logger,
	}
}

// SaveUserIcon saves user icon and returns URL
func (s *UploadService) SaveUserIcon(userID int, file *multipart.FileHeader) (string, error) {
	// 1. バリデーション
	if err := s.validateImage(file); err != nil {
		return "", err
	}

	// 2. ファイル名生成（重複回避）
	filename := s.generateFilename(userID, file.Filename)

	// 3. ファイル保存
	filepath := filepath.Join(s.uploadDir, filename)
	if err := s.saveFile(file, filepath); err != nil {
		return "", err
	}

	// 4. URL生成
	url := s.generateURL(filename)

	s.logger.Info("User icon uploaded successfully",
		"user_id", userID,
		"filename", filename,
		"size", file.Size)

	return url, nil
}

// validateImage validates the uploaded image file
func (s *UploadService) validateImage(file *multipart.FileHeader) error {
	// サイズチェック
	if file.Size > s.maxSize {
		return fmt.Errorf("ファイルサイズが大きすぎます（最大: %dMB）", s.maxSize/(1024*1024))
	}

	if file.Size == 0 {
		return errors.New("ファイルが空です")
	}

	// MIME typeチェック
	contentType := file.Header.Get("Content-Type")
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	if !contains(allowedTypes, contentType) {
		return errors.New("サポートされていない画像形式です（JPEG、PNG、GIF、WebPのみ）")
	}

	// 拡張子チェック
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	if !contains(allowedExts, ext) {
		return errors.New("サポートされていない画像形式です")
	}

	// ファイル内容の検証（マジックナンバー）
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("ファイルを読み込めませんでした: %w", err)
	}

	detectedType := http.DetectContentType(buffer[:n])
	if !strings.HasPrefix(detectedType, "image/") {
		return errors.New("画像ファイルではありません")
	}

	return nil
}

// generateFilename generates a unique filename for the user icon
func (s *UploadService) generateFilename(userID int, originalFilename string) string {
	ext := strings.ToLower(filepath.Ext(originalFilename))
	timestamp := time.Now().Unix()
	return fmt.Sprintf("user_%d_%d%s", userID, timestamp, ext)
}

// saveFile saves the uploaded file to disk
func (s *UploadService) saveFile(file *multipart.FileHeader, filepath string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("ファイルを作成できませんでした: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("ファイルの保存に失敗しました: %w", err)
	}

	return nil
}

// generateURL generates the URL for the uploaded file
func (s *UploadService) generateURL(filename string) string {
	return fmt.Sprintf("/uploads/icons/%s", filename)
}

// DeleteUserIcon deletes the user icon file
func (s *UploadService) DeleteUserIcon(iconURL string) error {
	if iconURL == "" {
		return nil
	}

	// URLからファイル名を抽出
	filename := filepath.Base(iconURL)

	// ファイルパスを構築
	filepath := filepath.Join(s.uploadDir, filename)

	// ファイルが存在するか確認
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		s.logger.Warn("Icon file not found for deletion",
			"filepath", filepath)
		return nil // ファイルが存在しない場合はエラーにしない
	}

	// ファイルを削除
	if err := os.Remove(filepath); err != nil {
		s.logger.Error("Failed to delete icon file",
			"error", err.Error(),
			"filepath", filepath)
		return fmt.Errorf("ファイルの削除に失敗しました: %w", err)
	}

	s.logger.Info("Icon file deleted successfully",
		"filepath", filepath)

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
