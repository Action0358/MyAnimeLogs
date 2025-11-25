package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	// DefaultAniListEndpoint - AniList GraphQL APIのデフォルトエンドポイント
	DefaultAniListEndpoint = "https://graphql.anilist.co"

	// DefaultHTTPClientTimeout - HTTPクライアントのデフォルトタイムアウト
	DefaultHTTPClientTimeout = 10 * time.Second
)

// AniListClient - AniList GraphQL API クライアント
type AniListClient struct {
	httpClient *http.Client
	endpoint   string
}

// genreTranslationMap - 英語から日本語へのジャンル変換マップ
var genreTranslationMap = map[string]string{
	// 基本ジャンル
	"Action":        "アクション",
	"Adventure":     "アドベンチャー",
	"Comedy":        "コメディ",
	"Drama":         "ドラマ",
	"Ecchi":         "エッチ",
	"Fantasy":       "ファンタジー",
	"Horror":        "ホラー",
	"Mahou Shoujo":  "魔法少女",
	"Mecha":         "メカ",
	"Music":         "音楽",
	"Mystery":       "ミステリー",
	"Psychological": "心理",
	"Romance":       "ロマンス",
	"Sci-Fi":        "SF",
	"Slice of Life": "日常",
	"Sports":        "スポーツ",
	"Supernatural":  "超常現象",
	"Thriller":      "スリラー",

	// 追加ジャンル
	"Historical": "歴史",
	"Military":   "ミリタリー",
	"Parody":     "パロディ",
	"School":     "学園",
	"Shoujo":     "少女",
	"Shounen":    "少年",
	"Seinen":     "青年",
	"Josei":      "女性",
}

// NewAniListClient - AniListクライアントを作成
// 環境変数 ANILIST_ENDPOINT が設定されている場合はそれを使用、なければデフォルトを使用
func NewAniListClient() *AniListClient {
	endpoint := os.Getenv("ANILIST_ENDPOINT")
	if endpoint == "" {
		endpoint = DefaultAniListEndpoint
	}

	return NewAniListClientWithEndpoint(endpoint)
}

// NewAniListClientWithEndpoint - カスタムエンドポイントでAniListクライアントを作成
// テスト時にモックサーバーを指定する場合などに使用
func NewAniListClientWithEndpoint(endpoint string) *AniListClient {
	if endpoint == "" {
		endpoint = DefaultAniListEndpoint
	}

	return &AniListClient{
		httpClient: &http.Client{
			Timeout: DefaultHTTPClientTimeout,
		},
		endpoint: endpoint,
	}
}

// CoverImage - AniListのカバー画像情報
type CoverImage struct {
	ExtraLarge string `json:"extraLarge"`
	Large      string `json:"large"`
	Medium     string `json:"medium"`
	MalId      *int   `json:"malId,omitempty"` // キャッシュキーとして使用
}

// GetCoverImageByMALId - MyAnimeList IDからカバー画像を取得
func (c *AniListClient) GetCoverImageByMALId(malId string) (*CoverImage, error) {
	// MAL IDを整数に変換
	malIdInt, err := strconv.Atoi(malId)
	if err != nil {
		return nil, fmt.Errorf("invalid MAL ID: %w", err)
	}

	// GraphQLクエリ（MAL IDも取得）
	query := `
	query ($malId: Int) {
		Media(idMal: $malId, type: ANIME) {
			idMal
			coverImage {
				extraLarge
				large
				medium
			}
		}
	}`

	variables := map[string]interface{}{
		"malId": malIdInt,
	}

	coverImage, err := c.executeGraphQLQuery(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get cover image by MAL ID %s: %w", malId, err)
	}

	return coverImage, nil
}

