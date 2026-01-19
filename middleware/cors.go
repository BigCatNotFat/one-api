package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	// AllowAllOrigins 和 AllowCredentials 不能同时为 true
	// 使用 AllowOriginFunc 来动态允许所有来源
	config.AllowOriginFunc = func(origin string) bool {
		return true // 允许所有来源
	}
	config.AllowCredentials = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	return cors.New(config)
}
