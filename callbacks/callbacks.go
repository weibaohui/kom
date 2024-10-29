package callbacks

import (
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func init() {
	kom.Clusters().SetRegisterCallbackFunc(RegisterDefaultCallbacks)
	klog.Infof("Register RegisterDefaultCallbacks func  to clusters")
}
func RegisterDefaultCallbacks(clusters *kom.ClusterInstances) func() {

	all := clusters.AllClusters()
	for _, c := range all {
		// 为每一个集群进行注册
		k := c.Kubectl

		queryCallback := k.Callback().Get()
		_ = queryCallback.Register("kom:get", Get)

		listCallback := k.Callback().List()
		_ = listCallback.Register("kom:list", List)

		createCallback := k.Callback().Create()
		_ = createCallback.Register("kom:create", Create)

		updateCallback := k.Callback().Update()
		_ = updateCallback.Register("kom:update", Update)

		patchCallback := k.Callback().Patch()
		_ = patchCallback.Register("kom:patch", Patch)

		deleteCallback := k.Callback().Delete()
		_ = deleteCallback.Register("kom:delete", Delete)

		execCallback := k.Callback().Exec()
		_ = execCallback.Register("kom:pod:exec", ExecuteCommand)

		logsCallback := k.Callback().Logs()
		_ = logsCallback.Register("kom:pod:logs", GetLogs)
	}

	return nil
}
