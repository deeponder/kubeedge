package devicecontroller

import (
	"time"

	"k8s.io/klog/v2"

	"github.com/kubeedge/beehive/pkg/core"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/informers"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/modules"
	"github.com/kubeedge/kubeedge/cloud/pkg/devicecontroller/config"
	"github.com/kubeedge/kubeedge/cloud/pkg/devicecontroller/controller"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

// DeviceController use beehive context message layer
type DeviceController struct {
	downstream *controller.DownstreamController
	upstream   *controller.UpstreamController
	enable     bool
}

var _ core.Module = (*DeviceController)(nil)

func newDeviceController(enable bool) *DeviceController {
	if !enable {
		return &DeviceController{enable: enable}
	}
	// 分为DeviceModelManager和DeviceManager, 和 edgeController类似
	downstream, err := controller.NewDownstreamController(informers.GetInformersManager().GetCRDInformerFactory())
	if err != nil {
		klog.Exitf("New downstream controller failed with error: %s", err)
	}
	// edge to apiServer比较简单， 只是消息的转发，调用k8s client api
	upstream, err := controller.NewUpstreamController(downstream)
	if err != nil {
		klog.Exitf("new upstream controller failed with error: %s", err)
	}
	return &DeviceController{
		downstream: downstream,
		upstream:   upstream,
		enable:     enable,
	}
}

func Register(dc *v1alpha1.DeviceController) {
	config.InitConfigure(dc)
	core.Register(newDeviceController(dc.Enable))
}

// Name of controller
func (dc *DeviceController) Name() string {
	return modules.DeviceControllerModuleName
}

// Group of controller
func (dc *DeviceController) Group() string {
	return modules.DeviceControllerModuleGroup
}

// Enable indicates whether enable this module
func (dc *DeviceController) Enable() bool {
	return dc.enable
}

// Start controller
func (dc *DeviceController) Start() {
	if err := dc.downstream.Start(); err != nil {
		klog.Exitf("start downstream failed with error: %s", err)
	}
	// wait for downstream controller to start and load deviceModels and devices
	// TODO think about sync   channel同步dc 是否ready？
	time.Sleep(1 * time.Second)
	if err := dc.upstream.Start(); err != nil {
		klog.Exitf("start upstream failed with error: %s", err)
	}
}
