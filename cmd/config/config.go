package config

import (
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
)

type AuthConfiguration struct {
	EnableLDAP        bool              `default:"true" yaml:"enableLDAP"`
	DisableLegacyUser bool              `default:"false" yaml:"disableLegacyUser"`
	LDAP              LDAPConfiguration `yaml:"ldap"`
}

type LDAPConfiguration struct {
	Server        string `yaml:"server"`
	BindDN        string `yaml:"bindDN"`
	BindPassword  string `yaml:"bindPassword"`
	BaseDN        string `yaml:"baseDN"`
	SearchPattern string `yaml:"searchPattern"`
	NameAttribute string `yaml:"nameAttribute"`
}

type KubernetesConfiguration struct {
	Namespace  string `default:"KubeWing-" yaml:"namespace"`
	KubeConfig string `default:"" yaml:"kubeConfig"`
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
	Debug                   bool                    `yaml:"debug" default:false`
	DefaultLanguage         string                  `yaml:"defaultLanguage" default:"en"`
	Auth                    AuthConfiguration       `yaml:"auth"`
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
	log.Infof("[config]     KubernetesNamespace: %v", c.Kube.Namespace)
	log.Infof("[config]     SQLEngine: %v", c.DB.SQLEngine)
	log.Infof("[config]     SQLDsn: %v", c.DB.SQLDsn)
	if c.SessionToken != "" {
		log.Infof("[config]     SessionToken: <configured>")
	} else {
		log.Infof("[config]     SessionToken: <empty>")
	}
}
