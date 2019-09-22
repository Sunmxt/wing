package config

import (
	"errors"
	machineryConfig "github.com/RichardKnop/machinery/v1/config"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type AuthConfiguration struct {
	EnableLDAP        bool              `default:"false" yaml:"enableLDAP"`
	DisableLegacyUser bool              `default:"false" yaml:"disableLegacyUser"`
	LDAP              LDAPConfiguration `yaml:"ldap"`
}

type LDAPConfiguration struct {
	Server                string            `yaml:"server"`
	BindDN                string            `yaml:"bindDN"`
	BindPassword          string            `yaml:"bindPassword"`
	BaseDN                string            `yaml:"baseDN"`
	SearchPattern         string            `yaml:"searchPattern"`
	NameAttribute         string            `yaml:"nameAttribute"`
	AcceptRegistration    bool              `yaml:"acceptRegistration"`
	RegisterRDN           string            `yaml:"registerRDN"`
	RegisterObjectClasses []string          `yaml:"registerObjectClasses"`
	RegisterAttributes    map[string]string `yaml:"registerAttributes"`
}

type KubernetesConfiguration struct {
	Namespace  string `default:"KubeWing-" yaml:"namespace"`
	KubeConfig string `default:"" yaml:"kubeConfig"`
}

type DatabaseConfiguration struct {
	SQLDsn    string `required:"true" yaml:"dsn"`
	SQLEngine string `required:"true" yaml:"engine"`
}

type JobConfiguration struct {
	Concurrency   int    `yaml:"concurrency" default:"1"`
	BrokerType    string `yaml:"brokerType" default:"redis"`
	Address       string `yaml:"address"`
	Port          uint16 `yaml:"port"`
	Password      string `yaml:"password"`
	Username      string `yaml:"username"`
	RedisDatabase string `yaml:"redisDatabase"`
	MachineryDSN  string `yaml:"machineryDSN"`

	MachineryConfig machineryConfig.Config `yaml:"-"`
}

func (c *JobConfiguration) GenerateMachineryDSN() (dsn string) {
	switch c.BrokerType {
	case "redis":
		dsn = "redis://"
		if c.Password != "" {
			dsn += c.Password + "@"
		}
		dsn += c.Address + ":" + strconv.FormatUint(uint64(c.Port), 10)
		if c.RedisDatabase != "" {
			dsn += "/" + c.RedisDatabase
		}
		return dsn
	}
	return ""
}

func (c *JobConfiguration) Clean() (err error) {
	failureIf := func(message string, assert bool) {
		if assert {
			err = errors.New(message)
			log.Error("[config] " + err.Error())
		}
	}
	failureIf("job broker address required", c.Address == "")
	failureIf("job broker port required", c.Port < 1)
	if err != nil {
		return err
	}
	switch c.BrokerType {
	case "redis":
		break
	default:
		failureIf("Unsupported job broker type: "+c.BrokerType, true)
		return
	}
	if c.MachineryDSN == "" {
		c.MachineryDSN = c.GenerateMachineryDSN()
	}

	// generate machinery configure.
	c.MachineryConfig.Broker = c.MachineryDSN
	c.MachineryConfig.DefaultQueue = "wing_default_jobs"
	c.MachineryConfig.ResultBackend = c.MachineryDSN
	c.MachineryConfig.ResultsExpireIn = 0
	c.MachineryConfig.NoUnixSignals = false

	return nil
}

type SessionConfiguration struct {
	Job   JobConfiguration `yaml:"job"`
	Token string           `yaml:"token" default:"zPD78HgLVKoQsyCbdnBb4fSVDoZXc40JGMvHNuJ+wBM="`
}

type WingConfiguration struct {
	Bind                    string                  `default:"0.0.0.0:8098" yaml:"bind"`
	DB                      DatabaseConfiguration   `yaml:"database"`
	Kube                    KubernetesConfiguration `yaml:"kubernetes"`
	InitialAdminCredentials string                  `yaml:"initialAdminCredentials" default:"admin"`
	Debug                   bool                    `yaml:"debug" default:false`
	DefaultLanguage         string                  `yaml:"defaultLanguage" default:"en"`
	Auth                    AuthConfiguration       `yaml:"auth"`
	Session                 SessionConfiguration    `yaml:"session"`
	NodeName                string                  `yaml:"nodeName" default:""`
}

func (c *WingConfiguration) Load(configFile string) error {
	if err := configor.Load(c, configFile); err != nil {
		return err
	}
	if err := c.Clean(); err != nil {
		return err
	}
	return nil
}

func (c *WingConfiguration) Clean() (err error) {
	if err = c.Session.Job.Clean(); err != nil {
		return err
	}
	return
}

func sensitiveMask(value interface{}) string {
	switch t := value.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
		if t == 0 {
			return "<empty>"
		}
	case string:
		if t == "" {
			return "<empty>"
		}
	}
	return "<configured>"
}

func (c *WingConfiguration) LogConfig() {
	log.Info("[config] configurations:")
	log.Infof("[config]     debug: %v", c.Debug)
	log.Infof("[config]     bind: %v", c.Bind)
	log.Infof("[config]     kubernetes.namespace: %v", c.Kube.Namespace)
	log.Infof("[config]     database.engine: %v", c.DB.SQLEngine)
	log.Infof("[config]     database.dsn: %v", c.DB.SQLDsn)
	log.Infof("[config]     session.job.concurrency: %v", c.Session.Job.Concurrency)
	log.Infof("[config]     session.job.brokerType: %v", c.Session.Job.BrokerType)
	log.Infof("[config]     session.job.address: %v", c.Session.Job.Address)
	log.Infof("[config]     session.job.port: %v", c.Session.Job.Port)
	log.Infof("[config]     session.job.password: %v", sensitiveMask(c.Session.Job.Password))
	log.Infof("[config]     session.token: %v", sensitiveMask(c.Session.Token))
	log.Infof("[config]     session.job.username: %v", c.Session.Job.Username)
	log.Infof("[config]     session.job.redisDatabase: %v", c.Session.Job.RedisDatabase)
	log.Infof("[config]     session.job.machineryDSN: %v", c.Session.Job.MachineryDSN)
}
