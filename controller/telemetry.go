package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/model"
)

// TelemetryPayload 遥测数据请求体
type TelemetryPayload struct {
	UserId     string                 `json:"userId"`
	SessionId  string                 `json:"sessionId"`
	Version    string                 `json:"version"`
	Timestamp  int64                  `json:"timestamp"`
	Statistics *AggregatedStatistics  `json:"statistics"`
}

// AggregatedStatistics 聚合统计数据
type AggregatedStatistics struct {
	Chat         map[string]map[string]int `json:"chat"`
	Tools        map[string]*ToolCount     `json:"tools"`
	ToolApproval map[string]*ApprovalCount `json:"toolApproval"`
	TextActions  map[string]*TextActionCount `json:"textActions"`
	Session      *SessionCount             `json:"session"`
	UI           *UICount                  `json:"ui"`
}

// ToolCount 工具计数
type ToolCount struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// ApprovalCount 审批计数
type ApprovalCount struct {
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
}

// TextActionCount 文本操作计数
type TextActionCount struct {
	Used     int `json:"used"`
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
}

// SessionCount 会话计数
type SessionCount struct {
	Started int `json:"started"`
	Ended   int `json:"ended"`
}

// UICount UI 计数
type UICount struct {
	BranchCreated int `json:"branchCreated"`
	PaneCount     int `json:"paneCount"`
}

// ReceiveTelemetry 接收遥测数据
// POST /api/telemetry
func ReceiveTelemetry(c *gin.Context) {
	var payload TelemetryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// 验证必填字段
	if payload.UserId == "" || payload.SessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing required fields: userId and sessionId",
		})
		return
	}

	// 计算使用度分数
	usageScore := calculateUsageScore(payload.Statistics)

	// 创建主事件记录
	event := &model.TelemetryEvent{
		UserId:          payload.UserId,
		SessionId:       payload.SessionId,
		Version:         payload.Version,
		ClientTimestamp: payload.Timestamp,
		ServerTimestamp: time.Now().UnixMilli(),
		UsageScore:      usageScore,
	}

	if err := model.CreateTelemetryEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save telemetry event: " + err.Error(),
		})
		return
	}

	// 处理统计数据
	if payload.Statistics != nil {
		stats := payload.Statistics

		// 保存聊天统计
		if stats.Chat != nil {
			var chats []*model.TelemetryChat
			for modelId, modes := range stats.Chat {
				for mode, count := range modes {
					if count > 0 {
						chats = append(chats, &model.TelemetryChat{
							EventId: event.Id,
							ModelId: modelId,
							Mode:    mode,
							Count:   count,
						})
					}
				}
			}
			if err := model.CreateTelemetryChats(chats); err != nil {
				// 记录错误但不中断请求
				c.Error(err)
			}
		}

		// 保存工具统计
		if stats.Tools != nil {
			var tools []*model.TelemetryTool
			for toolName, counts := range stats.Tools {
				if counts.Success > 0 || counts.Failed > 0 {
					tools = append(tools, &model.TelemetryTool{
						EventId:      event.Id,
						ToolName:     toolName,
						SuccessCount: counts.Success,
						FailedCount:  counts.Failed,
					})
				}
			}
			if err := model.CreateTelemetryTools(tools); err != nil {
				c.Error(err)
			}
		}

		// 保存工具审批统计
		if stats.ToolApproval != nil {
			var approvals []*model.TelemetryToolApproval
			for toolName, counts := range stats.ToolApproval {
				if counts.Approved > 0 || counts.Rejected > 0 {
					approvals = append(approvals, &model.TelemetryToolApproval{
						EventId:       event.Id,
						ToolName:      toolName,
						ApprovedCount: counts.Approved,
						RejectedCount: counts.Rejected,
					})
				}
			}
			if err := model.CreateTelemetryToolApprovals(approvals); err != nil {
				c.Error(err)
			}
		}

		// 保存文本操作统计
		if stats.TextActions != nil {
			var actions []*model.TelemetryTextAction
			for action, counts := range stats.TextActions {
				if counts.Used > 0 || counts.Accepted > 0 || counts.Rejected > 0 {
					actions = append(actions, &model.TelemetryTextAction{
						EventId:       event.Id,
						Action:        action,
						UsedCount:     counts.Used,
						AcceptedCount: counts.Accepted,
						RejectedCount: counts.Rejected,
					})
				}
			}
			if err := model.CreateTelemetryTextActions(actions); err != nil {
				c.Error(err)
			}
		}

		// 保存会话统计
		if stats.Session != nil && (stats.Session.Started > 0 || stats.Session.Ended > 0) {
			session := &model.TelemetrySession{
				EventId:      event.Id,
				StartedCount: stats.Session.Started,
				EndedCount:   stats.Session.Ended,
			}
			if err := model.CreateTelemetrySession(session); err != nil {
				c.Error(err)
			}
		}

		// 保存 UI 统计
		if stats.UI != nil && (stats.UI.BranchCreated > 0 || stats.UI.PaneCount > 0) {
			ui := &model.TelemetryUI{
				EventId:            event.Id,
				BranchCreatedCount: stats.UI.BranchCreated,
				PaneCount:          stats.UI.PaneCount,
			}
			if err := model.CreateTelemetryUI(ui); err != nil {
				c.Error(err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// GetTelemetryEvents 获取遥测事件列表
// GET /api/telemetry/events
func GetTelemetryEvents(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	events, err := model.GetTelemetryEvents(startTimestamp, endTimestamp, p*config.ItemsPerPage, config.ItemsPerPage)
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
		"data":    events,
	})
}

// GetTelemetryOverview 获取遥测概览统计
// GET /api/telemetry/stats/overview
func GetTelemetryOverview(c *gin.Context) {
	overview, err := model.GetTelemetryOverview()
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
		"data":    overview,
	})
}

