// todo set lease. wait for lease to not exist on create
package db

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/meta"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	configMapName = "state"
	configMapKey  = "state"
)

type ConfigMap struct {
	client corev1.ConfigMapInterface
}

func NewConfigMap() *ConfigMap {
	log.Debug().Msg("Creating new config map storage backend")
	cm := ConfigMap{
		client: newClient(),
	}
	return &cm
}

func (cm *ConfigMap) Load() (*types.StateV1alpha1, error) {
	k8sConfigMap, err := cm.client.Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil || k8sConfigMap == nil {
		log.Info().Err(err).Msg("State did not exist")
		state := types.MakeState()
		return &state, nil
	}
	data, exists := k8sConfigMap.Data[configMapKey]
	if !exists || data == "" {
		log.Error().Str("name", configMapName).Str("key", configMapKey).Msg("Failed to find configMap")
		return nil, errors.New("Failed to load state")
	}
	var stateWithVersion types.StateWithVersionVersion
	if err := json.Unmarshal([]byte(data), &stateWithVersion); err != nil {
		log.Err(err).Msg("Failed to unmarshal state into something with a defined version")
		return nil, errors.New("Failed to load state")
	}
	switch version := stateWithVersion.Version; version {
	case meta.VersionV1alpha1:
		var state types.StateV1alpha1
		if err := json.Unmarshal([]byte(data), &state); err != nil {
			log.Err(err).Any("version", version).Msg("Failed to unmarshal state")
			return nil, errors.New("Failed to load state")
		} else {
			// todo: set lease
			return &state, nil
		}
	default:
		log.Err(err).Any("version", version).Msg("unsupported version")
		return nil, errors.New("Failed to load state")
	}
}

func (cm *ConfigMap) Store(state *types.StateV1alpha1) {
	if !cm.existsInK8s() {
		log.Debug().Msg("State did not yet exist in k8s")
		_, err := cm.client.Create(context.Background(), cm.toK8s(state), metav1.CreateOptions{})
		if err != nil {
			log.Err(err).Msg("Failed to store k8s config map")
		}
	} else {
		_, err := cm.client.Update(context.Background(), cm.toK8s(state), metav1.UpdateOptions{})
		if err != nil {
			log.Err(err).Msg("Failed to update k8s config map")
		}
	}
}

func (cm *ConfigMap) existsInK8s() bool {
	k8sConfigMap, err := cm.client.Get(context.Background(), configMapName, metav1.GetOptions{})
	return k8sConfigMap != nil && err == nil
}

func (cm *ConfigMap) toK8s(state *types.StateV1alpha1) *v1.ConfigMap {
	if cm == nil || state == nil {
		log.Error().Msg("Cannot convert state to k8s config map - not defined")
		return nil
	}
	k8sConfigMap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: currentPodNamespace(),
		},
		Data: map[string]string{
			configMapKey: state.Marshal(),
		},
	}
	return &k8sConfigMap
}

func newClient() corev1.ConfigMapInterface {
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
	if namespace := os.Getenv("NAMESPACE"); namespace != "" {
		return namespace
	}
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
