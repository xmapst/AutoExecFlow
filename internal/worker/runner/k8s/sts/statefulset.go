package sts

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

type SStatefulSet struct {
	context.Context
	*types.SResource
	Client  appv1.StatefulSetInterface
	Storage storage.IStep
}

func (s *SStatefulSet) Restart() error {
	return kuberetry.RetryOnConflict(kuberetry.DefaultRetry, func() error {
		path := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().Format(time.RFC3339))
		_, err := s.Client.Patch(s.Context, s.GetName(), kubetypes.StrategicMergePatchType, []byte(path), metav1.PatchOptions{})
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *SStatefulSet) Scale(replicas int32) error {
	return kuberetry.RetryOnConflict(kuberetry.DefaultRetry, func() error {
		scale := &autoscalingv1.Scale{
			Spec: autoscalingv1.ScaleSpec{
				Replicas: replicas,
			},
		}
		scale.Namespace = s.GetNamespace()
		scale.Name = s.GetName()
		_, err := s.Client.UpdateScale(s.Context, s.GetName(), scale, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *SStatefulSet) Update() error {
	return kuberetry.RetryOnConflict(kuberetry.DefaultRetry, func() error {
		result, err := s.Client.Get(s.Context, s.GetName(), metav1.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				s.Storage.Log().Writef("get %s failed:%v", s.GetKind(), err)
			}
			return err
		}
		for k, container := range result.Spec.Template.Spec.Containers {
			image := strings.Split(container.Image, ":")
			if len(image) != 2 {
				continue
			}
			name, tag := image[0], image[1]
			s.Storage.Log().Writef("%s/%s container %s %s -> %s", result.Namespace, result.Name, container.Name, tag, s.GetImageTag())
			result.Spec.Template.Spec.Containers[k].Image = fmt.Sprintf("%s:%s", name, s.GetImageTag())
		}
		if s.IgnoreInitContainer == nil || !*s.IgnoreInitContainer {
			for k, container := range result.Spec.Template.Spec.InitContainers {
				image := strings.Split(container.Image, ":")
				if len(image) != 2 {
					continue
				}
				name, tag := image[0], image[1]
				s.Storage.Log().Writef("%s/%s init container %s %s -> %s", result.Namespace, result.Name, container.Name, tag, s.GetImageTag())
				result.Spec.Template.Spec.InitContainers[k].Image = fmt.Sprintf("%s:%s", name, s.GetImageTag())
			}
		}

		_, err = s.Client.Update(s.Context, result, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *SStatefulSet) Println() error {
	return kuberetry.RetryOnConflict(kuberetry.DefaultRetry, func() error {
		result, err := s.Client.Get(s.Context, s.GetName(), metav1.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				s.Storage.Log().Writef("get %s failed:%v", s.GetKind(), err)
			}
			return err
		}
		for _, container := range result.Spec.Template.Spec.InitContainers {
			s.Storage.Log().Writef("%s/%s init container: %s", result.Namespace, result.Name, container.Image)
		}
		for _, container := range result.Spec.Template.Spec.Containers {
			s.Storage.Log().Writef("%s/%s container: %s", result.Namespace, result.Name, container.Image)
		}
		return nil
	})
}
