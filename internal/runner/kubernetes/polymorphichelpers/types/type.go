package types

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource struct {
	Kind      string `yaml:"Kind" json:"kind"`
	Namespace string `yaml:"Namespace" json:"namespace"`
	Name      string `yaml:"Name" json:"name"`
	Replicas  *int32 `yaml:"Replicas" json:"replicas"`
	ImageTag  string `yaml:"ImageTag" json:"image_tag"`
}

func (r *Resource) Check() error {
	if r.Namespace == "" {
		r.Namespace = metav1.NamespaceDefault
	}
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Kind == "" {
		r.Kind = Deployment
	}
	switch r.Kind {
	case Deployment, DaemonSet, StatefulSet:
	default:
		return fmt.Errorf("kind must be one of %s, %s, %s", Deployment, DaemonSet, StatefulSet)
	}
	return nil
}

func (r *Resource) GetKind() string {
	return r.Kind
}

func (r *Resource) GetNamespace() string {
	return r.Namespace
}

func (r *Resource) GetName() string {
	return r.Name
}

func (r *Resource) GetReplicas() int32 {
	if r.Replicas == nil {
		return 0
	}
	return *r.Replicas
}

func (r *Resource) GetImageTag() string {
	return r.ImageTag
}
