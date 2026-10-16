package services

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// 放送開始日（released_on）の補完元。
// Annictのreleased_onは放送直前まで空のことが多く、Jikanは応答しないことがあるため、
// 次の順で補う: Annictの放送枠（日本の初回放送）→ AniListの開始日 → Jikan。

const (
	// releaseDateBatchSize - Annict searchWorks・AniList Pageの1リクエストあたりの作品数
	releaseDateBatchSize = 50

	// releaseDateSeasonLeadDays - シーズン開始日より何日前までの開始日を採用するか。
	// 海外での先行配信日（例: 中国配信が前年12月）を日本の放送開始日として取り込まないため
	releaseDateSeasonLeadDays = 31

	// releaseDateSeasonLagDays - シーズン開始日から何日後までの開始日を採用するか。
	// 海外での後年の放送・配信日や、MAL IDの対応違いで別作品の日付を取り込まないため。
	// 全シーズンの補完で、1年を超える差は大半が誤り（例: 1988年のTV作品に2017年の日付）だった
	releaseDateSeasonLagDays = 365
)

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

type annictProgramsData struct {
	SearchWorks struct {
		Nodes []struct {
			AnnictID int `json:"annictId"`
			Programs struct {
				Nodes []struct {
					StartedAt   string `json:"startedAt"`
					Rebroadcast bool   `json:"rebroadcast"`
				} `json:"nodes"`
			} `json:"programs"`
		} `json:"nodes"`
	} `json:"searchWorks"`
}

// FetchFirstBroadcastDates は指定したAnnict作品IDの初回放送日（日本時間のYYYY-MM-DD）を取得する。
// 再放送を除いた放送枠のうち最も早いものを使う。放送枠が未登録の作品はマップに含まれない。
func (c *AnnictClient) FetchFirstBroadcastDates(annictIDs []int) (map[int]string, error) {
	if len(annictIDs) == 0 {
		return map[int]string{}, nil
	}

	query := fmt.Sprintf(
		`{ searchWorks(annictIds: [%s], first: %d) { nodes { annictId programs(first: 10, orderBy: {field: STARTED_AT, direction: ASC}) { nodes { startedAt rebroadcast } } } } }`,
		joinIntIDs(annictIDs), len(annictIDs))
	var data annictProgramsData
	if err := c.postGraphQL(query, &data); err != nil {
		return nil, err
	}

	result := make(map[int]string, len(data.SearchWorks.Nodes))
	for _, node := range data.SearchWorks.Nodes {
		for _, p := range node.Programs.Nodes {
			if p.Rebroadcast {
				continue
			}
			startedAt, err := time.Parse(time.RFC3339, p.StartedAt)
			if err != nil {
				continue
			}
			result[node.AnnictID] = startedAt.In(jst).Format("2006-01-02")
			break
		}
	}
	return result, nil
}

type anilistStartDatesResponse struct {
	Data struct {
		Page struct {
			Media []struct {
				IDMal     int `json:"idMal"`
				StartDate struct {
					Year  *int `json:"year"`
					Month *int `json:"month"`
					Day   *int `json:"day"`
				} `json:"startDate"`
			} `json:"media"`
		} `json:"Page"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// GetStartDatesByMALIds はMAL IDごとの放送・公開開始日（YYYY-MM-DD）を取得する。
// 年月日がそろっている作品だけを返す（年や月だけの予定日は未確定とみなす）。
// malIDsはreleaseDateBatchSize件以下で渡すこと。
func (c *AniListClient) GetStartDatesByMALIds(ctx context.Context, malIDs []int) (map[int]string, error) {
	if len(malIDs) == 0 {
		return map[int]string{}, nil
	}

	query := `
	query ($ids: [Int], $perPage: Int) {
		Page(perPage: $perPage) {
			media(idMal_in: $ids, type: ANIME) {
				idMal
				startDate { year month day }
			}
		}
	}`
	variables := map[string]interface{}{
		"ids":     malIDs,
		"perPage": len(malIDs),
	}

	result, err := executeGraphQLQuery[anilistStartDatesResponse](c, ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get start dates by MAL IDs: %w", err)
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	dates := make(map[int]string, len(result.Data.Page.Media))
	for _, m := range result.Data.Page.Media {
		d := m.StartDate
		if d.Year == nil || d.Month == nil || d.Day == nil {
			continue
		}
		dates[m.IDMal] = fmt.Sprintf("%04d-%02d-%02d", *d.Year, *d.Month, *d.Day)
	}
	return dates, nil
}

// seasonStartDate は "YYYY-season" の開始日を返す
func seasonStartDate(season string) (time.Time, bool) {
	var year int
	var name string
	if _, err := fmt.Sscanf(season, "%d-%s", &year, &name); err != nil {
		return time.Time{}, false
	}
	months := map[string]time.Month{"winter": 1, "spring": 4, "summer": 7, "autumn": 10}
	month, ok := months[strings.ToLower(name)]
	if !ok {
		return time.Time{}, false
	}
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC), true
}

// isPlausibleReleaseDate は外部から取得した開始日を、その作品のシーズンの放送開始日として
// 採用してよいかを返す。シーズン開始より大きく前の日付（海外先行・旧作の初出など）は採用しない。
// シーズン後の日付は1年以内なら採用する（秋アニメ扱いの映画が12月公開、OVAの後発売など実在するため）。
func isPlausibleReleaseDate(season, date string) bool {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	start, ok := seasonStartDate(season)
	if !ok {
		return false
	}
	return !d.Before(start.AddDate(0, 0, -releaseDateSeasonLeadDays)) &&
		!d.After(start.AddDate(0, 0, releaseDateSeasonLagDays))
}
