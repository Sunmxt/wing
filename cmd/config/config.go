package config

import (
	"errors"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	machineryConfig "github.com/RichardKnop/machinery/v1/config"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
)

type AuthConfiguration struct {
	EnableLDAP        bool              `default:"false" yaml:"enableLDAP" env:"ENABLE_LDAP"`
	DisableLegacyUser bool              `default:"false" yaml:"disableLegacyUser" env:"DISABLE_LEGACY_USER"`
	LDAP              LDAPConfiguration `yaml:"ldap"`
}

type LDAPConfiguration struct {
	Server                string            `yaml:"server" env:"LDAP_SERVER"`
	BindDN                string            `yaml:"bindDN" env:"LDAP_BIND_DN"`
	BindPassword          string            `yaml:"bindPassword" env:"LDAP_BIND_PASSWORD"`
	BaseDN                string            `yaml:"baseDN" env:"LDAP_BASE_DN"`
	SearchPattern         string            `yaml:"searchPattern" env:"LDAP_SEARCH_PATTERN"`
	NameAttribute         string            `yaml:"nameAttribute" env:"LDAP_NAME_ATTRIBUTE"`
	AcceptRegistration    bool              `yaml:"acceptRegistration" env:"LDAP_ACCEPT_REGISTERATION"`
	RegisterRDN           string            `yaml:"registerRDN" env:"LDAP_REGISTER_ROOT_DN"`
	RegisterObjectClasses []string          `yaml:"registerObjectClasses"`
	RegisterAttributes    map[string]string `yaml:"registerAttributes"`
}

type KubernetesConfiguration struct {
	Namespace  string `default:"KubeWing-" yaml:"namespace"`
	KubeConfig string `default:"" yaml:"kubeConfig"`
}

type DatabaseConfiguration struct {
	SQLDsn    string            `yaml:"dsn" default:""`
	SQLEngine string            `required:"true" yaml:"engine"`
	Database  string            `yaml:"database"`
	User      string            `yaml:"user"`
	Password  string            `yaml:"password"`
	Address   string            `yaml:"address"`
	Port      uint16            `yaml:"port"`
	Options   map[string]string `yaml:"options"`
}

func (c *DatabaseConfiguration) fillMySQLDsn() error {
	c.SQLDsn = ""
	if c.User != "" || c.Password != "" {
		c.SQLDsn = c.SQLDsn + c.User + ":" + c.Password + "@"
	}
	if c.Address == "" {
		return errors.New("Database address should not be empty.")
	}
	if c.Port == 0 {
		return errors.New("Database port should be greater than 0.")
	}
	c.SQLDsn = c.SQLDsn + "tcp(" + c.Address + ":" + strconv.FormatUint(uint64(c.Port), 10) + ")"
	if c.Database == "" {
		return errors.New("Database not specified.")
	}
	c.SQLDsn = c.SQLDsn + "/" + c.Database
	if c.Options == nil {
		c.Options = map[string]string{}
	}
	c.Options["parseTime"] = "true"
	optParts := make([]string, 0, len(c.Options))
	for k, v := range c.Options {
		optParts = append(optParts, k+"="+v)
	}
	c.SQLDsn = c.SQLDsn + "?" + strings.Join(optParts, "&")
	return nil
}

func (c *DatabaseConfiguration) Clean() (err error) {
	if c.SQLDsn == "" {
		switch c.SQLEngine {
		case "mysql":
			err = c.fillMySQLDsn()
		}
	}
	return err
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
	GitWorkingDir string `yaml:"gitWorkingDir" default:"/var/lib/wing/git"`

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
	failureIf := func(message string, assert bool) bool {
		if assert {
			err = errors.New(message)
			log.Error("[config] " + err.Error())
			return true
		}
		return false
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

	c.GitWorkingDir, err = filepath.Abs(c.GitWorkingDir)
	if err != nil {
		failureIf(err.Error(), true)
		return err
	}
	if failureIf("git working directory should not be blank.", c.GitWorkingDir == "") {
		return err
	}
	return nil
}

type SessionConfiguration struct {
	Job   JobConfiguration `yaml:"job"`
	Token string           `yaml:"token" default:"zPD78HgLVKoQsyCbdnBb4fSVDoZXc40JGMvHNuJ+wBM="`
}

type GelfConfiguration struct {
	Endpoint string                 `yaml:"endpoint" env:"GELF_ENDPOINT"`
	Tags     map[string]interface{} `yaml:"tags"`
}

type LoggingConfiguration struct {
	Driver string             `yaml:"driver" env:"LOG_DRIVER"`
	Gelf   *GelfConfiguration `yaml:"gelf"`
}

type WingConfiguration struct {
	Bind                    string                  `default:"0.0.0.0:8098" yaml:"bind"`
	ExternalURL             string                  `yaml:"externalURL"`
	DB                      DatabaseConfiguration   `yaml:"database"`
	Kube                    KubernetesConfiguration `yaml:"kubernetes"`
	InitialAdminCredentials string                  `yaml:"initialAdminCredentials" default:"admin"`
	Debug                   bool                    `yaml:"debug" default:false`
	DefaultLanguage         string                  `yaml:"defaultLanguage" default:"en"`
	Auth                    AuthConfiguration       `yaml:"auth"`
	Session                 SessionConfiguration    `yaml:"session"`
	NodeName                string                  `yaml:"nodeName" default:"" env:"NODE_NAME"`
	Log                     LoggingConfiguration    `yaml:"logging"`
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
	if err = c.DB.Clean(); err != nil {
		return err
	}
	if c.ExternalURL == "" {
		return errors.New("externalURl cannot be empty.")
	}
	_, err = url.Parse(c.ExternalURL)
	return
}

func maskString(target string, strs ...string) string {
	for _, str := range strs {
		if str == "" {
			continue
		}
		target = strings.Replace(target, str, "xxxxxxxx", -1)
	}
	return target
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
	log.Infof("[config]     database.user: %v", c.DB.User)
	log.Infof("[config]     database.password: %v", sensitiveMask(c.DB.Password))
	log.Infof("[config]     database.database: %v", c.DB.Database)
	log.Infof("[config]     database.address: %v", c.DB.Address)
	log.Infof("[config]     database.port: %v", c.DB.Port)
	log.Infof("[config]     database.dsn: %v", maskString(c.DB.SQLDsn, c.DB.Password))
	log.Infof("[config]     session.token: %v", sensitiveMask(c.Session.Token))
	log.Infof("[config]     session.job.concurrency: %v", c.Session.Job.Concurrency)
	log.Infof("[config]     session.job.brokerType: %v", c.Session.Job.BrokerType)
	log.Infof("[config]     session.job.address: %v", c.Session.Job.Address)
	log.Infof("[config]     session.job.port: %v", c.Session.Job.Port)
	log.Infof("[config]     session.job.password: %v", sensitiveMask(c.Session.Job.Password))
	log.Infof("[config]     session.job.username: %v", c.Session.Job.Username)
	log.Infof("[config]     session.job.redisDatabase: %v", c.Session.Job.RedisDatabase)
	log.Infof("[config]     session.job.machineryDSN: %v", maskString(c.Session.Job.MachineryDSN, c.Session.Job.Password))
	log.Infof("[config]     session.job.gitWorkingDir: %v", c.Session.Job.GitWorkingDir)
	log.Infof("[config]     logging.driver: %v", c.Log.Driver)
	if c.Log.Gelf != nil {
		log.Infof("[config]     logging.gelf.endpoint: %v", c.Log.Gelf.Endpoint)
	}
}
