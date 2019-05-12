package config

import (
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
)

type KubernetesConfiguration struct {
	NamespacePrefix string `default:"KubeWing-" yaml:"namespacePrefix"`
}

type DatabaseConfiguration struct {
	SQLDsn    string `required:"true" yaml:"dsn"`
	SQLEngine string `required:"true" yaml:"engine"`
}

type WingConfiguration struct {
	Bind                    string                  `default:"0.0.0.0:8098" yaml:"bind"`
	DB                      DatabaseConfiguration   `yaml:"database"`
	Kube                    KubernetesConfiguration `yaml:"kubernetes"`
	InitialAdminCredentials string                  `yaml:"initialAdminCredentials" default:"admin"`
	SessionToken            string                  `yaml:"sessionToken" default:"zPD78HgLVKoQsyCbdnBb4fSVDoZXc40JGMvHNuJ+wBM="`
}

func (c *WingConfiguration) Load(configFile string) error {
	if err := configor.Load(c, configFile); err != nil {
		return err
	}
	return nil
}

func (c *WingConfiguration) LogConfig() {
	log.Info("[config] configurations:")
	log.Infof("[config]     Bind: %v", c.Bind)
	log.Infof("[config]     NamespacePrefix: %v", c.Kube.NamespacePrefix)
	log.Infof("[config]     SQLEngine: %v", c.DB.SQLEngine)
	log.Infof("[config]     SQLDsn: %v", c.DB.SQLDsn)
	log.Infof("[config]     SessionToken: %v", c.SessionToken)
}
