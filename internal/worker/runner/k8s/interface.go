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

type IResource interface {
	Scale(replicas int32) error
	Update() error
	Println() error
	Restart() error
}

func ResourceFor(ctx context.Context, storage storage.IStep, client *kubernetes.Clientset, resource *types.SResource) IResource {
	var rs IResource
	switch resource.GetKind() {
	case types.Deployment:
		rs = &deploy.SDeployment{
			Context:   ctx,
			Client:    client.AppsV1().Deployments(resource.GetNamespace()),
			SResource: resource,
			Storage:   storage,
		}
	case types.DaemonSet:
		rs = &ds.SDaemonSet{
			Context:   ctx,
			Client:    client.AppsV1().DaemonSets(resource.GetNamespace()),
			SResource: resource,
			Storage:   storage,
		}
	case types.StatefulSet:
		rs = &sts.SStatefulSet{
			Context:   ctx,
			Client:    client.AppsV1().StatefulSets(resource.GetNamespace()),
			SResource: resource,
			Storage:   storage,
		}
	}
	return rs
}
