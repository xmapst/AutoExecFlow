package types

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SResource struct {
	Kind                string `yaml:"Kind" json:"kind"`
	Namespace           string `yaml:"Namespace" json:"namespace"`
	Name                string `yaml:"Name" json:"name"`
	Replicas            *int32 `yaml:"Replicas" json:"replicas"`
	ImageTag            string `yaml:"ImageTag" json:"image_tag"`
	IgnoreInitContainer *bool  `json:"ignore_init_container" yaml:"IgnoreInitContainer"`
}

func (r *SResource) Check() error {
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

func (r *SResource) GetKind() string {
	return r.Kind
}

func (r *SResource) GetNamespace() string {
	return r.Namespace
}

func (r *SResource) GetName() string {
	return r.Name
}

func (r *SResource) GetReplicas() int32 {
	if r.Replicas == nil {
		return 0
	}
	return *r.Replicas
}

func (r *SResource) GetImageTag() string {
	return r.ImageTag
}
