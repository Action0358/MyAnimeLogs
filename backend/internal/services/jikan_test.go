package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"24 min per ep", "24 min per ep", 24},
		{"1 hr 30 min", "1 hr 30 min", 90},
		{"1 hr", "1 hr", 60},
		{"45 min", "45 min", 45},
		{"90 sec", "90 sec", 1},
		{"Unknown", "Unknown", 0},
		{"Empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDuration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJikanClient_GetAnimeDuration(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.Path, "/anime/1")

			response := models.JikanAnimeResponse{
				Data: models.JikanAnimeData{
					MalID:    1,
					Title:    "Test Anime",
					Duration: "24 min per ep",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewJikanClient(logger)
		client.baseURL = server.URL

		duration, err := client.GetAnimeDuration("1")
		require.NoError(t, err)
		assert.Equal(t, 24, duration)
	})

	t.Run("Error - Empty MAL ID", func(t *testing.T) {
		client := NewJikanClient(logger)
		_, err := client.GetAnimeDuration("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mal_anime_id is required")
	})
}
