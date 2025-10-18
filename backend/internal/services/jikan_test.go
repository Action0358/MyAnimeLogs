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

func TestJikanClient_GetAnimeEpisodes(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")

	t.Run("Success - Single Page", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// Handle anime detail endpoint
			if r.URL.Path == "/anime/1" {
				animeResponse := models.JikanAnimeResponse{
					Data: models.JikanAnimeData{
						MalID:    1,
						Title:    "Test Anime",
						Duration: "24 min per ep",
						Episodes: 2,
					},
				}
				json.NewEncoder(w).Encode(animeResponse)
				return
			}

			// Handle episodes endpoint
			if r.URL.Path == "/anime/1/episodes" {
				episodeResponse := models.JikanEpisodeResponse{
					Data: []models.JikanEpisode{
						{MalID: 1, URL: "https://myanimelist.net/anime/1/Test_Anime/episode/1", Title: "Episode 1"},
						{MalID: 2, URL: "https://myanimelist.net/anime/1/Test_Anime/episode/2", Title: "Episode 2"},
					},
					Pagination: models.JikanPagination{
						LastVisiblePage: 1,
						HasNextPage:     false,
					},
				}
				json.NewEncoder(w).Encode(episodeResponse)
				return
			}
		}))
		defer server.Close()

		client := NewJikanClient(logger)
		client.baseURL = server.URL

		episodes, err := client.GetAnimeEpisodes("1")
		require.NoError(t, err)
		assert.Len(t, episodes, 2)
		assert.Equal(t, "Episode 1", episodes[0].Title)
		assert.Equal(t, 24, episodes[0].Duration)
	})

	t.Run("Success - Multiple Pages", func(t *testing.T) {
		episodePageCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// Handle anime detail endpoint
			if r.URL.Path == "/anime/1" {
				animeResponse := models.JikanAnimeResponse{
					Data: models.JikanAnimeData{
						MalID:    1,
						Title:    "Test Anime",
						Duration: "24 min per ep",
						Episodes: 2,
					},
				}
				json.NewEncoder(w).Encode(animeResponse)
				return
			}

			// Handle episodes endpoint with pagination
			if r.URL.Path == "/anime/1/episodes" {
				episodePageCount++

				var response models.JikanEpisodeResponse
				if episodePageCount == 1 {
					response = models.JikanEpisodeResponse{
						Data: []models.JikanEpisode{
							{MalID: 1, URL: "https://myanimelist.net/anime/1/Test_Anime/episode/1", Title: "Episode 1"},
						},
						Pagination: models.JikanPagination{
							LastVisiblePage: 2,
							HasNextPage:     true,
						},
					}
				} else {
					response = models.JikanEpisodeResponse{
						Data: []models.JikanEpisode{
							{MalID: 2, URL: "https://myanimelist.net/anime/1/Test_Anime/episode/2", Title: "Episode 2"},
						},
						Pagination: models.JikanPagination{
							LastVisiblePage: 2,
							HasNextPage:     false,
						},
					}
				}
				json.NewEncoder(w).Encode(response)
				return
			}
		}))
		defer server.Close()

		client := NewJikanClient(logger)
		client.baseURL = server.URL

		episodes, err := client.GetAnimeEpisodes("1")
		require.NoError(t, err)
		assert.Len(t, episodes, 2)
	})

	t.Run("Error - Empty MAL ID", func(t *testing.T) {
		client := NewJikanClient(logger)
		_, err := client.GetAnimeEpisodes("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mal_anime_id is required")
	})

	t.Run("Error - 404 Not Found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewJikanClient(logger)
		client.baseURL = server.URL

		_, err := client.GetAnimeEpisodes("999999")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrJikanNotFound)
	})

	t.Run("Error - Invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewJikanClient(logger)
		client.baseURL = server.URL

		_, err := client.GetAnimeEpisodes("1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})
}

func TestJikanClient_RateLimiting(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")

		// Handle anime detail endpoint
		if r.URL.Path == "/anime/1" {
			animeResponse := models.JikanAnimeResponse{
				Data: models.JikanAnimeData{
					MalID:    1,
					Title:    "Test Anime",
					Duration: "24 min per ep",
					Episodes: 1,
				},
			}
			json.NewEncoder(w).Encode(animeResponse)
			return
		}

		// Handle episodes endpoint
		response := models.JikanEpisodeResponse{
			Data: []models.JikanEpisode{
				{MalID: 1, URL: "https://myanimelist.net/anime/1/Test_Anime/episode/1", Title: "Episode 1"},
			},
			Pagination: models.JikanPagination{
				LastVisiblePage: 1,
				HasNextPage:     false,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewJikanClient(logger)
	client.baseURL = server.URL

	// Make multiple requests - rate limiter should prevent more than 3 per second
	for i := 0; i < 3; i++ {
		_, err := client.GetAnimeEpisodes("1")
		require.NoError(t, err)
	}

	// Each GetAnimeEpisodes call makes 2 requests (anime detail + episodes)
	assert.Equal(t, 6, requestCount)
}

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

func TestExtractEpisodeNumber(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    int
		expectError bool
	}{
		{
			name:     "正常なURL",
			url:      "https://myanimelist.net/anime/53065/Sono_Bisque_Doll_wa_Koi_wo_Suru_Season_2/episode/1",
			expected: 1,
		},
		{
			name:     "エピソード番号が2桁",
			url:      "https://myanimelist.net/anime/53065/Test_Anime/episode/12",
			expected: 12,
		},
		{
			name:     "エピソード番号が3桁",
			url:      "https://myanimelist.net/anime/53065/Test_Anime/episode/123",
			expected: 123,
		},
		{
			name:        "エピソード番号がない",
			url:         "https://myanimelist.net/anime/53065/",
			expectError: true,
		},
		{
			name:        "不正なURL",
			url:         "invalid-url",
			expectError: true,
		},
		{
			name:        "空のURL",
			url:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractEpisodeNumber(tt.url)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestJikanClient_ExtractEpisodeNumber(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	client := NewJikanClient(logger)

	t.Run("Success", func(t *testing.T) {
		episodeNumber, err := client.ExtractEpisodeNumber("https://myanimelist.net/anime/53065/Test/episode/5")
		require.NoError(t, err)
		assert.Equal(t, 5, episodeNumber)
	})

	t.Run("Error", func(t *testing.T) {
		_, err := client.ExtractEpisodeNumber("invalid-url")
		assert.Error(t, err)
	})
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
					Episodes: 12,
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

// TestJikanClient_RealAPI tests against the real Jikan API (integration test)
// This test is skipped by default. Run with: go test -v -run TestJikanClient_RealAPI
func TestJikanClient_RealAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := common.NewLogger(common.LogLevelInfo, "text")
	client := NewJikanClient(logger)

	t.Run("Real API - Attack on Titan", func(t *testing.T) {
		// MAL ID for Attack on Titan (進撃の巨人)
		episodes, err := client.GetAnimeEpisodes("16498")
		require.NoError(t, err)
		assert.Greater(t, len(episodes), 0, "Should have episodes")

		if len(episodes) > 0 {
			assert.Greater(t, episodes[0].Duration, 0, "Should have duration")
			t.Logf("Successfully fetched %d episodes with duration %d min", len(episodes), episodes[0].Duration)
		}
	})

	t.Run("Real API - Get Duration", func(t *testing.T) {
		duration, err := client.GetAnimeDuration("16498")
		require.NoError(t, err)
		assert.Greater(t, duration, 0, "Should have duration")
		t.Logf("Duration: %d minutes", duration)
	})
}
