package db

type ConfigMap struct {
}

func NewConfigMap() *ConfigMap {
	return &ConfigMap{}
}

// todo set lease. wait for lease to not exist on create
