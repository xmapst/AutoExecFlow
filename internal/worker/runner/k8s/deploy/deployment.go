package deploy

import (
	"context"
	"fmt"
	"strings"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubetypes "k8s.io/apimachinery/pkg/types"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	kuberetry "k8s.io/client-go/util/retry"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/k8s/types"
)

type Deployment struct {
	context.Context
	Client appv1.DeploymentInterface
	*types.Resource
	Storage storage.IStep
}

func (d *Deployment) Restart() error {
	return kuberetry.RetryOnConflict(kuberetry.DefaultRetry, func() error {
		path := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().Format(time.RFC3339))
		_, err := d.Client.Patch(d.Context, d.GetName(), kubetypes.StrategicMergePatchType, []byte(path), metav1.PatchOptions{})
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *Deployment) Scale(replicas int32) error {
	return kuberetry.RetryOnConflict(kuberetry.DefaultRetry, func() error {
		scale := &autoscalingv1.Scale{
			Spec: autoscalingv1.ScaleSpec{
				Replicas: replicas,
			},
		}
		scale.Namespace = d.GetNamespace()
		scale.Name = d.GetName()
		_, err := d.Client.UpdateScale(d.Context, d.GetName(), scale, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *Deployment) Update() error {
	return kuberetry.RetryOnConflict(kuberetry.DefaultRetry, func() error {
		result, err := d.Client.Get(d.Context, d.GetName(), metav1.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				d.Storage.Log().Writef("get %s failed:%v", d.GetKind(), err)
			}
			return err
		}
		for k, container := range result.Spec.Template.Spec.Containers {
			image := strings.Split(container.Image, ":")
			if len(image) != 2 {
				continue
			}
			name, tag := image[0], image[1]
			d.Storage.Log().Writef("%s/%s container %s %s -> %s", result.Namespace, result.Name, container.Name, tag, d.GetImageTag())
			result.Spec.Template.Spec.Containers[k].Image = fmt.Sprintf("%s:%s", name, d.GetImageTag())
		}
		if d.IgnoreInitContainer == nil || !*d.IgnoreInitContainer {
			for k, container := range result.Spec.Template.Spec.InitContainers {
				image := strings.Split(container.Image, ":")
				if len(image) != 2 {
					continue
				}
				name, tag := image[0], image[1]
				d.Storage.Log().Writef("%s/%s init container %s %s -> %s", result.Namespace, result.Name, container.Name, tag, d.GetImageTag())
				result.Spec.Template.Spec.InitContainers[k].Image = fmt.Sprintf("%s:%s", name, d.GetImageTag())
			}
		}

		_, err = d.Client.Update(d.Context, result, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *Deployment) Println() error {
	return kuberetry.RetryOnConflict(kuberetry.DefaultRetry, func() error {
		result, err := d.Client.Get(d.Context, d.GetName(), metav1.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				d.Storage.Log().Writef("get %s failed:%v", d.GetKind(), err)
			}
			return err
		}
		for _, container := range result.Spec.Template.Spec.InitContainers {
			d.Storage.Log().Writef("%s/%s init container: %s", result.Namespace, result.Name, container.Image)
		}
		for _, container := range result.Spec.Template.Spec.Containers {
			d.Storage.Log().Writef("%s/%s container: %s", result.Namespace, result.Name, container.Image)
		}
		return nil
	})
}
