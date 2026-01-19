package model

import (
	"time"

	"github.com/songquanpeng/one-api/common"
	"gorm.io/gorm"
)

// TelemetryEvent 遥测事件主表
type TelemetryEvent struct {
	Id              int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId          string `json:"user_id" gorm:"type:varchar(36);index;not null"`           // 用户匿名 ID (UUID)
	SessionId       string `json:"session_id" gorm:"type:varchar(36);index;not null"`        // 会话 ID (UUID)
	Version         string `json:"version" gorm:"type:varchar(20)"`                          // 插件版本号
	ClientTimestamp int64  `json:"client_timestamp" gorm:"bigint"`                           // 客户端时间戳（毫秒）
	ServerTimestamp int64  `json:"server_timestamp" gorm:"bigint;index"`                     // 服务器接收时间戳
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (TelemetryEvent) TableName() string {
	return "telemetry_events"
}

// TelemetryChat 聊天统计表
type TelemetryChat struct {
	Id      int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	EventId int64  `json:"event_id" gorm:"index;not null"`
	ModelId string `json:"model_id" gorm:"type:varchar(100);index;not null"` // 模型 ID
	Mode    string `json:"mode" gorm:"type:varchar(20);not null"`            // 对话模式: agent/chat/normal
	Count   int    `json:"count" gorm:"default:0"`                           // 次数
}

func (TelemetryChat) TableName() string {
	return "telemetry_chats"
}

// TelemetryTool 工具执行统计表
type TelemetryTool struct {
	Id           int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	EventId      int64  `json:"event_id" gorm:"index;not null"`
	ToolName     string `json:"tool_name" gorm:"type:varchar(100);index;not null"` // 工具名称
	SuccessCount int    `json:"success_count" gorm:"default:0"`                    // 成功次数
	FailedCount  int    `json:"failed_count" gorm:"default:0"`                     // 失败次数
}

func (TelemetryTool) TableName() string {
	return "telemetry_tools"
}

// TelemetryToolApproval 工具审批统计表
type TelemetryToolApproval struct {
	Id            int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	EventId       int64  `json:"event_id" gorm:"index;not null"`
	ToolName      string `json:"tool_name" gorm:"type:varchar(100);index;not null"` // 工具名称
	ApprovedCount int    `json:"approved_count" gorm:"default:0"`                   // 批准次数
	RejectedCount int    `json:"rejected_count" gorm:"default:0"`                   // 拒绝次数
}

func (TelemetryToolApproval) TableName() string {
	return "telemetry_tool_approvals"
}

// TelemetryTextAction 文本操作统计表
type TelemetryTextAction struct {
	Id            int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	EventId       int64  `json:"event_id" gorm:"index;not null"`
	Action        string `json:"action" gorm:"type:varchar(50);index;not null"` // 操作类型: polish/expand/condense/translate/custom
	UsedCount     int    `json:"used_count" gorm:"default:0"`                   // 触发次数
	AcceptedCount int    `json:"accepted_count" gorm:"default:0"`               // 接受次数
	RejectedCount int    `json:"rejected_count" gorm:"default:0"`               // 拒绝次数
}

func (TelemetryTextAction) TableName() string {
	return "telemetry_text_actions"
}

// TelemetrySession 会话统计表
type TelemetrySession struct {
	Id           int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	EventId      int64 `json:"event_id" gorm:"index;not null"`
	StartedCount int   `json:"started_count" gorm:"default:0"` // 会话开始次数
	EndedCount   int   `json:"ended_count" gorm:"default:0"`   // 会话结束次数
}

func (TelemetrySession) TableName() string {
	return "telemetry_sessions"
}

// TelemetryUI UI 交互统计表
type TelemetryUI struct {
	Id                  int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	EventId             int64 `json:"event_id" gorm:"index;not null"`
	BranchCreatedCount  int   `json:"branch_created_count" gorm:"default:0"` // 新建分支次数
	PaneCount           int   `json:"pane_count" gorm:"default:0"`           // 对话列数量
}

func (TelemetryUI) TableName() string {
	return "telemetry_uis"
}

// ========== 数据库操作方法 ==========

// CreateTelemetryEvent 创建遥测事件
func CreateTelemetryEvent(event *TelemetryEvent) error {
	return DB.Create(event).Error
}

// CreateTelemetryChat 批量创建聊天统计
func CreateTelemetryChats(chats []*TelemetryChat) error {
	if len(chats) == 0 {
		return nil
	}
	return DB.Create(&chats).Error
}

// CreateTelemetryTools 批量创建工具统计
func CreateTelemetryTools(tools []*TelemetryTool) error {
	if len(tools) == 0 {
		return nil
	}
	return DB.Create(&tools).Error
}

// CreateTelemetryToolApprovals 批量创建工具审批统计
func CreateTelemetryToolApprovals(approvals []*TelemetryToolApproval) error {
	if len(approvals) == 0 {
		return nil
	}
	return DB.Create(&approvals).Error
}

// CreateTelemetryTextActions 批量创建文本操作统计
func CreateTelemetryTextActions(actions []*TelemetryTextAction) error {
	if len(actions) == 0 {
		return nil
	}
	return DB.Create(&actions).Error
}

// CreateTelemetrySessions 创建会话统计
func CreateTelemetrySession(session *TelemetrySession) error {
	if session == nil {
		return nil
	}
	return DB.Create(session).Error
}

// CreateTelemetryUI 创建 UI 统计
func CreateTelemetryUI(ui *TelemetryUI) error {
	if ui == nil {
		return nil
	}
	return DB.Create(ui).Error
}

// ========== 查询方法 ==========

// GetTelemetryEvents 获取遥测事件列表
func GetTelemetryEvents(startTimestamp, endTimestamp int64, startIdx, num int) ([]*TelemetryEvent, error) {
	var events []*TelemetryEvent
	tx := DB.Model(&TelemetryEvent{})
	if startTimestamp > 0 {
		tx = tx.Where("server_timestamp >= ?", startTimestamp)
	}
	if endTimestamp > 0 {
		tx = tx.Where("server_timestamp <= ?", endTimestamp)
	}
	err := tx.Order("id desc").Limit(num).Offset(startIdx).Find(&events).Error
	return events, err
}

// TelemetryOverview 概览统计结构
type TelemetryOverview struct {
	TotalEvents   int64 `json:"total_events"`
	TotalUsers    int64 `json:"total_users"`
	TotalSessions int64 `json:"total_sessions"`
	TodayEvents   int64 `json:"today_events"`
	TodayUsers    int64 `json:"today_users"`
}

// GetTelemetryOverview 获取遥测概览统计
func GetTelemetryOverview() (*TelemetryOverview, error) {
	overview := &TelemetryOverview{}
	
	// 总事件数
	DB.Model(&TelemetryEvent{}).Count(&overview.TotalEvents)
	
	// 总用户数（去重）
	DB.Model(&TelemetryEvent{}).Distinct("user_id").Count(&overview.TotalUsers)
	
	// 总会话数（去重）
	DB.Model(&TelemetryEvent{}).Distinct("session_id").Count(&overview.TotalSessions)
	
	// 今日统计
	todayStart := time.Now().Truncate(24 * time.Hour).UnixMilli()
	DB.Model(&TelemetryEvent{}).Where("server_timestamp >= ?", todayStart).Count(&overview.TodayEvents)
	DB.Model(&TelemetryEvent{}).Where("server_timestamp >= ?", todayStart).Distinct("user_id").Count(&overview.TodayUsers)
	
	return overview, nil
}

// TelemetryChatStat 聊天统计结构
type TelemetryChatStat struct {
	ModelId    string `json:"model_id"`
	Mode       string `json:"mode"`
	TotalCount int64  `json:"total_count"`
}

// GetTelemetryChatStats 获取聊天统计
func GetTelemetryChatStats(startTimestamp, endTimestamp int64) ([]*TelemetryChatStat, error) {
	var stats []*TelemetryChatStat
	tx := DB.Model(&TelemetryChat{}).
		Select("model_id, mode, SUM(count) as total_count").
		Group("model_id, mode").
		Order("total_count desc")
	
	if startTimestamp > 0 || endTimestamp > 0 {
		tx = tx.Joins("JOIN telemetry_events ON telemetry_chats.event_id = telemetry_events.id")
		if startTimestamp > 0 {
			tx = tx.Where("telemetry_events.server_timestamp >= ?", startTimestamp)
		}
		if endTimestamp > 0 {
			tx = tx.Where("telemetry_events.server_timestamp <= ?", endTimestamp)
		}
	}
	
	err := tx.Find(&stats).Error
	return stats, err
}

// TelemetryToolStat 工具统计结构
type TelemetryToolStat struct {
	ToolName     string `json:"tool_name"`
	SuccessCount int64  `json:"success_count"`
	FailedCount  int64  `json:"failed_count"`
}

// GetTelemetryToolStats 获取工具统计
func GetTelemetryToolStats(startTimestamp, endTimestamp int64) ([]*TelemetryToolStat, error) {
	var stats []*TelemetryToolStat
	tx := DB.Model(&TelemetryTool{}).
		Select("tool_name, SUM(success_count) as success_count, SUM(failed_count) as failed_count").
		Group("tool_name").
		Order("success_count desc")
	
	if startTimestamp > 0 || endTimestamp > 0 {
		tx = tx.Joins("JOIN telemetry_events ON telemetry_tools.event_id = telemetry_events.id")
		if startTimestamp > 0 {
			tx = tx.Where("telemetry_events.server_timestamp >= ?", startTimestamp)
		}
		if endTimestamp > 0 {
			tx = tx.Where("telemetry_events.server_timestamp <= ?", endTimestamp)
		}
	}
	
	err := tx.Find(&stats).Error
	return stats, err
}

// TelemetryDAUStat 日活统计结构
type TelemetryDAUStat struct {
	Date      string `json:"date"`
	UserCount int64  `json:"user_count"`
}

// GetTelemetryDAUStats 获取日活统计（最近 N 天）
func GetTelemetryDAUStats(days int) ([]*TelemetryDAUStat, error) {
	var stats []*TelemetryDAUStat
	
	// 计算起始时间
	startTime := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour).UnixMilli()
	
	// 根据数据库类型选择日期格式化函数
	dateFormat := "DATE(FROM_UNIXTIME(server_timestamp/1000))"
	if common.UsingPostgreSQL {
		dateFormat = "DATE(TO_TIMESTAMP(server_timestamp/1000))"
	}
	if common.UsingSQLite {
		dateFormat = "DATE(server_timestamp/1000, 'unixepoch')"
	}
	
	err := DB.Model(&TelemetryEvent{}).
		Select(dateFormat + " as date, COUNT(DISTINCT user_id) as user_count").
		Where("server_timestamp >= ?", startTime).
		Group("date").
		Order("date asc").
		Find(&stats).Error
	
	return stats, err
}

