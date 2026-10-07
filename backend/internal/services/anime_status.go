package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"myanimelogs/internal/constants"
)

// enrichAnimeWithWatchProviders - TMDBから配信サービス情報を取得してアニメ情報を拡張
// series_tmdb_idがある場合はそれを優先的に使用（シリーズ作品の統合管理）
// determineMediaType determines the TMDB media type ("tv" or "movie") based on
// DB media field, model data, and AniList format as a fallback.
//
// The second return value reports whether the type is backed by an actual source
// value rather than the default. Annict's media field is one of tv/movie/ova/web/other:
// only tv and movie map cleanly onto TMDB's two types, so ova/web/other (and a missing
// value) fall through to "tv" as a guess. Callers use the flag to decide whether
// retrying the opposite media type is worth an extra TMDB round trip.
func (s *AnimeService) determineMediaType(ctx context.Context, dbMediaType sql.NullString, modelMedia *string, malID string) (string, bool) {
	if dbMediaType.Valid {
		switch dbMediaType.String {
		case "movie":
			return "movie", true
		case "tv":
			return "tv", true
		}
	}
	if modelMedia != nil {
		switch *modelMedia {
		case "movie":
			return "movie", true
		case "tv":
			return "tv", true
		}
	}
	if malID != "" {
		format, err := s.anilistClient.GetMediaFormatByMALId(ctx, malID)
		if err == nil && format == "MOVIE" {
			return "movie", true
		}
	}
	return "tv", false
}

// seasonStartGracePeriodDays - シーズン開始月のうち、released_on未確定の作品を
// 「放送予定」扱いとする猶予日数。この期間を過ぎてもreleased_onが無ければ、
// 従来通りシーズン期間内＝放送中とみなす
const seasonStartGracePeriodDays = 14

// determineStatusBySeason はシーズンと放送開始日を考慮してアニメの放送状態を判定します。
// TMDBの情報が利用できない場合のフォールバックとして使用されます。
func determineStatusBySeason(season *string, releasedOn *string) string {
	var seasonName string
	if season != nil {
		seasonName = *season
	}
	return determineStatusBySeasonAt(seasonName, releasedOn, time.Now().UTC())
}

// determineStatusBySeasonAt はシーズンと放送開始日を考慮してアニメの放送状態を判定します。
// 基準時刻を引数で受け取ることでテスタビリティを確保しています。
func determineStatusBySeasonAt(seasonName string, releasedOn *string, now time.Time) string {
	// released_onが未来の日付なら放送予定
	if releasedOn != nil && *releasedOn != "" {
		// UTCで比較してサーバータイムゾーンの影響を排除
		releaseDate, err := time.ParseInLocation("2006-01-02", *releasedOn, time.UTC)
		if err == nil {
			today := now.Truncate(24 * time.Hour)
			if releaseDate.After(today) {
				return constants.AnimeStatusUpcoming.String()
			}
		}
	}

	if seasonName == "" {
		return constants.AnimeStatusFinished.String()
	}

	// シーズン名（例: "YYYY-season"）をパース
	var seasonYear int
	var seasonStr string
	_, err := fmt.Sscanf(seasonName, "%d-%s", &seasonYear, &seasonStr)
	if err != nil {
		return constants.AnimeStatusFinished.String() // パース失敗時は完結とみなす
	}

	currentYear := now.Year()
	currentMonth := int(now.Month())

	// シーズンの開始月と終了月を判定
	var seasonStartMonth, seasonEndMonth int
	switch seasonStr {
	case "winter": // 1月, 2月, 3月
		seasonStartMonth = 1
		seasonEndMonth = 3
	case "spring": // 4月, 5月, 6月
		seasonStartMonth = 4
		seasonEndMonth = 6
	case "summer": // 7月, 8月, 9月
		seasonStartMonth = 7
		seasonEndMonth = 9
	case "autumn", "fall": // 10月, 11月, 12月
		seasonStartMonth = 10
		seasonEndMonth = 12
	default:
		return constants.AnimeStatusFinished.String() // 未知のシーズン名は完結とみなす
	}

	// 判定ロジック
	if seasonYear > currentYear {
		// 未来のシーズン → 放送予定
		return constants.AnimeStatusUpcoming.String()
	} else if seasonYear == currentYear {
		// 同じ年のシーズン
		if currentMonth < seasonStartMonth {
			// シーズン開始月より前 → 放送予定
			return constants.AnimeStatusUpcoming.String()
		} else if currentMonth <= seasonEndMonth {
			// シーズン開始月の最初の14日間は、released_onが未確定（Annict/Jikan双方とも
			// 放送開始日を取得できていない）作品を「放送予定」として扱う。
			// released_onが無くても season の月範囲に入っていれば放送中とみなす現行ロジックは、
			// シーズン開始直後は多くの作品がまだ未確定のreleased_onを抱えており、
			// 実際にはまだ初回放送されていない作品まで「放送中」と誤表示してしまうため。
			if currentMonth == seasonStartMonth && now.Day() <= seasonStartGracePeriodDays &&
				(releasedOn == nil || *releasedOn == "") {
				return constants.AnimeStatusUpcoming.String()
			}
			// シーズン期間中 → 放送中
			return constants.AnimeStatusAiring.String()
		} else {
			// シーズン終了月以後 → 完結
			return constants.AnimeStatusFinished.String()
		}
	} else {
		// 過去のシーズン → 完結
		return constants.AnimeStatusFinished.String()
	}
}
