package polymorphichelpers

import (
	"context"

	"k8s.io/client-go/kubernetes"

	"github.com/xmapst/AutoExecFlow/internal/runner/kubernetes/polymorphichelpers/daemonset"
	"github.com/xmapst/AutoExecFlow/internal/runner/kubernetes/polymorphichelpers/deployment"
	"github.com/xmapst/AutoExecFlow/internal/runner/kubernetes/polymorphichelpers/statefulset"
	"github.com/xmapst/AutoExecFlow/internal/runner/kubernetes/polymorphichelpers/types"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
)

type Resource interface {
	Scale(replicas int32) error
	Update() error
	Println() error
	Restart() error
}

func ResourceFor(ctx context.Context, storage backend.IStep, client *kubernetes.Clientset, resource *types.Resource) Resource {
	var rs Resource
	switch resource.GetKind() {
	case types.Deployment:
		rs = &deployment.Deployment{
			Context:  ctx,
			Client:   client.AppsV1().Deployments(resource.GetNamespace()),
			Resource: resource,
			Storage:  storage,
		}
	case types.DaemonSet:
		rs = &daemonset.DaemonSet{
			Context:  ctx,
			Client:   client.AppsV1().DaemonSets(resource.GetNamespace()),
			Resource: resource,
			Storage:  storage,
		}
	case types.StatefulSet:
		rs = &statefulset.StatefulSet{
			Context:  ctx,
			Client:   client.AppsV1().StatefulSets(resource.GetNamespace()),
			Resource: resource,
			Storage:  storage,
		}
	}
	return rs
}
