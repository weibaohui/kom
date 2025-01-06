package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func RegisterInit() {
	klog.Infof("RegisterInit")
	kom.Clusters().SetRegisterCallbackFunc(RegisterDefaultCallbacks)
	klog.Infof("Register RegisterDefaultCallbacks func  to clusters")
}
func init() {
	RegisterInit()
}
func RegisterDefaultCallbacks(c *kom.ClusterInst) func() {

	klog.V(4).Infof("RegisterDefaultCallbacks for cluster %s", c.ID)
	prefix := c.ID

	// 为每一个集群进行注册
	k := c.Kubectl

	queryCallback := k.Callback().Get()
	_ = queryCallback.Register(fmt.Sprintf("%s:get", prefix), Get)

	listCallback := k.Callback().List()
	_ = listCallback.Register(fmt.Sprintf("%s:list", prefix), List)

	watchCallback := k.Callback().Watch()
	_ = watchCallback.Register(fmt.Sprintf("%s:watch", prefix), Watch)

	createCallback := k.Callback().Create()
	_ = createCallback.Register(fmt.Sprintf("%s:create", prefix), Create)

	updateCallback := k.Callback().Update()
	_ = updateCallback.Register(fmt.Sprintf("%s:update", prefix), Update)

	patchCallback := k.Callback().Patch()
	_ = patchCallback.Register(fmt.Sprintf("%s:patch", prefix), Patch)

	deleteCallback := k.Callback().Delete()
	_ = deleteCallback.Register(fmt.Sprintf("%s:delete", prefix), Delete)

	execCallback := k.Callback().Exec()
	_ = execCallback.Register(fmt.Sprintf("%s:pod:exec", prefix), ExecuteCommand)

	streamExecCallback := k.Callback().StreamExec()
	_ = streamExecCallback.Register(fmt.Sprintf("%s:pod:stream:exec", prefix), StreamExecuteCommand)

	logsCallback := k.Callback().Logs()
	_ = logsCallback.Register(fmt.Sprintf("%s:pod:logs", prefix), GetLogs)

	describeCallback := k.Callback().Describe()
	_ = describeCallback.Register(fmt.Sprintf("%s:describe", prefix), Describe)

	return nil
}
