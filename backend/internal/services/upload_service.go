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
	uploadDir  string
	publicPath string // Public URL path (e.g., "/uploads")
	maxSize    int64
	logger     *common.Logger
}

func NewUploadService(uploadDir string, publicPath string, maxSize int64, logger *common.Logger) *UploadService {
	// アップロードディレクトリが存在しない場合は作成
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("Failed to create upload directory",
			"error", err.Error(),
			"dir", uploadDir)
	}

	return &UploadService{
		uploadDir:  uploadDir,
		publicPath: publicPath,
		maxSize:    maxSize,
		logger:     logger,
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
	// ファイル名をサニタイズしてから拡張子を取得
	sanitized := sanitizeFilename(originalFilename)
	ext := strings.ToLower(filepath.Ext(sanitized))

	// 拡張子が不正な場合はデフォルトを使用
	if ext == "" || len(ext) > 10 {
		ext = ".jpg"
	}

	timestamp := time.Now().Unix()
	return fmt.Sprintf("user_%d_%d%s", userID, timestamp, ext)
}

// saveFile saves the uploaded file to disk
func (s *UploadService) saveFile(file *multipart.FileHeader, targetPath string) error {
	// パストラバーサル対策: 最終パスがアップロードディレクトリ内にあることを確認
	absUploadDir, err := filepath.Abs(s.uploadDir)
	if err != nil {
		return fmt.Errorf("ディレクトリパスの取得に失敗しました: %w", err)
	}

	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("ファイルパスの取得に失敗しました: %w", err)
	}

	// 最終パスがアップロードディレクトリ内にあることを確認
	if !strings.HasPrefix(absTargetPath, absUploadDir+string(filepath.Separator)) &&
		absTargetPath != absUploadDir {
		s.logger.Warn("Path traversal attempt detected in saveFile",
			"upload_dir", absUploadDir,
			"target_path", absTargetPath)
		return errors.New("不正なファイルパスです")
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(absTargetPath)
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
	// publicPathから相対的なサブディレクトリを計算
	// uploadDir: "./uploads/icons" -> subdir: "icons"
	// publicPath: "/uploads" -> URL: "/uploads/icons/filename"
	subdir := filepath.Base(s.uploadDir)
	return fmt.Sprintf("%s/%s/%s", s.publicPath, subdir, filename)
}

// DeleteUserIcon deletes the user icon file
func (s *UploadService) DeleteUserIcon(iconURL string) error {
	if iconURL == "" {
		return nil
	}

	// URLからファイル名を抽出（クエリパラメータを除去）
	filename, err := s.extractFilenameFromURL(iconURL)
	if err != nil {
		s.logger.Warn("Failed to extract filename from URL",
			"url", iconURL,
			"error", err.Error())
		return err
	}

	// ファイル名のサニタイズ（パストラバーサル対策）
	filename = sanitizeFilename(filename)
	if filename == "" || filename == "." || filename == ".." {
		s.logger.Warn("Invalid filename detected",
			"original_url", iconURL,
			"sanitized_filename", filename)
		return errors.New("無効なファイル名です")
	}

	// 追加のバリデーション: ファイル名に不正な文字が含まれていないか確認
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		s.logger.Warn("Invalid characters in filename after sanitization",
			"filename", filename,
			"original_url", iconURL)
		return errors.New("無効なファイル名です")
	}

	// ファイルパスを構築
	targetPath := filepath.Join(s.uploadDir, filename)

	// パストラバーサル対策: 最終パスがアップロードディレクトリ内にあることを確認
	absUploadDir, err := filepath.Abs(s.uploadDir)
	if err != nil {
		s.logger.Error("Failed to get absolute upload directory path",
			"error", err.Error())
		return fmt.Errorf("ディレクトリパスの取得に失敗しました: %w", err)
	}

	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		s.logger.Error("Failed to get absolute target path",
			"error", err.Error())
		return fmt.Errorf("ファイルパスの取得に失敗しました: %w", err)
	}

	// 最終パスがアップロードディレクトリ内にあることを確認
	if !strings.HasPrefix(absTargetPath, absUploadDir+string(filepath.Separator)) &&
		absTargetPath != absUploadDir {
		s.logger.Warn("Path traversal attempt detected",
			"upload_dir", absUploadDir,
			"target_path", absTargetPath,
			"original_url", iconURL)
		return errors.New("不正なファイルパスです")
	}

	// ファイルが存在するか確認
	if _, err := os.Stat(absTargetPath); os.IsNotExist(err) {
		s.logger.Warn("Icon file not found for deletion",
			"filepath", absTargetPath)
		return nil // ファイルが存在しない場合はエラーにしない
	}

	// ファイルを削除
	if err := os.Remove(absTargetPath); err != nil {
		s.logger.Error("Failed to delete icon file",
			"error", err.Error(),
			"filepath", absTargetPath)
		return fmt.Errorf("ファイルの削除に失敗しました: %w", err)
	}

	s.logger.Info("Icon file deleted successfully",
		"filepath", absTargetPath)

	return nil
}

// extractFilenameFromURL extracts filename from URL, handling query parameters
func (s *UploadService) extractFilenameFromURL(iconURL string) (string, error) {
	// URLパスからファイル名を抽出（クエリパラメータを除去）
	// 例: "/uploads/icons/user_1_123.jpg?v=456" -> "user_1_123.jpg"

	// クエリパラメータを含む可能性があるため、'?'で分割
	urlWithoutQuery := strings.Split(iconURL, "?")[0]

	// パスからファイル名を取得
	filename := filepath.Base(urlWithoutQuery)

	if filename == "" || filename == "." || filename == "/" {
		return "", errors.New("無効なURLです")
	}

	return filename, nil
}

// sanitizeFilename removes dangerous characters from filename
func sanitizeFilename(filename string) string {
	// パストラバーサル対策: 危険な文字を削除
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "\x00", "") // NULL byte
	return filename
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
