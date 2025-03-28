# MCP 使用教程

## 启动服务

### 环境要求
- Go 1.16或更高版本
- 已配置好的Kubernetes集群
- 默认读取Kubeconfig文件

### 启动命令
```go
mcp.RunMCPServer("kom mcp server", "0.0.1", 3619)
```
 
## AI工具集成

### Claude Desktop
1. 打开Claude Desktop设置面板
2. 在API配置区域添加MCP Server地址
3. 启用SSE事件监听功能
4. 验证连接状态

### Cursor
1. 进入Cursor设置界面
2. 找到扩展服务配置选项
3. 添加MCP Server的URL（例如：http://localhost:3619/sse）
4. 开启实时事件通知

### Windsurf
1. 访问配置中心
2. 设置API服务器地址
3. 启用实时事件通知
4. 测试连接

### 常见问题
1. 确保MCP Server正常运行且端口可访问
2. 检查网络连接是否正常
3. 验证SSE连接是否成功建立
4. 查看工具日志以排查连接问题
 
## 支持

如有问题或建议，欢迎提交Issue或Pull Request。