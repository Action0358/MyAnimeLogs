package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"myanimelogs/internal/constants"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	// DefaultAniListEndpoint - AniList GraphQL APIのデフォルトエンドポイント
	DefaultAniListEndpoint = "https://graphql.anilist.co"

	// DefaultHTTPClientTimeout - HTTPクライアントのデフォルトタイムアウト
	DefaultHTTPClientTimeout = 10 * time.Second
)

// AniListClient - AniList GraphQL API クライアント
type AniListClient struct {
	httpClient  *http.Client
	endpoint    string
	rateLimiter *rate.Limiter
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

	// レートリミッターを作成
	// AniList APIは90リクエスト/分なので、安全マージンを考慮して制限
	// rate.Every()を使用して、リクエスト間の間隔を計算
	requestsPerMinute := float64(constants.AniListAPIRateLimit) * constants.AniListAPIRateLimitSafetyMargin
	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMinute)), 1)

	return &AniListClient{
		httpClient: &http.Client{
			Timeout: DefaultHTTPClientTimeout,
		},
		endpoint:    endpoint,
		rateLimiter: limiter,
	}
}

// CoverImage - AniListのカバー画像情報
type CoverImage struct {
	ExtraLarge string `json:"extraLarge"`
	Large      string `json:"large"`
	Medium     string `json:"medium"`
	MalId      *int   `json:"malId,omitempty"` // キャッシュキーとして使用
}

