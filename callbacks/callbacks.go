package callbacks

import (
	"github.com/weibaohui/kom/kom"
)

func RegisterDefaultCallbacks() {

	queryCallback := kom.Init().Callback().Get()
	_ = queryCallback.Register("k8sorm:get", Get)

	listCallback := kom.Init().Callback().List()
	_ = listCallback.Register("k8sorm:list", List)

	createCallback := kom.Init().Callback().Create()
	_ = createCallback.Register("k8sorm:create", Create)

	updateCallback := kom.Init().Callback().Update()
	_ = updateCallback.Register("k8sorm:update", Update)

	patchCallback := kom.Init().Callback().Patch()
	_ = patchCallback.Register("k8sorm:patch", Patch)

	deleteCallback := kom.Init().Callback().Delete()
	_ = deleteCallback.Register("k8sorm:delete", Delete)
}
