package db

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type MockConfigMapClient struct {
	ConfigMaps []v1.ConfigMap
}

func (c *MockConfigMapClient) Create(ctx context.Context, configMap *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error) {
	fmt.Println("herre")
	c.ConfigMaps = append(c.ConfigMaps, *configMap)
	return nil, nil
}
func (c MockConfigMapClient) Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	for i, cm := range c.ConfigMaps {
		if cm.Name == configMap.Name {
			c.ConfigMaps[i] = *configMap
			return nil, nil
		}
	}
	return nil, errors.New("Failed to find matching config map")
}
func (c *MockConfigMapClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return nil
}
func (c *MockConfigMapClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return nil
}
func (c *MockConfigMapClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	for _, cm := range c.ConfigMaps {
		if cm.Name == name {
			return &cm, nil
		}
	}
	return nil, errors.New("Failed to find matching config map")
}
func (c *MockConfigMapClient) List(ctx context.Context, opts metav1.ListOptions) (*v1.ConfigMapList, error) {
	return nil, nil
}
func (c *MockConfigMapClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (c *MockConfigMapClient) Patch(ctx context.Context, name string, pt k8sTypes.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ConfigMap, err error) {
	return nil, nil
}
func (c *MockConfigMapClient) Apply(ctx context.Context, configMap *corev1.ConfigMapApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ConfigMap, err error) {
	return nil, nil
}

func makeWithK8sConfigMap(k8sConfigMap ...v1.ConfigMap) *ConfigMap {
	cm := ConfigMap{
		client: &MockConfigMapClient{
			ConfigMaps: k8sConfigMap,
		},
	}
	return &cm
}

func TestConfigMapLoadNoConfigMapReturnsAnEmptyState(t *testing.T) {
	assert.NotNil(t, makeWithK8sConfigMap().Load())
}

func TestConfigMapSaveAndThenExistsInK8s(t *testing.T) {
	t.Setenv("NAMESPACE", "default")
	cm := makeWithK8sConfigMap()
	state := types.MakeState()
	cm.Store(&state)
	assert.True(t, cm.existsInK8s())
}

func TestConfigMapWithUnknownRepReturnsNilOnLoad(t *testing.T) {
	namespace := "test"
	t.Setenv("NAMESPACE", namespace)
	k8sConfigMap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			configMapKey: `{"a": "b"}`,
		},
	}
	cm := makeWithK8sConfigMap(k8sConfigMap)
	assert.Nil(t, cm.Load())
}
