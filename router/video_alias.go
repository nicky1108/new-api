package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func rerouteVideoGenerationsAlias(c *gin.Context, router *gin.Engine) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil || router == nil {
		return false
	}

	path := c.Request.URL.Path
	switch {
	case c.Request.Method == http.MethodPost && path == "/v1/videos/generations":
		c.Request.URL.Path = "/v1/videos"
	case c.Request.Method == http.MethodGet && strings.HasPrefix(path, "/v1/videos/generations/"):
		taskID := strings.TrimPrefix(path, "/v1/videos/generations/")
		if taskID == "" || strings.Contains(taskID, "/") {
			return false
		}
		c.Request.URL.Path = "/v1/videos/" + taskID
	default:
		return false
	}

	router.HandleContext(c)
	return true
}
