package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/runner/kubernetes/polymorphichelpers"
	"github.com/xmapst/AutoExecFlow/internal/runner/kubernetes/polymorphichelpers/types"
	"github.com/xmapst/AutoExecFlow/internal/storage"
)

type Kubectl struct {
	kubeConf            *rest.Config
	client              *kubernetes.Clientset
	dynamicClient       *dynamic.DynamicClient
	storage             storage.IStep
	subCommand          string
	workspace           string
	Config              string            `json:"kube_config" yaml:"KubeConfig"`
	Namespace           string            `json:"namespace" yaml:"Namespace"`
	ImageTag            string            `json:"image_tag" yaml:"ImageTag"`
	IgnoreInitContainer *bool             `json:"ignore_init_container" yaml:"IgnoreInitContainer"`
	Resources           []*types.Resource `json:"resources" yaml:"Resources"`
}

func New(storage storage.IStep, command, workspace string) (*Kubectl, error) {
	return &Kubectl{
		storage:    storage,
		subCommand: command,
		workspace:  workspace,
	}, nil
}

func (k *Kubectl) init() error {
	content, err := k.storage.Content()
	if err != nil {
		return err
	}
	if err = json.Unmarshal([]byte(content), k); err != nil {
		if err = yaml.Unmarshal([]byte(content), k); err != nil {
			return err
		}
	}

	if k.Config == "" {
		k.Config, err = k.storage.Env().Get("KUBECONFIG")
		if err != nil {
			return err
		}
	}

	file, err := os.ReadFile(k.Config)
	if err != nil {
		file = []byte(k.Config)
	}
	k.kubeConf, err = clientcmd.RESTConfigFromKubeConfig(file)
	if err != nil {
		return err
	}
	k.kubeConf.Burst = 1000
	k.kubeConf.QPS = 500
	k.kubeConf.Timeout = 5 * time.Minute

	k.client, err = kubernetes.NewForConfig(k.kubeConf)
	if err != nil {
		return err
	}

	k.dynamicClient, err = dynamic.NewForConfig(k.kubeConf)
	if err != nil {
		return err
	}
	for kk, res := range k.Resources {
		if res.Name == "" {
			return errors.New("name is empty")
		}
		k.Resources[kk].Namespace, err = k.getResourceValue(res.Namespace, k.Namespace, "NAMESPACE")
		if err != nil {
			return err
		}
		k.Resources[kk].ImageTag, err = k.getResourceValue(res.ImageTag, k.ImageTag, "IMAGE_TAG")
		if err != nil {
			return err
		}
		k.Resources[kk].Kind, err = k.getResourceValue(res.Kind, types.Deployment, "KIND")
		if err != nil {
			return err
		}
		if res.IgnoreInitContainer == nil {
			k.Resources[kk].IgnoreInitContainer = k.IgnoreInitContainer
		}
	}
	return nil
}

func (k *Kubectl) Run(ctx context.Context) (code int64, err error) {
	timeout, err := k.storage.Timeout()
	if err != nil {
		return common.CodeSystemErr, err
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err = k.init(); err != nil {
		return common.CodeSystemErr, err
	}
	var wg sync.WaitGroup
	var errCh = make(chan error, len(k.Resources))
	var done = make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case _err, ok := <-errCh:
				if !ok {
					return
				}
				err = errors.Join(err, _err)
			case <-ctx.Done():
				if ctx.Err() != nil {
					k.storage.Log().Write(ctx.Err().Error())
					err = ctx.Err()
				}
				return
			}
		}
	}()
	for _, resource := range k.Resources {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _err := k.run(ctx, resource); _err != nil {
				k.storage.Log().Writef("%s/%s Error: %s", resource.GetNamespace(), resource.GetName(), _err.Error())
				errCh <- _err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	<-done
	if err != nil {
		return common.CodeSystemErr, err
	}
	return
}

func (k *Kubectl) getResourceValue(lValue, gValue, env string) (string, error) {
	if lValue != "" {
		return lValue, nil
	}
	if gValue != "" {
		return gValue, nil
	}
	value, err := k.storage.Env().Get(env)
	if err != nil {
		return "", err
	}
	if value == "" {
		value, err = k.storage.GlobalEnv().Get(env)
	}
	return value, err
}

func (k *Kubectl) run(ctx context.Context, resource *types.Resource) (err error) {
	var rs = polymorphichelpers.ResourceFor(ctx, k.storage, k.client, resource)
	switch k.subCommand {
	case "restart":
		if err = rs.Restart(); err != nil {
			return err
		}
	case "update":
		if err = rs.Update(); err != nil {
			return err
		}
	case "scale":
		if err = rs.Scale(resource.GetReplicas()); err != nil {
			return err
		}
	case "status":
	default:
		return fmt.Errorf("unknown command: %s", k.subCommand)
	}
	if err = polymorphichelpers.Status(ctx, k.storage, k.dynamicClient, resource); err != nil {
		return err
	}
	return rs.Println()
}

func (k *Kubectl) Clear() error {
	return nil
}
