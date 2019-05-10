package config

//import (
//	"github.com/jinzhu/configor"
//)

type WingConfiguration struct {
	Bind string `default:"0.0.0.0:8098"`
}

func (c *WingConfiguration) Load(configFile string) error {
	return nil
}

func (c *WingConfiguration) LogConfig() {
    
}
