package deployment

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

// RolloutHistoryDeploymentTool 创建一个查询Deployment升级历史的工具
func RolloutHistoryDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"get_deployment_rollout_history",
		mcp.WithDescription("查询Deployment的升级历史。对应kubectl命令: kubectl rollout history deployment/<name> -n <namespace> / Query deployment rollout history. Equivalent kubectl command: kubectl rollout history deployment/<name> -n <namespace>"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RolloutHistoryDeploymentHandler 处理查询Deployment升级历史的请求
func RolloutHistoryDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
	}

	klog.Infof("Getting rollout history for deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	result, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Rollout().History()
	if err != nil {
		return nil, err
	}

	return utils.TextResult(result, meta)
}

// RolloutUndoDeploymentTool 创建一个回滚Deployment的工具
func RolloutUndoDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"undo_deployment_rollout",
		mcp.WithDescription("回滚Deployment到指定版本，如果不指定版本则回滚到上一个版本 / Rollback deployment to specific revision, or previous revision if not specified"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
		mcp.WithNumber("revision", mcp.Description("回滚到的版本号，可选 / Target revision number, optional")),
	)
}

// RolloutUndoDeploymentHandler 处理回滚Deployment的请求
func RolloutUndoDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
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

	return utils.TextResult(result, meta)
}

// RolloutPauseDeploymentTool 创建一个暂停Deployment升级的工具
func RolloutPauseDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"pause_deployment_rollout",
		mcp.WithDescription("暂停Deployment的升级过程 / Pause deployment rollout"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RolloutPauseDeploymentHandler 处理暂停Deployment升级的请求
func RolloutPauseDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
	}

	klog.Infof("Pausing rollout for deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Rollout().Pause()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully paused deployment rollout", meta)
}

// RolloutResumeDeploymentTool 创建一个恢复Deployment升级的工具
func RolloutResumeDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"resume_deployment_rollout",
		mcp.WithDescription("恢复Deployment的升级过程 / Resume deployment rollout"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RolloutResumeDeploymentHandler 处理恢复Deployment升级的请求
func RolloutResumeDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
	}

	klog.Infof("Resuming rollout for deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Rollout().Resume()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully resumed deployment rollout", meta)
}

// RolloutStatusDeploymentTool 创建一个查询Deployment升级状态的工具
func RolloutStatusDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"get_deployment_rollout_status",
		mcp.WithDescription("查询Deployment的升级状态 / Query deployment rollout status"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RolloutStatusDeploymentHandler 处理查询Deployment升级状态的请求
func RolloutStatusDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
	}

	klog.Infof("Getting rollout status for deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	result, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Rollout().Status()
	if err != nil {
		return nil, err
	}

	return utils.TextResult(result, meta)
}
