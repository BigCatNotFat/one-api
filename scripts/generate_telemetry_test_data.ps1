# 生成遥测测试数据脚本
# 向本地 one-api 服务发送模拟的遥测数据

$baseUrl = "http://localhost:3000/api/telemetry"

# 模型列表
$models = @("gpt-4", "gpt-4o", "claude-3-opus", "claude-3-sonnet", "gemini-pro")
# 模式列表
$modes = @("agent", "chat", "normal")
# 工具列表
$tools = @("codebase_search", "read_file", "write", "grep", "run_terminal_cmd", "list_dir", "web_search")
# 文本操作列表
$textActions = @("polish", "expand", "condense", "translate", "custom")

# 生成随机 UUID
function New-RandomGuid {
    return [guid]::NewGuid().ToString()
}

# 生成随机用户ID (模拟10个不同用户)
$userIds = @()
for ($i = 0; $i -lt 10; $i++) {
    $userIds += New-RandomGuid
}

Write-Host "开始生成测试数据..." -ForegroundColor Cyan
Write-Host "目标地址: $baseUrl" -ForegroundColor Gray

# 生成过去24小时的数据，每半小时一批
$now = Get-Date
$hoursBack = 24
$totalRequests = 0
$successRequests = 0

for ($h = $hoursBack; $h -ge 0; $h--) {
    for ($m = 0; $m -lt 60; $m += 30) {
        # 计算时间戳
        $timestamp = $now.AddHours(-$h).AddMinutes(-$m)
        $timestampMs = [long](($timestamp.ToUniversalTime() - [datetime]'1970-01-01').TotalMilliseconds)
        
        # 每个时间段随机生成1-5个事件
        $eventCount = Get-Random -Minimum 1 -Maximum 6
        
        for ($e = 0; $e -lt $eventCount; $e++) {
            # 随机选择用户
            $userId = $userIds | Get-Random
            $sessionId = New-RandomGuid
            
            # 构建统计数据
            $chatData = @{}
            # 随机添加1-3个模型的聊天记录
            $modelCount = Get-Random -Minimum 1 -Maximum 4
            for ($mc = 0; $mc -lt $modelCount; $mc++) {
                $model = $models | Get-Random
                $mode = $modes | Get-Random
                if (-not $chatData.ContainsKey($model)) {
                    $chatData[$model] = @{}
                }
                $chatData[$model][$mode] = Get-Random -Minimum 1 -Maximum 10
            }
            
            # 工具使用统计
            $toolsData = @{}
            $toolCount = Get-Random -Minimum 0 -Maximum 5
            for ($tc = 0; $tc -lt $toolCount; $tc++) {
                $tool = $tools | Get-Random
                if (-not $toolsData.ContainsKey($tool)) {
                    $toolsData[$tool] = @{
                        success = Get-Random -Minimum 0 -Maximum 20
                        failed = Get-Random -Minimum 0 -Maximum 3
                    }
                }
            }
            
            # 工具审批统计
            $toolApprovalData = @{}
            if ((Get-Random -Minimum 0 -Maximum 2) -eq 1) {
                $tool = $tools | Get-Random
                $toolApprovalData[$tool] = @{
                    approved = Get-Random -Minimum 1 -Maximum 10
                    rejected = Get-Random -Minimum 0 -Maximum 3
                }
            }
            
            # 文本操作统计
            $textActionsData = @{}
            if ((Get-Random -Minimum 0 -Maximum 2) -eq 1) {
                $action = $textActions | Get-Random
                $textActionsData[$action] = @{
                    used = Get-Random -Minimum 1 -Maximum 5
                    accepted = Get-Random -Minimum 0 -Maximum 5
                    rejected = Get-Random -Minimum 0 -Maximum 2
                }
            }
            
            # 构建请求体
            $payload = @{
                userId = $userId
                sessionId = $sessionId
                version = "1.0.0"
                timestamp = $timestampMs
                statistics = @{
                    chat = $chatData
                    tools = $toolsData
                    toolApproval = $toolApprovalData
                    textActions = $textActionsData
                    session = @{
                        started = Get-Random -Minimum 0 -Maximum 2
                        ended = Get-Random -Minimum 0 -Maximum 2
                    }
                    ui = @{
                        branchCreated = Get-Random -Minimum 0 -Maximum 3
                        paneCount = Get-Random -Minimum 1 -Maximum 4
                    }
                }
            }
            
            $jsonBody = $payload | ConvertTo-Json -Depth 10
            
            try {
                $response = Invoke-RestMethod -Uri $baseUrl -Method Post -Body $jsonBody -ContentType "application/json" -ErrorAction Stop
                if ($response.success) {
                    $successRequests++
                }
            } catch {
                Write-Host "请求失败: $_" -ForegroundColor Red
            }
            
            $totalRequests++
        }
        
        # 显示进度
        $timeStr = $timestamp.ToString("MM-dd HH:mm")
        Write-Host "`r已发送 $totalRequests 个请求 (成功: $successRequests) - 当前时间段: $timeStr" -NoNewline
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "数据生成完成!" -ForegroundColor Green
Write-Host "总请求数: $totalRequests" -ForegroundColor Cyan
Write-Host "成功请求: $successRequests" -ForegroundColor Cyan
Write-Host "模拟用户数: $($userIds.Count)" -ForegroundColor Cyan
Write-Host "时间范围: 过去 $hoursBack 小时" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "现在可以访问 http://localhost:3000 查看遥测统计页面了!" -ForegroundColor Yellow

