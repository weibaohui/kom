package callbacks

import (
	"github.com/weibaohui/kom/kom"
)

func RegisterDefaultCallbacks(clusters *kom.ClusterInstances) {

	all := clusters.All()
	for _, c := range all {
		// 为每一个集群进行注册
		k := c.Kom

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
	}

}
