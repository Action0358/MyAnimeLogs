package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

// annictGraphQLEndpoint - シリーズ情報はREST API（/v1/works）に無いため、GraphQL APIから取得する。
// テストでモックサーバーへ差し替えられるよう変数にしている
var annictGraphQLEndpoint = "https://api.annict.com/graphql"

const (
	// annictSeriesBatchSize - 1リクエストで問い合わせる作品数。
	// searchWorksのfirstの上限内に収め、約1万件でも200リクエスト程度で済む大きさにする
	annictSeriesBatchSize = 50

	// annictSeriesRecheckDays - 所属シリーズが空だった作品を再確認するまでの日数。
	// 放送前の新作はAnnict側でシリーズ登録が後から行われることがあるため
	annictSeriesRecheckDays = 7
)

// nonFranchiseSeriesNames - Annictの「シリーズ」のうち、作品群ではなくレーベル・放送枠・
// 制作会社などで無関係な作品を束ねたもの。関連作品の手がかりにすると無関係な作品が
// 並ぶため保存しない（例: 「AnimeFestaオリジナル」は44作品の別々の作品を含む）。
// ローカルDBの全シリーズのうち、タイトルとシリーズ名の重なりが小さいものを目視で選んだ。
var nonFranchiseSeriesNames = map[string]bool{
	"AnimeFestaオリジナル":  true, // レーベル
	"「プチプチ・アニメ」放送作品":   true, // 放送枠
	"世界名作劇場":           true, // 放送枠（作品ごとに原作が異なる）
	"週刊ラノベアニメ":         true, // 番組
	"PROJECT A":        true, // オムニバス企画
	"P.A.WORKSの“お仕事\"": true, // 制作会社の作品群
}

// nonFranchiseSeriesPattern - 年度ごとに名前が変わる若手育成企画（アニメミライ2012〜 等）
var nonFranchiseSeriesPattern = regexp.MustCompile(`^(アニメミライ|あにめたまご|アニメたまご)\s*\d*$`)

// isFranchiseSeries はAnnictのシリーズ名が作品群（関連作品の手がかりに使えるもの）かを返す
func isFranchiseSeries(name string) bool {
	return !nonFranchiseSeriesNames[name] && !nonFranchiseSeriesPattern.MatchString(name)
}

type annictSeriesResponse struct {
	Data struct {
		SearchWorks struct {
			Nodes []struct {
				AnnictID   int `json:"annictId"`
				SeriesList struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"seriesList"`
			} `json:"nodes"`
		} `json:"searchWorks"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// FetchSeriesNames は指定したAnnict作品IDの所属シリーズ名を取得する。
// 戻り値のマップに含まれない作品は、Annict側で見つからなかったもの。
func (c *AnnictClient) FetchSeriesNames(annictIDs []int) (map[int][]string, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("annict API token is not set")
	}
	if len(annictIDs) == 0 {
		return map[int][]string{}, nil
	}

	ids := make([]string, len(annictIDs))
	for i, id := range annictIDs {
		ids[i] = strconv.Itoa(id)
	}
	query := fmt.Sprintf(
		`{ searchWorks(annictIds: [%s], first: %d) { nodes { annictId seriesList { nodes { name } } } } }`,
		strings.Join(ids, ","), len(annictIDs))
	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal graphql query: %w", err)
	}

	c.applyRateLimit()

	req, err := http.NewRequest(http.MethodPost, annictGraphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MyAnimeLogs/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("status %d: %w", resp.StatusCode, ErrRateLimitExceeded)
		}
		return nil, fmt.Errorf("annict graphql returned status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var parsed annictSeriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal annict graphql response: %w", err)
	}
	if len(parsed.Errors) > 0 {
		return nil, fmt.Errorf("annict graphql error: %s", parsed.Errors[0].Message)
	}

	result := make(map[int][]string, len(parsed.Data.SearchWorks.Nodes))
	for _, node := range parsed.Data.SearchWorks.Nodes {
		names := make([]string, 0, len(node.SeriesList.Nodes))
		for _, s := range node.SeriesList.Nodes {
			if name := strings.TrimSpace(s.Name); name != "" && isFranchiseSeries(name) {
				names = append(names, name)
			}
		}
		result[node.AnnictID] = names
	}
	return result, nil
}

// SyncSeriesNames はシリーズ未取得の作品（および所属シリーズが空のまま
// annictSeriesRecheckDays日経った作品）について、Annictから所属シリーズを取得して保存する。
// 新規作品の取り込み直後と、既存作品の初回補完の両方をこの1つの処理で担う。
// 戻り値は保存した作品数。
func (c *AnnictClient) SyncSeriesNames(ctx context.Context) (int, error) {
	if c.DB == nil {
		return 0, fmt.Errorf("database is not configured")
	}

	rows, err := c.DB.QueryContext(ctx, `
		SELECT annict_id FROM animes
		WHERE annict_id IS NOT NULL
		AND (
			annict_series_synced_at IS NULL
			OR (
				cardinality(annict_series_names) = 0
				AND annict_series_synced_at < NOW() - make_interval(days => $1)
			)
		)
		ORDER BY annict_id
	`, annictSeriesRecheckDays)
	if err != nil {
		return 0, fmt.Errorf("failed to query animes for series sync: %w", err)
	}
	var annictIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, fmt.Errorf("failed to scan annict_id for series sync: %w", err)
		}
		annictIDs = append(annictIDs, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("failed to iterate animes for series sync: %w", err)
	}

	if len(annictIDs) == 0 {
		return 0, nil
	}
	if c.logger != nil {
		c.logger.Info("Starting Annict series sync", "target_count", len(annictIDs))
	}

	saved := 0
	for start := 0; start < len(annictIDs); start += annictSeriesBatchSize {
		if err := ctx.Err(); err != nil {
			return saved, err
		}
		end := start + annictSeriesBatchSize
		if end > len(annictIDs) {
			end = len(annictIDs)
		}
		batch := annictIDs[start:end]

		seriesByID, err := c.FetchSeriesNames(batch)
		if err != nil {
			// 1バッチの失敗で残りを止めない。未保存の作品は次回の同期で再び対象になる
			if c.logger != nil {
				c.logger.Warn("Failed to fetch Annict series", "batch_start", batch[0], "error", err)
			}
			continue
		}

		for _, id := range batch {
			// Annict側で見つからなかった作品も空配列で保存し、毎回の再問い合わせを防ぐ
			names := seriesByID[id]
			if names == nil {
				names = []string{}
			}
			if _, err := c.DB.ExecContext(ctx, `
				UPDATE animes
				SET annict_series_names = $1, annict_series_synced_at = NOW()
				WHERE annict_id = $2
			`, pq.Array(names), id); err != nil {
				if c.logger != nil {
					c.logger.Warn("Failed to save Annict series", "annict_id", id, "error", err)
				}
				continue
			}
			saved++
		}
	}

	if c.logger != nil {
		c.logger.Info("Annict series sync completed", "saved_count", saved, "target_count", len(annictIDs))
	}
	return saved, nil
}
