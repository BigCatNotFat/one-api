package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

var (
	queryRateLimiter   *cache.Cache
	queryFailLimiter   *cache.Cache
	queryLockList      *cache.Cache
	queryLimiterMutex  sync.Mutex
)

func init() {
	// 查询次数限制：5分钟过期
	queryRateLimiter = cache.New(5*time.Minute, 10*time.Minute)
	// 失败次数限制：1小时过期
	queryFailLimiter = cache.New(1*time.Hour, 2*time.Hour)
	// 锁定列表：10分钟过期
	queryLockList = cache.New(10*time.Minute, 15*time.Minute)
}

// TokenQueryRateLimit 令牌查询限流中间件
// 规则：
// - 每个IP每5分钟最多5次查询
// - 失败3次后锁定10分钟
func TokenQueryRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		// 检查是否在锁定列表中
		lockKey := fmt.Sprintf("lock_%s", ip)
		if _, found := queryLockList.Get(lockKey); found {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "查询失败次数过多，请10分钟后再试",
			})
			c.Abort()
			return
		}

		// 检查查询频率
		rateKey := fmt.Sprintf("rate_%s", ip)
		queryLimiterMutex.Lock()
		
		if count, found := queryRateLimiter.Get(rateKey); found {
			currentCount := count.(int)
			if currentCount >= 5 {
				queryLimiterMutex.Unlock()
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "查询过于频繁，请5分钟后再试",
				})
				c.Abort()
				return
			}
			queryRateLimiter.Set(rateKey, currentCount+1, cache.DefaultExpiration)
		} else {
			queryRateLimiter.Set(rateKey, 1, cache.DefaultExpiration)
		}
		
		queryLimiterMutex.Unlock()

		c.Next()

		// 请求处理完成后，检查是否失败
		// 如果响应中success为false，增加失败计数
		if c.Writer.Status() == http.StatusOK {
			// 获取响应中的success字段需要在QueryTokenQuota中手动调用RecordQueryFailure
		}
	}
}

// RecordQueryFailure 记录查询失败
func RecordQueryFailure(ip string) {
	failKey := fmt.Sprintf("fail_%s", ip)
	lockKey := fmt.Sprintf("lock_%s", ip)
	
	queryLimiterMutex.Lock()
	defer queryLimiterMutex.Unlock()
	
	if count, found := queryFailLimiter.Get(failKey); found {
		currentCount := count.(int)
		newCount := currentCount + 1
		queryFailLimiter.Set(failKey, newCount, cache.DefaultExpiration)
		
		// 失败3次，锁定10分钟
		if newCount >= 3 {
			queryLockList.Set(lockKey, true, 10*time.Minute)
			queryFailLimiter.Delete(failKey) // 清除失败计数
		}
	} else {
		queryFailLimiter.Set(failKey, 1, cache.DefaultExpiration)
	}
}

// ResetQueryFailure 重置查询失败计数（查询成功时调用）
func ResetQueryFailure(ip string) {
	failKey := fmt.Sprintf("fail_%s", ip)
	queryLimiterMutex.Lock()
	defer queryLimiterMutex.Unlock()
	queryFailLimiter.Delete(failKey)
}
