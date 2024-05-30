// todo set lease. wait for lease to not exist on create
package db

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/meta"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	configMapName = "state"
	configMapKey  = "state"
)

type ConfigMap struct {
	client v1.ConfigMapInterface
}

func NewConfigMap() *ConfigMap {
	log.Debug().Msg("Creating new config map storage backend")
	cm := ConfigMap{
		client: newClient(),
	}
	return &cm
}

func (cm *ConfigMap) Load() *types.StateV1alpha1 {
	rawCM, err := cm.client.Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil || rawCM == nil {
		log.Err(err).Msg("Failed to get state")
		return nil
	}
	data, exists := rawCM.Data[configMapKey]
	if !exists || data == "" {
		log.Error().Str("name", configMapName).Str("key", configMapKey).Msg("Failed to find configMap")
		return nil
	}
	var stateWithVersion types.StateWithVersionVersion
	if err := json.Unmarshal([]byte(data), &stateWithVersion); err != nil {
		log.Err(err).Msg("Failed to unmarshal state into something with a defined version")
		return nil
	}
	switch version := stateWithVersion.Version; version {
	case meta.VersionV1alpha1:
		var state types.StateV1alpha1
		if err := json.Unmarshal([]byte(data), &state); err != nil {
			log.Err(err).Any("version", version).Msg("Failed to unmarshal state")
			return nil
		} else {
			return &state
		}
	default:
		log.Err(err).Any("version", version).Msg("unsupported version")
		return nil
	}
}

func (cm *ConfigMap) Store(state *types.StateV1alpha1) {
	panic("not implemented")
}

func newClient() v1.ConfigMapInterface {
	clientSet, error := inClusterClientSet()
	assertNotNil(error)
	return clientSet.CoreV1().ConfigMaps(currentPodNamespace())
}

func inClusterClientSet() (*kubernetes.Clientset, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	if clientSet, err := kubernetes.NewForConfig(k8sConfig); err != nil {
		return nil, err
	} else {
		return clientSet, nil
	}
}

func currentPodNamespace() string {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Err(err).Msg("Failed to get service account file. Returning an empty namespace")
		return ""
	}
	if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
		return ns
	} else {
		log.Error().Msg("Namespace could not be found from the service account file")
		return ""
	}
}

func assertNotNil(err error) {
	if err != nil {
		panic(err)
	}
}
