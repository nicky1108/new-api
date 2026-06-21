package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVideoGenerationsAliasReroutesToRegisteredVideoEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetVideoRouter(r)
	r.NoRoute(func(c *gin.Context) {
		if rerouteVideoGenerationsAlias(c, r) {
			return
		}
		c.Status(http.StatusNotFound)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/videos/generations", nil)
	r.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatal("/v1/videos/generations was not rerouted to the registered video endpoint")
	}
}