// TelemetryTextActionStat 文本操作统计结构
type TelemetryTextActionStat struct {
	Action        string `json:"action"`
	UsedCount     int64  `json:"used_count"`
	AcceptedCount int64  `json:"accepted_count"`
	RejectedCount int64  `json:"rejected_count"`
}

// GetTelemetryTextActionStats 获取文本操作统计
func GetTelemetryTextActionStats(startTimestamp, endTimestamp int64) ([]*TelemetryTextActionStat, error) {
	var stats []*TelemetryTextActionStat
	tx := DB.Model(&TelemetryTextAction{}).
		Select("action, SUM(used_count) as used_count, SUM(accepted_count) as accepted_count, SUM(rejected_count) as rejected_count").
		Group("action").
		Order("used_count desc")
	
	if startTimestamp > 0 || endTimestamp > 0 {
		tx = tx.Joins("JOIN telemetry_events ON telemetry_text_actions.event_id = telemetry_events.id")
		if startTimestamp > 0 {
			tx = tx.Where("telemetry_events.server_timestamp >= ?", startTimestamp)
		}
		if endTimestamp > 0 {
			tx = tx.Where("telemetry_events.server_timestamp <= ?", endTimestamp)
		}
	}
	
	err := tx.Find(&stats).Error
	return stats, err
}

