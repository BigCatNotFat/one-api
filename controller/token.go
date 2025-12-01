package controller

import (
	"fmt"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/network"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/model"
	"net/http"
	"strconv"
)

func GetAllTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	order := c.Query("order")
	tokens, err := model.GetAllUserTokens(userId, p*config.ItemsPerPage, config.ItemsPerPage, order)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokens,
	})
	return
}

func SearchTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	keyword := c.Query("keyword")
	tokens, err := model.SearchUserTokens(userId, keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokens,
	})
	return
}

func GetToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	token, err := model.GetTokenByIds(id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    token,
	})
	return
}

func GetTokenStatus(c *gin.Context) {
	tokenId := c.GetInt(ctxkey.TokenId)
	userId := c.GetInt(ctxkey.Id)
	token, err := model.GetTokenByIds(tokenId, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	expiredAt := token.ExpiredTime
	if expiredAt == -1 {
		expiredAt = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"object":          "credit_summary",
		"total_granted":   token.RemainQuota,
		"total_used":      0, // not supported currently
		"total_available": token.RemainQuota,
		"expires_at":      expiredAt * 1000,
	})
}

func validateToken(c *gin.Context, token model.Token) error {
	if len(token.Name) > 30 {
		return fmt.Errorf("令牌名称过长")
	}
	if token.Subnet != nil && *token.Subnet != "" {
		err := network.IsValidSubnets(*token.Subnet)
		if err != nil {
			return fmt.Errorf("无效的网段：%s", err.Error())
		}
	}
	return nil
}

func AddToken(c *gin.Context) {
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数错误：%s", err.Error()),
		})
		return
	}

	cleanToken := model.Token{
		UserId:         c.GetInt(ctxkey.Id),
		Name:           token.Name,
		Key:            random.GenerateKey(),
		CreatedTime:    helper.GetTimestamp(),
		AccessedTime:   helper.GetTimestamp(),
		ExpiredTime:    token.ExpiredTime,
		RemainQuota:    token.RemainQuota,
		UnlimitedQuota: token.UnlimitedQuota,
		Models:         token.Models,
		Subnet:         token.Subnet,
		QueryPIN:       random.GetRandomNumberString(4), // 生成4位数字PIN码
	}
	err = cleanToken.Insert()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanToken,
	})
	return
}

func DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	err := model.DeleteTokenById(id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func UpdateToken(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	statusOnly := c.Query("status_only")
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数错误：%s", err.Error()),
		})
		return
	}
	cleanToken, err := model.GetTokenByIds(token.Id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if token.Status == model.TokenStatusEnabled {
		if cleanToken.Status == model.TokenStatusExpired && cleanToken.ExpiredTime <= helper.GetTimestamp() && cleanToken.ExpiredTime != -1 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌已过期，无法启用，请先修改令牌过期时间，或者设置为永不过期",
			})
			return
		}
		if cleanToken.Status == model.TokenStatusExhausted && cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌可用额度已用尽，无法启用，请先修改令牌剩余额度，或者设置为无限额度",
			})
			return
		}
	}
	if statusOnly != "" {
		cleanToken.Status = token.Status
	} else {
		// If you add more fields, please also update token.Update()
		cleanToken.Name = token.Name
		cleanToken.ExpiredTime = token.ExpiredTime
		cleanToken.RemainQuota = token.RemainQuota
		cleanToken.UnlimitedQuota = token.UnlimitedQuota
		cleanToken.Models = token.Models
		cleanToken.Subnet = token.Subnet
		cleanToken.QueryPIN = token.QueryPIN
	}
	err = cleanToken.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanToken,
	})
	return
}

// QueryTokenQuota 公开查询令牌额度信息（无需登录）
func QueryTokenQuota(c *gin.Context) {
	key := c.Query("key")
	ip := c.ClientIP()

	// 辅助函数：失败响应（带延迟）
	failWithDelay := func(message string) {
		// 延迟1-3秒返回，防止暴力破解
		delay := random.RandRange(1, 4) // 1到3秒随机延迟
		time.Sleep(time.Duration(delay) * time.Second)
		middleware.RecordQueryFailure(ip)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": message,
		})
	}

	if key == "" {
		failWithDelay("请提供API Key")
		return
	}

	// 归一化 API Key：去空格，去掉 sk- 前缀，并只保留第一段作为真实 key
	key = strings.TrimSpace(key)
	key = strings.TrimPrefix(key, "sk-")
	parts := strings.Split(key, "-")
	if len(parts) > 0 {
		key = parts[0]
	}

	// 获取令牌信息
	token, err := model.GetTokenByKey(key)
	if err != nil {
		failWithDelay("查询失败，请检查API Key是否正确")
		return
	}

	// 查询成功，重置失败计数
	middleware.ResetQueryFailure(ip)

	// 获取分页参数
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	// 限制只返回最近24小时的数据
	now := helper.GetTimestamp()
	oneDayAgo := now - 86400
	startTimestamp := oneDayAgo
	endTimestamp := now

	// 如果用户提供了时间范围，也要确保不超过24小时
	userStart, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	userEnd, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	if userStart > 0 && userStart > oneDayAgo {
		startTimestamp = userStart
	}
	if userEnd > 0 && userEnd < now {
		endTimestamp = userEnd
	}

	// 查询使用日志（限制为最近24小时）
	logs, err := model.GetTokenLogs(token.Name, startTimestamp, endTimestamp, p*10, 10)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "查询日志失败",
		})
		return
	}

	// 计算总使用额度
	totalUsedQuota := model.SumTokenUsedQuota(token.Name, startTimestamp, endTimestamp)

	// 构建响应数据
	expiredTime := token.ExpiredTime
	if expiredTime == -1 {
		expiredTime = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"name":             token.Name,
			"status":           token.Status,
			"remain_quota":     token.RemainQuota,
			"used_quota":       token.UsedQuota,
			"unlimited_quota":  token.UnlimitedQuota,
			"expired_time":     expiredTime,
			"created_time":     token.CreatedTime,
			"total_used_quota": totalUsedQuota,
			"logs":             logs,
			"time_range":       "最近24小时",
		},
	})
}
