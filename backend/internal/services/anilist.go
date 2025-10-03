package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// AniListClient - AniList GraphQL API クライアント
type AniListClient struct {
	httpClient *http.Client
	endpoint   string
}

// NewAniListClient - AniListクライアントを作成
func NewAniListClient() *AniListClient {
	return &AniListClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		endpoint: "https://graphql.anilist.co",
	}
}

// CoverImage - AniListのカバー画像情報
type CoverImage struct {
	ExtraLarge string `json:"extraLarge"`
	Large      string `json:"large"`
	Medium     string `json:"medium"`
}

// GetCoverImageByMALId - MyAnimeList IDからカバー画像を取得
func (c *AniListClient) GetCoverImageByMALId(malId string) (*CoverImage, error) {
	// MAL IDを整数に変換
	malIdInt, err := strconv.Atoi(malId)
	if err != nil {
		return nil, fmt.Errorf("invalid MAL ID: %w", err)
	}

	// GraphQLクエリ
	query := `
	query ($malId: Int) {
		Media(idMal: $malId, type: ANIME) {
			coverImage {
				extraLarge
				large
				medium
			}
		}
	}`

	// リクエストボディ作成
	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"malId": malIdInt,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// HTTPリクエスト作成
	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// リクエスト送信
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// レスポンス解析
	var result struct {
		Data struct {
			Media struct {
				CoverImage CoverImage `json:"coverImage"`
			} `json:"Media"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// GraphQLエラーチェック
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	// データが空の場合
	if result.Data.Media.CoverImage.Large == "" {
		return nil, fmt.Errorf("no cover image found for MAL ID: %s", malId)
	}

	return &result.Data.Media.CoverImage, nil
}

// GetCoverImageByTitle - タイトルでカバー画像を検索（フォールバック用）
func (c *AniListClient) GetCoverImageByTitle(title string) (*CoverImage, error) {
	// GraphQLクエリ
	query := `
	query ($search: String) {
		Media(search: $search, type: ANIME) {
			coverImage {
				extraLarge
				large
				medium
			}
		}
	}`

	// リクエストボディ作成
	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"search": title,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// HTTPリクエスト作成
	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// リクエスト送信
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// レスポンス解析
	var result struct {
		Data struct {
			Media struct {
				CoverImage CoverImage `json:"coverImage"`
			} `json:"Media"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// GraphQLエラーチェック
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	// データが空の場合
	if result.Data.Media.CoverImage.Large == "" {
		return nil, fmt.Errorf("no cover image found for title: %s", title)
	}

	return &result.Data.Media.CoverImage, nil
}