// coverImageResponse - カバー画像取得のGraphQLレスポンス構造体
type coverImageResponse struct {
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

// genresResponse - ジャンル取得のGraphQLレスポンス構造体
type genresResponse struct {
	Data struct {
		Media struct {
			Genres []string `json:"genres"`
		} `json:"Media"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// descriptionResponse - あらすじ取得のGraphQLレスポンス構造体
type descriptionResponse struct {
	Data struct {
		Media struct {
			Description string `json:"description"`
		} `json:"Media"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// processCoverImageResponse - カバー画像レスポンスを処理する共通ヘルパー
func processCoverImageResponse(result *coverImageResponse) (*CoverImage, error) {
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

// processGenresResponse - ジャンルレスポンスを処理する共通ヘルパー
func processGenresResponse(result *genresResponse) ([]string, error) {
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

// GetCoverImageByMALId - MyAnimeList IDからカバー画像を取得
func (c *AniListClient) GetCoverImageByMALId(ctx context.Context, malId string) (*CoverImage, error) {
	malIdInt, err := c.parseMALId(malId)
	if err != nil {
		return nil, err
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

	result, err := executeGraphQLQuery[coverImageResponse](c, ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get cover image by MAL ID %s: %w", malId, err)
	}

	return processCoverImageResponse(result)
}

// GetCoverImageByTitle - タイトルでカバー画像を検索（フォールバック用）
func (c *AniListClient) GetCoverImageByTitle(ctx context.Context, title string) (*CoverImage, error) {
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

	result, err := executeGraphQLQuery[coverImageResponse](c, ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get cover image by title %s: %w", title, err)
	}

	return processCoverImageResponse(result)
}

// GetGenresByMALId - MyAnimeList IDからジャンル情報を取得
func (c *AniListClient) GetGenresByMALId(ctx context.Context, malId string) ([]string, error) {
	malIdInt, err := c.parseMALId(malId)
	if err != nil {
		return nil, err
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

	result, err := executeGraphQLQuery[genresResponse](c, ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres by MAL ID %s: %w", malId, err)
	}

	return processGenresResponse(result)
}

// GetDescriptionByMALId - MyAnimeList IDからあらすじを取得（日本語優先）
func (c *AniListClient) GetDescriptionByMALId(ctx context.Context, malId string) (string, error) {
	malIdInt, err := c.parseMALId(malId)
	if err != nil {
		return "", err
	}

	// GraphQLクエリ（日本語のあらすじを取得）
	// AniListはdescriptionをHTMLタグ付きで返すことがある
	query := `
	query ($malId: Int) {
		Media(idMal: $malId, type: ANIME) {
			description(asHtml: false)
		}
	}`

	variables := map[string]interface{}{
		"malId": malIdInt,
	}

	result, err := executeGraphQLQuery[descriptionResponse](c, ctx, query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to get description by MAL ID %s: %w", malId, err)
	}

	// GraphQLエラーチェック
	if len(result.Errors) > 0 {
		return "", fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	return result.Data.Media.Description, nil
}

// mediaFormatResponse - メディアフォーマット取得用のレスポンス構造体
type mediaFormatResponse struct {
	Data struct {
		Media struct {
			Format string `json:"format"`
		} `json:"Media"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// GetMediaFormatByMALId - MAL IDからメディアフォーマット（TV, MOVIE, OVA等）を取得
// 戻り値: "TV", "MOVIE", "OVA", "ONA", "SPECIAL", "MUSIC" など
func (c *AniListClient) GetMediaFormatByMALId(ctx context.Context, malId string) (string, error) {
	malIdInt, err := c.parseMALId(malId)
	if err != nil {
		return "", err
	}

	query := `
	query ($malId: Int) {
		Media(idMal: $malId, type: ANIME) {
			format
		}
	}`

	variables := map[string]interface{}{
		"malId": malIdInt,
	}

	result, err := executeGraphQLQuery[mediaFormatResponse](c, ctx, query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to get media format by MAL ID %s: %w", malId, err)
	}

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	return result.Data.Media.Format, nil
}

// executeGraphQLQuery - GraphQLクエリを送信し、レスポンスを解析する共通ヘルパー（ジェネリクス版）
func executeGraphQLQuery[T any](c *AniListClient, ctx context.Context, query string, variables map[string]interface{}) (*T, error) {
	maxRetries := constants.AniListMaxRetries

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// レートリミッターを使用してリクエストを制御
		// Wait()はトークンが利用可能になるまでブロックする
		// contextを伝播することで、キャンセルシグナルを適切に処理
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		// リクエストボディ作成
		requestBody := map[string]interface{}{
			"query":     query,
			"variables": variables,
		}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		// HTTPリクエスト作成（contextを伝播）
		req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// リクエスト送信
		resp, err := c.httpClient.Do(req)
		if err != nil {
			// ネットワークエラー: 指数バックオフでリトライ
			if attempt < maxRetries {
				if waitErr := sleepWithContext(ctx, retryBackoff(attempt, 0, false)); waitErr != nil {
					return nil, waitErr
				}
				continue
			}
			return nil, fmt.Errorf("failed to send request: %w", err)
		}

		// レート制限(429): サーバーのRetry-Afterを尊重して待機後にリトライ
		// 固定の短いバックオフ(2/4秒)では1分単位のリセット窓を超えられず、
		// リトライも429で失敗していたため、Retry-Afterを優先する
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter, hasRetryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			resp.Body.Close()
			if attempt < maxRetries {
				if waitErr := sleepWithContext(ctx, retryBackoff(attempt, retryAfter, hasRetryAfter)); waitErr != nil {
					return nil, waitErr
				}
				continue
			}
			return nil, fmt.Errorf("rate limited after %d attempts", maxRetries)
		}

		// ステータスコード確認
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// レスポンス解析
		var result T
		decodeErr := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
		}

		return &result, nil
	}

	return nil, fmt.Errorf("failed after %d attempts", maxRetries)
}

// parseRetryAfter parses an HTTP Retry-After header expressed as delay-seconds.
// Returns (0, false) when the header is absent or not a valid non-negative integer.
func parseRetryAfter(header string) (time.Duration, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return 0, false
	}
	secs, err := strconv.Atoi(header)
	if err != nil || secs < 0 {
		return 0, false
	}
	// 防御: 過大な秒数を time.Second と乗算すると time.Duration(int64ナノ秒)が
	// オーバーフローして負値にラップし、上限キャップを回避して即時リトライになる。
	// どのみち retryBackoff で AniListMaxRetryBackoff にキャップされるため、
	// その秒数で事前にクランプして安全に防ぐ。
	maxSecs := int(constants.AniListMaxRetryBackoff / time.Second)
	if secs > maxSecs {
		secs = maxSecs
	}
	return time.Duration(secs) * time.Second, true
}

// retryBackoff computes how long to wait before the next attempt. When retryAfter > 0
// (from the server's Retry-After header) it is honored; otherwise an exponential backoff
// (2s, 4s, 8s, ...) is used. The result is capped at AniListMaxRetryBackoff and gets
// additive-only jitter so concurrent callers desync without ever undershooting a
// server-mandated delay.
func retryBackoff(attempt int, retryAfter time.Duration, hasRetryAfter bool) time.Duration {
	var d time.Duration
	if hasRetryAfter {
		// サーバー指定のRetry-Afterを尊重（0秒の即時リトライ指示も含む）
		d = retryAfter
	} else {
		d = constants.AniListBaseRetryDelay * time.Duration(1<<uint(attempt-1))
	}
	if d > constants.AniListMaxRetryBackoff {
		d = constants.AniListMaxRetryBackoff
	}
	if d > 0 {
		// additive jitter: 0 〜 +10%
		d += time.Duration(rand.Int63n(int64(d)/10 + 1))
	}
	return d
}

// translateGenre - ジャンル名を英語から日本語に変換
func translateGenre(genreEN string) string {
	if jp, ok := genreTranslationMap[genreEN]; ok {
		return jp
	}
	// マッピングがない場合は英語のまま返す
	return genreEN
}

// parseMALId - MAL ID文字列を整数に変換するヘルパー関数
// 複数のメソッドで使用される共通ロジックを集約
func (c *AniListClient) parseMALId(malId string) (int, error) {
	malIdInt, err := strconv.Atoi(malId)
	if err != nil {
		return 0, fmt.Errorf("invalid MAL ID: %w", err)
	}
	return malIdInt, nil
}
