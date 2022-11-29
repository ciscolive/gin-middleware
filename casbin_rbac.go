package middleware

import (
	"github.com/ciscolive/gin-admin/global"
	"github.com/ciscolive/gin-admin/model/common/response"
	"github.com/ciscolive/gin-admin/service"
	"github.com/ciscolive/gin-admin/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

var casbinService = service.Context.System.Casbin

func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.Config.System.Env != "develop" {
			getClaims, _ := utils.GetClaims(c)
			obj := c.Request.URL.Path                       // 获取请求的PATH
			act := c.Request.Method                         // 获取请求方法
			sub := strconv.Itoa(int(getClaims.AuthorityID)) // 获取用户的角色
			e := casbinService.Casbin()                     // 判断策略中是否存在
			success, _ := e.Enforce(sub, obj, act)
			if !success {
				response.FailWithDetailed(gin.H{}, "权限不足", c)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
