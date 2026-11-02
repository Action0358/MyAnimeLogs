package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterDevelopmentRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	status := func(environment string) int {
		r := gin.New()
		registerDevelopmentRoutes(r, environment, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test-hot-reload", nil))
		return w.Code
	}

	t.Run("開発環境では登録する", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, status("development"))
	})

	for _, environment := range []string{"staging", "production", "test"} {
		t.Run(environment+"では登録しない", func(t *testing.T) {
			assert.Equal(t, http.StatusNotFound, status(environment))
		})
	}

	t.Run("公開先の環境ではAnnictのテストルートも登録しない", func(t *testing.T) {
		r := gin.New()
		registerDevelopmentRoutes(r, "staging", nil)
		assert.Empty(t, r.Routes())
	})
}
