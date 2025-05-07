package node

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// TaintNodeTool 返回一个用于为 Kubernetes 节点添加污点的工具，等同于 kubectl taint nodes 命令。
func TaintNodeTool() mcp.Tool {
	return mcp.NewTool(
		"taint_k8s_node",
		mcp.WithDescription("为节点添加污点，等同于kubectl taint nodes <node> <key>=<value>:<effect> / Add taint to node, equivalent to kubectl taint nodes <node> <key>=<value>:<effect>"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Required(), mcp.Description("节点名称 / The name of the node")),
		mcp.WithString("taint", mcp.Description("污点表达式，格式为key=value:effect / Taint expression in format key=value:effect")),
	)
}

// TaintNodeHandler 处理为指定 Kubernetes 节点添加污点的请求。
// 根据请求参数解析目标集群、节点名称和污点表达式，并将污点应用到对应节点。
// 成功时返回操作结果文本，失败时返回错误。
func TaintNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	taint := request.Params.Arguments["taint"].(string)

	klog.Infof("Adding taint %s to node %s in cluster %s", taint, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().Taint(taint)
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully added taint to node", meta)
}
