package cmd

import (
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/RichardKnop/machinery/v1"
	machineryLog "github.com/RichardKnop/machinery/v1/log"
	"github.com/denisbrodbeck/machineid"
	"github.com/gemnasium/logrus-graylog-hook/v3"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/cmd/runtime"
	"git.stuhome.com/Sunmxt/wing/controller"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
)

const HelpTextHeader string = `Wing is server of application platform.

command:
    init        init platform.
    serve       run platform server.
    worker      run platform distrubuted job worker.

options:
`

type Wing struct {
	ConfigFile string
	ShowHelp   bool
	Debug      bool
	Runtime    runtime.WingRuntime
	flags      *flag.FlagSet

	Args    []string
	Opts    []string
	Command []string
}

func (c *Wing) ParseCommand(args []string) (cmd []string, flags []string) {
	cmd, flags = make([]string, 0, 1), make([]string, 0)
	if len(args) > 0 {
		cmd = append(cmd, args[0])
	}
	for _, v := range args[1:] {
		if strings.HasPrefix(v, "-") {
			flags = append(flags, v)
		} else {
			cmd = append(cmd, v)
		}
	}

	c.Flags().Parse(flags)
	return cmd, flags
}

func (c *Wing) Flags() *flag.FlagSet {
	if c.flags == nil {
		c.flags = flag.NewFlagSet("wing_config", flag.ContinueOnError)
		c.flags.StringVar(&c.ConfigFile, "config", "", "configuration file.")
		c.flags.BoolVar(&c.ShowHelp, "help", false, "show help.")
		c.flags.BoolVar(&c.Debug, "debug", false, "debug mode.")
	}

	return c.flags
}

func (c *Wing) ServerInit() error {
	if c.Runtime.Config == nil {
		c.Runtime.Config = &config.WingConfiguration{}
	}

	if err := c.Runtime.Config.Load(c.ConfigFile); err != nil {
		c.LogConfig()
		log.Error("Cannot load configuration: " + err.Error())
		return err
	}

	err := c.initMachineID()
	if err != nil {
		log.Error("Cannot determine machine ID. Wing refuses to launch for safety consideration. exiting...")
		return err
	}

	if c.Runtime.JobServer, err = machinery.NewServer(&c.Runtime.Config.Session.Job.MachineryConfig); err != nil {
		log.Error("[Worker] cannot create machinery server: " + err.Error())
		return err
	}

	c.Runtime.GitlabWebhookEventHub = gitlab.NewEventHub()
	c.Runtime.GitlabWebhookEventHub.Logger = log.WithFields(log.Fields{
		"module": "gitlab_webhook_hub",
	})

	if err = controller.RegisterTasks(&c.Runtime); err != nil {
		return err
	}

	// Remote logger
	switch c.Runtime.Config.Log.Driver {
	case "":
	case "gelf":
		hook := graylog.NewGraylogHook(c.Runtime.Config.Log.Gelf.Endpoint, c.Runtime.Config.Log.Gelf.Tags)
		log.AddHook(hook)
	default:
		log.Warn("Unsupported log driver: " + c.Runtime.Config.Log.Driver)
	}

	return nil
}
func (c *Wing) Parse() string {
	if len(os.Args) < 2 {
		c.Help()
		return ""
	}

	c.Args, c.Opts = c.ParseCommand(os.Args)
	c.Command = c.Args[1:]

	if len(c.Command) < 1 {
		c.Help()
		return ""
	}

	if err := c.ServerInit(); err != nil {
		return ""
	}

	return c.Command[0]
}

func (c *Wing) LogConfig() {
	log.Infof("[config] Options:")
	c.Flags().VisitAll(func(f *flag.Flag) {
		log.Infof("[config]     -%v=%v", f.Name, f.Value.String())
	})
	c.Runtime.Config.LogConfig()
}

func (c *Wing) Help() {
	fmt.Println(HelpTextHeader)
	c.Flags().PrintDefaults()
	fmt.Println("\nRun \"" + os.Args[0] + " <command> --help\" to get help of specific command.")
}

func (c *Wing) initMachineID() error {
	var idmeta []byte

	writeMachineID := func() {
		id := fmt.Sprintf("%x", md5.Sum(idmeta))
		log.Info("Generated machine ID: " + id)
		c.Runtime.MachineID = id
	}

	// machine ID from configuration "NodeName".
	if c.Runtime.Config.NodeName != "" {
		idmeta = []byte(c.Runtime.Config.NodeName)
		writeMachineID()
		log.Info("Generate machine ID from node name.")
		return nil
	}

	// Get machine ID from environment.
	if id, err := machineid.ID(); err == nil {
		if len(id) <= 128 {
			log.Info("Generate machine ID from environment.")
			idmeta = []byte(id)
			writeMachineID()
			return nil
		}
		log.Warn("Machine ID too long.")
	} else {
		log.Warn("Cannot retrieve machine ID from running environment: " + err.Error())
	}

	log.Info("Generate random machind ID.")

	machineIDPath := "/var/run/wing_machine_id"

	file, err := os.OpenFile(machineIDPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Error("Open machine id file failure:" + err.Error())
		return err
	}
	defer file.Close()

	var info os.FileInfo
	if info, err = file.Stat(); err != nil {
		log.Error("cannot get fileinfo:" + err.Error())
		return err
	}
	if info.Size() == 0 {
		uuid := strings.Replace(uuid.NewV1().String(), "-", "", -1)
		idmeta = []byte(uuid)
		file.Write(idmeta)
		writeMachineID()
		return nil

	} else if info.Size() > 1024 {
		err = errors.New("Machine ID file may be broken. You may delete \"" + machineIDPath + "\" then restart Wing server.")
		log.Error(err.Error())
		return err
	}

	if idmeta, err = ioutil.ReadAll(file); err != nil {
		err = errors.New("Cannot read machine ID file:" + err.Error())
		log.Error(err.Error())
		return err
	}
	writeMachineID()

	return nil
}

func (c *Wing) Exec() {
	c.initLogger()

	switch cmd := c.Parse(); cmd {
	case "serve":
		c.Serve()
	case "init":
		c.Init()
	case "worker":
		c.Worker()
	case "":
	default:
		fmt.Println("Unrecognized command - \"" + cmd + "\".")
	}
}

func (c *Wing) initLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	machineryLog.Set(log.WithFields(log.Fields{
		"module": "machinery",
	}))
}
