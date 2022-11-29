package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ciscolive/gin-admin/global"
	"github.com/ciscolive/gin-admin/model/common/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LimitConfig struct {
	GenerationKey func(c *gin.Context) string                   // GenerationKey 根据业务生成key 下面CheckOrMark查询生成
	CheckOrMark   func(key string, expire int, limit int) error // 检查函数,用户可修改具体逻辑,更加灵活
	Expire        int                                           // Expire key 过期时间
	Limit         int                                           // Limit 周期时间
}

func (l LimitConfig) LimitWithTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := l.CheckOrMark(l.GenerationKey(c), l.Expire, l.Limit)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": response.ERROR, "msg": err})
			c.Abort()
			return
		}
		c.Next()
	}
}

// DefaultGenerationKey 默认生成key
func DefaultGenerationKey(c *gin.Context) string {
	return "GVA_Limit" + c.ClientIP()
}

func DefaultCheckOrMark(key string, expire int, limit int) (err error) {
	// 判断是否开启redis
	if global.Redis == nil {
		return err
	}
	if err = SetLimitWithTime(key, limit, time.Duration(expire)*time.Second); err != nil {
		global.Logger.Error("limit", zap.Error(err))
	}
	return err
}

func DefaultLimit() gin.HandlerFunc {
	return LimitConfig{
		GenerationKey: DefaultGenerationKey,
		CheckOrMark:   DefaultCheckOrMark,
		Expire:        global.Config.System.LimitTimeIP,
		Limit:         global.Config.System.LimitCountIP,
	}.LimitWithTime()
}

// SetLimitWithTime 设置访问次数
func SetLimitWithTime(key string, limit int, expiration time.Duration) error {
	count, err := global.Redis.Exists(context.Background(), key).Result()
	if err != nil {
		return err
	}
	if count == 0 {
		pipe := global.Redis.TxPipeline()
		pipe.Incr(context.Background(), key)
		pipe.Expire(context.Background(), key, expiration)
		_, err = pipe.Exec(context.Background())
		return err
	}
	// 次数
	times, err := global.Redis.Get(context.Background(), key).Int()
	if err != nil {
		return err
	}
	if times >= limit {
		t, err := global.Redis.PTTL(context.Background(), key).Result()
		if err != nil {
			return errors.New("请求太过频繁，请稍后再试")
		}
		return errors.New("请求太过频繁, 请 " + t.String() + " 秒后尝试")
	}
	return global.Redis.Incr(context.Background(), key).Err()
}
