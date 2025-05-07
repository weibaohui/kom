package deployment

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

// RolloutHistoryDeploymentTool 创建一个用于查询指定Kubernetes集群中Deployment升级历史的工具。
func RolloutHistoryDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_deployment_rollout_history",
		mcp.WithDescription("查询Deployment的升级历史。对应kubectl命令: kubectl rollout history deployment/<name> -n <namespace> / Query deployment rollout history. Equivalent kubectl command: kubectl rollout history deployment/<name> -n <namespace>"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RolloutHistoryDeploymentHandler 查询指定 Kubernetes 集群中 Deployment 的升级历史记录。
func RolloutHistoryDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	klog.Infof("Getting rollout history for deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	result, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Rollout().History()
	if err != nil {
		return nil, err
	}

	return tools.TextResult(result, meta)
}

// RolloutUndoDeploymentTool 创建一个用于回滚 Kubernetes Deployment 的工具。
// 支持指定目标版本号进行回滚，若未指定则回滚到上一个版本。
func RolloutUndoDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"undo_k8s_deployment_rollout",
		mcp.WithDescription("回滚Deployment到指定版本，如果不指定版本则回滚到上一个版本 / Rollback deployment to specific revision, or previous revision if not specified"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
		mcp.WithNumber("revision", mcp.Description("回滚到的版本号，可选 / Target revision number, optional")),
	)
}

// RolloutUndoDeploymentHandler 处理将指定 Kubernetes Deployment 回滚到特定修订版本（或上一个版本）的请求，并返回回滚结果文本。
func RolloutUndoDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	var revision int
	if revisionVal, ok := request.Params.Arguments["revision"].(float64); ok {
		revision = int(revisionVal)
	}

	klog.Infof("Rolling back deployment %s/%s in cluster %s to revision %s", meta.Namespace, meta.Name, meta.Cluster, revision)

	result, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Rollout().Undo(revision)
	if err != nil {
		return nil, err
	}

	return tools.TextResult(result, meta)
}

// RolloutPauseDeploymentTool 返回一个用于暂停指定Kubernetes Deployment升级过程的工具。
func RolloutPauseDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"pause_k8s_deployment_rollout",
		mcp.WithDescription("暂停Deployment的升级过程 / Pause deployment rollout"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RolloutPauseDeploymentHandler 处理暂停指定 Kubernetes Deployment 升级的请求，并返回操作结果。
func RolloutPauseDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	klog.Infof("Pausing rollout for deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Rollout().Pause()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully paused deployment rollout", meta)
}

// RolloutResumeDeploymentTool 创建一个用于恢复Kubernetes Deployment升级过程的工具。
func RolloutResumeDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"resume_k8s_deployment_rollout",
		mcp.WithDescription("恢复Deployment的升级过程 / Resume deployment rollout"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RolloutResumeDeploymentHandler 恢复指定 Kubernetes 集群中 Deployment 的升级进程。
// 
// 根据请求参数定位目标 Deployment，并执行 rollout resume 操作，使暂停的升级流程继续进行。
// 成功时返回操作结果文本，失败时返回错误。
func RolloutResumeDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	klog.Infof("Resuming rollout for deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Rollout().Resume()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully resumed deployment rollout", meta)
}

// RolloutStatusDeploymentTool 创建一个用于查询指定 Kubernetes Deployment 升级状态的工具。
func RolloutStatusDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_deployment_rollout_status",
		mcp.WithDescription("查询Deployment的升级状态 / Query deployment rollout status"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RolloutStatusDeploymentHandler 查询指定 Kubernetes Deployment 的升级（rollout）状态，并返回状态信息。
func RolloutStatusDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	klog.Infof("Getting rollout status for deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	result, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Rollout().Status()
	if err != nil {
		return nil, err
	}

	return tools.TextResult(result, meta)
}
