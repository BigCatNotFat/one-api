package model

import (
	"time"

	"github.com/songquanpeng/one-api/common"
	"gorm.io/gorm"
)

// TelemetryEvent 遥测事件主表
type TelemetryEvent struct {
	Id              int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId          string    `json:"user_id" gorm:"type:varchar(36);index;not null"`  // 用户匿名 ID (UUID)
	SessionId       string    `json:"session_id" gorm:"type:varchar(36);index;not null"` // 会话 ID (UUID)
	Version         string    `json:"version" gorm:"type:varchar(20)"`                 // 插件版本号
	ClientTimestamp int64     `json:"client_timestamp" gorm:"bigint"`                  // 客户端时间戳（毫秒）
	ServerTimestamp int64     `json:"server_timestamp" gorm:"bigint;index"`            // 服务器接收时间戳
	UsageScore      int       `json:"usage_score" gorm:"default:0"`                    // 使用度分数
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

// TelemetryTimelineStat 时间轴统计结构（每半小时一个数据点）
type TelemetryTimelineStat struct {
	TimeSlot         string  `json:"time_slot"`          // 时间段标识 (格式: "2024-01-15 14:00" 或 "2024-01-15 14:30")
	Timestamp        int64   `json:"timestamp"`          // 时间段起始时间戳（毫秒）
	ActiveUsers      int64   `json:"active_users"`       // 活跃用户数
	EventCount       int64   `json:"event_count"`        // 事件数量
	SessionCount     int64   `json:"session_count"`      // 会话数量
	ChatCount        int64   `json:"chat_count"`         // 聊天次数
	ToolSuccessCount int64   `json:"tool_success_count"` // 工具成功次数
	ToolFailedCount  int64   `json:"tool_failed_count"`  // 工具失败次数
	AvgUsageScore    float64 `json:"avg_usage_score"`    // 平均使用度分数
}

// GetTelemetryTimelineStats 获取时间轴统计（按半小时分组，最近24小时）
func GetTelemetryTimelineStats(hours int) ([]*TelemetryTimelineStat, error) {
	if hours <= 0 {
		hours = 24 // 默认24小时
	}
	if hours > 168 {
		hours = 168 // 最多7天
	}

	// 计算时间范围
	now := time.Now()
	// 对齐到下一个半小时边界（这样才能包含当前时间段的数据）
	minutes := now.Minute()
	if minutes >= 30 {
		// 当前在 xx:30-xx:59，对齐到下一个整点
		now = now.Truncate(time.Hour).Add(time.Hour)
	} else {
		// 当前在 xx:00-xx:29，对齐到 xx:30
		now = now.Truncate(time.Hour).Add(30 * time.Minute)
	}
	endTime := now.UnixMilli()
	startTime := now.Add(time.Duration(-hours) * time.Hour).UnixMilli()

	// 生成所有时间段
	timeSlots := make(map[int64]*TelemetryTimelineStat)
	slotOrder := make([]int64, 0)
	for ts := startTime; ts < endTime; ts += 30 * 60 * 1000 { // 每30分钟
		t := time.UnixMilli(ts)
		timeSlots[ts] = &TelemetryTimelineStat{
			TimeSlot:  t.Format("2006-01-02 15:04"),
			Timestamp: ts,
		}
		slotOrder = append(slotOrder, ts)
	}

	// 根据数据库类型选择时间槽计算方式
	// 使用 client_timestamp（客户端上报时间）来统计，这样更能反映用户实际活动时间
	var timeSlotExpr string
	if common.UsingPostgreSQL {
		// PostgreSQL: 将时间戳对齐到30分钟
		timeSlotExpr = "FLOOR(client_timestamp / 1800000) * 1800000"
	} else if common.UsingSQLite {
		// SQLite
		timeSlotExpr = "(client_timestamp / 1800000) * 1800000"
	} else {
		// MySQL
		timeSlotExpr = "FLOOR(client_timestamp / 1800000) * 1800000"
	}

	// 查询事件统计
	type EventStat struct {
		TimeSlot      int64   `gorm:"column:time_slot"`
		ActiveUsers   int64   `gorm:"column:active_users"`
		EventCount    int64   `gorm:"column:event_count"`
		SessionCount  int64   `gorm:"column:session_count"`
		AvgUsageScore float64 `gorm:"column:avg_usage_score"`
	}
	var eventStats []EventStat
	err := DB.Model(&TelemetryEvent{}).
		Select(timeSlotExpr+" as time_slot, COUNT(DISTINCT user_id) as active_users, COUNT(*) as event_count, COUNT(DISTINCT session_id) as session_count, AVG(usage_score) as avg_usage_score").
		Where("client_timestamp >= ? AND client_timestamp < ?", startTime, endTime).
		Group("time_slot").
		Find(&eventStats).Error
	if err != nil {
		return nil, err
	}

	// 合并事件统计
	for _, stat := range eventStats {
		if slot, ok := timeSlots[stat.TimeSlot]; ok {
			slot.ActiveUsers = stat.ActiveUsers
			slot.EventCount = stat.EventCount
			slot.SessionCount = stat.SessionCount
			slot.AvgUsageScore = stat.AvgUsageScore
		}
	}

	// 查询聊天统计
	type ChatStat struct {
		TimeSlot  int64 `gorm:"column:time_slot"`
		ChatCount int64 `gorm:"column:chat_count"`
	}
	var chatStats []ChatStat
	err = DB.Model(&TelemetryChat{}).
		Select(timeSlotExpr+" as time_slot, SUM(count) as chat_count").
		Joins("JOIN telemetry_events ON telemetry_chats.event_id = telemetry_events.id").
		Where("telemetry_events.client_timestamp >= ? AND telemetry_events.client_timestamp < ?", startTime, endTime).
		Group("time_slot").
		Find(&chatStats).Error
	if err == nil {
		for _, stat := range chatStats {
			if slot, ok := timeSlots[stat.TimeSlot]; ok {
				slot.ChatCount = stat.ChatCount
			}
		}
	}

	// 查询工具统计
	type ToolStat struct {
		TimeSlot     int64 `gorm:"column:time_slot"`
		SuccessCount int64 `gorm:"column:success_count"`
		FailedCount  int64 `gorm:"column:failed_count"`
	}
	var toolStats []ToolStat
	err = DB.Model(&TelemetryTool{}).
		Select(timeSlotExpr+" as time_slot, SUM(success_count) as success_count, SUM(failed_count) as failed_count").
		Joins("JOIN telemetry_events ON telemetry_tools.event_id = telemetry_events.id").
		Where("telemetry_events.client_timestamp >= ? AND telemetry_events.client_timestamp < ?", startTime, endTime).
		Group("time_slot").
		Find(&toolStats).Error
	if err == nil {
		for _, stat := range toolStats {
			if slot, ok := timeSlots[stat.TimeSlot]; ok {
				slot.ToolSuccessCount = stat.SuccessCount
				slot.ToolFailedCount = stat.FailedCount
			}
		}
	}

	// 按时间顺序返回结果
	result := make([]*TelemetryTimelineStat, 0, len(slotOrder))
	for _, ts := range slotOrder {
		result = append(result, timeSlots[ts])
	}

	return result, nil
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