// TelemetryToolApprovalStat 工具审批统计结构
type TelemetryToolApprovalStat struct {
	ToolName      string `json:"tool_name"`
	ApprovedCount int64  `json:"approved_count"`
	RejectedCount int64  `json:"rejected_count"`
}

// GetTelemetryToolApprovalStats 获取工具审批统计
func GetTelemetryToolApprovalStats(startTimestamp, endTimestamp int64) ([]*TelemetryToolApprovalStat, error) {
	var stats []*TelemetryToolApprovalStat
	tx := DB.Model(&TelemetryToolApproval{}).
		Select("tool_name, SUM(approved_count) as approved_count, SUM(rejected_count) as rejected_count").
		Group("tool_name").
		Order("approved_count desc")
	
	if startTimestamp > 0 || endTimestamp > 0 {
		tx = tx.Joins("JOIN telemetry_events ON telemetry_tool_approvals.event_id = telemetry_events.id")
		if startTimestamp > 0 {
			tx = tx.Where("telemetry_events.server_timestamp >= ?", startTimestamp)
		}
		if endTimestamp > 0 {
			tx = tx.Where("telemetry_events.server_timestamp <= ?", endTimestamp)
		}
	}
	
	err := tx.Find(&stats).Error
	return stats, err
}

// MigrateTelemetryTables 迁移遥测相关表
func MigrateTelemetryTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&TelemetryEvent{},
		&TelemetryChat{},
		&TelemetryTool{},
		&TelemetryToolApproval{},
		&TelemetryTextAction{},
		&TelemetrySession{},
		&TelemetryUI{},
	)
}