// GetCoverImageByTitle - タイトルでカバー画像を検索（フォールバック用）
func (c *AniListClient) GetCoverImageByTitle(title string) (*CoverImage, error) {
	// GraphQLクエリ（MAL IDも取得してキャッシュキーとして使用）
	query := `
	query ($search: String) {
		Media(search: $search, type: ANIME) {
			idMal
			coverImage {
				extraLarge
				large
				medium
			}
		}
	}`

	variables := map[string]interface{}{
		"search": title,
	}

	coverImage, err := c.executeGraphQLQuery(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get cover image by title %s: %w", title, err)
	}

	return coverImage, nil
}

// executeGraphQLQuery - GraphQLクエリを送信し、レスポンスを解析する共通ヘルパー
func (c *AniListClient) executeGraphQLQuery(query string, variables map[string]interface{}) (*CoverImage, error) {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// リクエストボディ作成
		requestBody := map[string]interface{}{
			"query":     query,
			"variables": variables,
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
			if attempt < maxRetries {
				// 指数バックオフ: 2秒, 4秒, 8秒...
				time.Sleep(retryDelay * time.Duration(1<<(attempt-1)))
				continue
			}
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		// レート制限の場合はリトライ
		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt < maxRetries {
				// 指数バックオフ: 2秒, 4秒, 8秒...
				time.Sleep(retryDelay * time.Duration(attempt-1))
				continue
			}
			return nil, fmt.Errorf("rate limited after %d attempts", maxRetries)
		}

		// ステータスコード確認
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// レスポンス解析
		var result struct {
			Data struct {
				Media struct {
					IdMal      *int       `json:"idMal"`
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
			return nil, fmt.Errorf("no cover image found")
		}

		// MAL IDをCoverImageに設定
		coverImage := result.Data.Media.CoverImage
		coverImage.MalId = result.Data.Media.IdMal

		return &coverImage, nil
	}

	return nil, fmt.Errorf("failed after %d attempts", maxRetries)
}

// GetGenresByMALId - MyAnimeList IDからジャンル情報を取得
func (c *AniListClient) GetGenresByMALId(malId string) ([]string, error) {
	// MAL IDを整数に変換
	malIdInt, err := strconv.Atoi(malId)
	if err != nil {
		return nil, fmt.Errorf("invalid MAL ID: %w", err)
	}

	// GraphQLクエリ
	query := `
	query ($malId: Int) {
		Media(idMal: $malId, type: ANIME) {
			genres
		}
	}`

	variables := map[string]interface{}{
		"malId": malIdInt,
	}

	genres, err := c.executeGenresQuery(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres by MAL ID %s: %w", malId, err)
	}

	return genres, nil
}

// executeGenresQuery - ジャンル取得用のGraphQLクエリを実行
func (c *AniListClient) executeGenresQuery(query string, variables map[string]interface{}) ([]string, error) {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// リクエストボディ作成
		requestBody := map[string]interface{}{
			"query":     query,
			"variables": variables,
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
			if attempt < maxRetries {
				time.Sleep(retryDelay * time.Duration(1<<(attempt-1)))
				continue
			}
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		// レート制限の場合はリトライ
		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt < maxRetries {
				time.Sleep(retryDelay * time.Duration(attempt))
				continue
			}
			return nil, fmt.Errorf("rate limited after %d attempts", maxRetries)
		}

		// ステータスコード確認
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// レスポンス解析
		var result struct {
			Data struct {
				Media struct {
					Genres []string `json:"genres"`
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

		// ジャンルを日本語に変換
		var translatedGenres []string
		for _, genre := range result.Data.Media.Genres {
			translatedGenre := translateGenre(genre)
			translatedGenres = append(translatedGenres, translatedGenre)
		}

		return translatedGenres, nil
	}

	return nil, fmt.Errorf("failed after %d attempts", maxRetries)
}

// translateGenre - ジャンル名を英語から日本語に変換
func translateGenre(genreEN string) string {
	if jp, ok := genreTranslationMap[genreEN]; ok {
		return jp
	}
	// マッピングがない場合は英語のまま返す
	return genreEN
}
