package middleware

import (
	"github.com/ciscolive/gin-admin/config"
	"github.com/ciscolive/gin-admin/global"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Cors 直接放行所有跨域请求并放行所有 OPTIONS 方法
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token,X-Token,X-User-Id")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS,DELETE,PUT")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type, New-Token, New-Expires-At")
		c.Header("Access-Control-Allow-Credentials", "true")
		// 放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next() // 处理请求
	}
}

// CorsByRules 按照配置处理跨域请求
func CorsByRules() gin.HandlerFunc {
	if global.Config.Cors.Mode == "allow-all" {
		return Cors()
	}
	return func(c *gin.Context) {
		whitelist := checkCors(c.GetHeader("origin"))
		// 通过检查, 添加请求头
		if whitelist != nil {
			c.Header("Access-Control-Allow-Origin", whitelist.AllowOrigin)
			c.Header("Access-Control-Allow-Headers", whitelist.AllowHeaders)
			c.Header("Access-Control-Allow-Methods", whitelist.AllowMethods)
			c.Header("Access-Control-Expose-Headers", whitelist.ExposeHeaders)
			if whitelist.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}
		// 严格白名单模式且未通过检查，直接拒绝处理请求
		switch {
		case c.Request.Method == "GET" && c.Request.URL.Path == "/health":
			c.Next()
		case whitelist == nil && global.Config.Cors.Mode == "strict-whitelist":
			c.AbortWithStatus(http.StatusForbidden)
		case c.Request.Method == "OPTIONS":
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func checkCors(origin string) *config.CORSWhitelist {
	for _, whitelist := range global.Config.Cors.Whitelist {
		// 遍历配置中的跨域头，寻找匹配项
		if origin == whitelist.AllowOrigin {
			return &whitelist
		}
	}
	return nil
}
