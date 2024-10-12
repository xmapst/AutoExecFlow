package k8s

import (
	"context"

	"k8s.io/client-go/kubernetes"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/k8s/deploy"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/k8s/ds"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/k8s/sts"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/k8s/types"
)

type Resource interface {
	Scale(replicas int32) error
	Update() error
	Println() error
	Restart() error
}

func ResourceFor(ctx context.Context, storage storage.IStep, client *kubernetes.Clientset, resource *types.Resource) Resource {
	var rs Resource
	switch resource.GetKind() {
	case types.Deployment:
		rs = &deploy.Deployment{
			Context:  ctx,
			Client:   client.AppsV1().Deployments(resource.GetNamespace()),
			Resource: resource,
			Storage:  storage,
		}
	case types.DaemonSet:
		rs = &ds.DaemonSet{
			Context:  ctx,
			Client:   client.AppsV1().DaemonSets(resource.GetNamespace()),
			Resource: resource,
			Storage:  storage,
		}
	case types.StatefulSet:
		rs = &sts.StatefulSet{
			Context:  ctx,
			Client:   client.AppsV1().StatefulSets(resource.GetNamespace()),
			Resource: resource,
			Storage:  storage,
		}
	}
	return rs
}
