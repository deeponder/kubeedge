package cloudhub

import (
	"os"

	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/channelq"
	hubconfig "github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/config"
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/servers"
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/servers/httpserver"
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/servers/udsserver"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/client"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/informers"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/modules"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

var DoneTLSTunnelCerts = make(chan bool, 1)

type cloudHub struct {
	enable               bool
	informersSyncedFuncs []cache.InformerSynced
	messageq             *channelq.ChannelMessageQueue
}

var _ core.Module = (*cloudHub)(nil)

func newCloudHub(enable bool) *cloudHub {
	crdFactory := informers.GetInformersManager().GetCRDInformerFactory()
	// declare used informer
	clusterObjectSyncInformer := crdFactory.Reliablesyncs().V1alpha1().ClusterObjectSyncs()
	objectSyncInformer := crdFactory.Reliablesyncs().V1alpha1().ObjectSyncs()
	// 初始化ChannelMessageQueue， 每个Node都维护一个消息的队列
	messageq := channelq.NewChannelMessageQueue(objectSyncInformer.Lister(), clusterObjectSyncInformer.Lister(), client.GetCRDClient())
	ch := &cloudHub{
		enable:   enable,
		messageq: messageq,
	}
	ch.informersSyncedFuncs = append(ch.informersSyncedFuncs, clusterObjectSyncInformer.Informer().HasSynced)
	ch.informersSyncedFuncs = append(ch.informersSyncedFuncs, objectSyncInformer.Informer().HasSynced)
	return ch
}

func Register(hub *v1alpha1.CloudHub) {
	// 简单的参数校验 和 ca等证书的加载
	hubconfig.InitConfigure(hub)
	core.Register(newCloudHub(hub.Enable))
}

func (a *cloudHub) Name() string {
	return modules.CloudHubModuleName
}

func (a *cloudHub) Group() string {
	return modules.CloudHubModuleGroup
}

// Enable indicates whether enable this module
func (a *cloudHub) Enable() bool {
	return a.enable
}

func (a *cloudHub) Start() {
	if !cache.WaitForCacheSync(beehiveContext.Done(), a.informersSyncedFuncs...) {
		klog.Errorf("unable to sync caches for objectSyncController")
		os.Exit(1)
	}

	// start dispatch message from the cloud to edge node
	go a.messageq.DispatchMessage()

	// check whether the certificates exist in the local directory,
	// and then check whether certificates exist in the secret, generate if they don't exist
	// 创建相关的Cert Secret资源
	if err := httpserver.PrepareAllCerts(); err != nil {
		klog.Exit(err)
	}
	// TODO: Will improve in the future
	DoneTLSTunnelCerts <- true
	close(DoneTLSTunnelCerts)

	// generate Token
	// token用于edge连上来的的校验， 默认12小时有效期。 本质为JWT， 单独起一个协程，定时更新
	// 生成后，存入/更新 Secret资源, tokensecret
	if err := httpserver.GenerateToken(); err != nil {
		klog.Exit(err)
	}

	// HttpServer mainly used to issue certificates for the edge
	// 要么校验证书，要么校验上一步生成的jwt
	go httpserver.StartHTTPServer()

	servers.StartCloudHub(a.messageq)

	if hubconfig.Config.UnixSocket.Enable {
		// The uds server is only used to communicate with csi driver from kubeedge on cloud.
		// It is not used to communicate between cloud and edge.
		// cloud端的存储组件通信
		go udsserver.StartServer(hubconfig.Config.UnixSocket.Address)
	}
}
