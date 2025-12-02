# 问题修复说明

## 问题
访问 `http://localhost:3000/query` 显示空白页面

## 原因
`web/standalone-query.html` 文件是空的（0字节），导致 embed 无内容可返回

## 已修复
- ✅ 重新创建了 HTML 文件（8KB）
- ✅ 添加了错误日志到 `router/web.go`

## 部署步骤

### 1. 重新构建镜像
```bash
docker build -t xuzhen222/one-api-custom:latest .
```

### 2. 重启服务
```bash
docker-compose down
docker-compose up -d
```

### 3. 查看启动日志
```bash
docker logs one-api
```

**预期看到：**
```
[INFO] Standalone query page loaded successfully, size: 8051 bytes
```

### 4. 测试访问
```
http://localhost:3000/query
```

**预期结果：**
- 显示渐变紫色背景
- 显示"硅基之梦"Logo
- 显示查询输入框
- 可以输入 API Key 并查询

## 验证成功标志

1. 页面不再是空白
2. 可以看到完整的UI界面
3. 查询功能正常工作
4. Docker 日志中显示"Standalone query page loaded successfully"
