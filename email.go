package middleware

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ciscolive/gin-admin/plugin/email/utils"
	sUtils "github.com/ciscolive/gin-admin/utils"

	"github.com/ciscolive/gin-admin/global"
	"github.com/ciscolive/gin-admin/model/system"
	"github.com/ciscolive/gin-admin/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var userService = service.Context.System.User

func ErrorToEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var username string
		claims, _ := sUtils.GetClaims(c)
		if claims.Username != "" {
			username = claims.Username
		} else {
			id, _ := strconv.Atoi(c.Request.Header.Get("x-user-id"))
			user, err := userService.FindUserByID(id)
			if err != nil {
				username = "Unknown"
			} else {
				username = user.Username
			}
		}
		// 再重新写回请求体body中，ioutil.ReadAll会清空c.Request.Body中的数据
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		record := system.SysOperationRecord{
			Ip:     c.ClientIP(),
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
			Agent:  c.Request.UserAgent(),
			Body:   string(body),
		}
		now := time.Now()
		c.Next()
		latency := time.Since(now)
		status := c.Writer.Status()
		record.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		// str := "接收到的请求为" + record.Body + "\n" + "请求方式为" + record.Method + "\n" + "报错信息如下" + record.ErrorMessage + "\n" + "耗时" + latency.String() + "\n"
		str := fmt.Sprintf(`
		收到请求:%s;
		请求方式:%s;
		报错信息:%s;
		请求耗时:%s;
		`, record.Body, record.Method, record.ErrorMessage, latency.String())
		if status != 200 {
			subject := "【GIN中间件监控告警】" + username + "通过地址" + record.Ip + "访问URL-" + record.Path + "异常"
			if err := utils.ErrorToEmail(subject, str); err != nil {
				global.Logger.Error("GIN中间件监控告警工作异常：", zap.Error(err))
			}
		}
	}
}
