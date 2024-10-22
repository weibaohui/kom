package callbacks

import (
	"github.com/weibaohui/kom/kom"
)

func RegisterDefaultCallbacks() {

	queryCallback := kom.Init().Callback().Get()
	_ = queryCallback.Register("kom:get", Get)

	listCallback := kom.Init().Callback().List()
	_ = listCallback.Register("kom:list", List)

	createCallback := kom.Init().Callback().Create()
	_ = createCallback.Register("kom:create", Create)

	updateCallback := kom.Init().Callback().Update()
	_ = updateCallback.Register("kom:update", Update)

	patchCallback := kom.Init().Callback().Patch()
	_ = patchCallback.Register("kom:patch", Patch)

	deleteCallback := kom.Init().Callback().Delete()
	_ = deleteCallback.Register("kom:delete", Delete)
}