// GetTelemetryChatStats 获取聊天统计
// GET /api/telemetry/stats/chat
func GetTelemetryChatStats(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	stats, err := model.GetTelemetryChatStats(startTimestamp, endTimestamp)
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
		"data":    stats,
	})
}

// GetTelemetryToolStats 获取工具统计
// GET /api/telemetry/stats/tools
func GetTelemetryToolStats(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	stats, err := model.GetTelemetryToolStats(startTimestamp, endTimestamp)
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
		"data":    stats,
	})
}

// GetTelemetryDAUStats 获取日活统计
// GET /api/telemetry/stats/dau
func GetTelemetryDAUStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.Query("days"))
	if days <= 0 {
		days = 30 // 默认 30 天
	}
	if days > 365 {
		days = 365 // 最多 365 天
	}

	stats, err := model.GetTelemetryDAUStats(days)
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
		"data":    stats,
	})
}

// GetTelemetryTextActionStats 获取文本操作统计
// GET /api/telemetry/stats/text-actions
func GetTelemetryTextActionStats(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	stats, err := model.GetTelemetryTextActionStats(startTimestamp, endTimestamp)
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
		"data":    stats,
	})
}

// GetTelemetryToolApprovalStats 获取工具审批统计
// GET /api/telemetry/stats/tool-approvals
func GetTelemetryToolApprovalStats(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	stats, err := model.GetTelemetryToolApprovalStats(startTimestamp, endTimestamp)
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
		"data":    stats,
	})
}

// GetTelemetryAllStats 获取所有统计数据（聚合接口）
// GET /api/telemetry/stats/all
func GetTelemetryAllStats(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	days, _ := strconv.Atoi(c.Query("days"))
	if days <= 0 {
		days = 30
	}

	// 获取概览
	overview, _ := model.GetTelemetryOverview()
	
	// 获取聊天统计
	chatStats, _ := model.GetTelemetryChatStats(startTimestamp, endTimestamp)
	
	// 获取工具统计
	toolStats, _ := model.GetTelemetryToolStats(startTimestamp, endTimestamp)
	
	// 获取日活统计
	dauStats, _ := model.GetTelemetryDAUStats(days)
	
	// 获取文本操作统计
	textActionStats, _ := model.GetTelemetryTextActionStats(startTimestamp, endTimestamp)
	
	// 获取工具审批统计
	toolApprovalStats, _ := model.GetTelemetryToolApprovalStats(startTimestamp, endTimestamp)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"overview":      overview,
			"chat":          chatStats,
			"tools":         toolStats,
			"dau":           dauStats,
			"textActions":   textActionStats,
			"toolApprovals": toolApprovalStats,
		},
	})
}

// GetTelemetryTimelineStats 获取时间轴统计（按半小时分组）
// GET /api/telemetry/stats/timeline
func GetTelemetryTimelineStats(c *gin.Context) {
	hours, _ := strconv.Atoi(c.Query("hours"))
	if hours <= 0 {
		hours = 24 // 默认24小时
	}

	stats, err := model.GetTelemetryTimelineStats(hours)
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
		"data":    stats,
	})
}

// calculateUsageScore 计算使用度分数
// 简单将所有操作数量相加，分数越高表示使用率越高，分数为0表示没有使用
func calculateUsageScore(stats *AggregatedStatistics) int {
	if stats == nil {
		return 0
	}

	score := 0

	// 聊天次数
	if stats.Chat != nil {
		for _, modes := range stats.Chat {
			for _, count := range modes {
				score += count
			}
		}
	}

	// 工具使用次数（成功+失败）
	if stats.Tools != nil {
		for _, tool := range stats.Tools {
			score += tool.Success + tool.Failed
		}
	}

	// 工具审批次数（批准+拒绝）
	if stats.ToolApproval != nil {
		for _, approval := range stats.ToolApproval {
			score += approval.Approved + approval.Rejected
		}
	}

	// 文本操作次数
	if stats.TextActions != nil {
		for _, action := range stats.TextActions {
			score += action.Used + action.Accepted + action.Rejected
		}
	}

	// 会话次数
	if stats.Session != nil {
		score += stats.Session.Started + stats.Session.Ended
	}

	// UI 交互（分支创建）
	if stats.UI != nil {
		score += stats.UI.BranchCreated
	}

	return score
}
