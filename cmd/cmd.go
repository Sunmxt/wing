package cmd

import (
	"flag"
	"fmt"
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/common"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

const HelpTextHeader string = `Wing is server of application platform.

command:
    init        init platform.
    serve       run platform web server.

    instance    service instance resource.

options:
`

var CommandDepth map[string]int = map[string]int{
	"init":     0,
	"serve":    0,
	"instance": 1,
}

type Wing struct {
	ConfigFile string
	ShowHelp   bool
	Debug      bool
	Runtime    common.WingRuntime
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

	if c.Runtime.Config == nil {
		c.Runtime.Config = &config.WingConfiguration{}
	}

	if err := c.Runtime.Config.Load(c.ConfigFile); err != nil {
		c.LogConfig()
		log.Error("Cannot load configuration: " + err.Error())
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

func (c *Wing) Exec() {
	c.initLogger()

	switch cmd := c.Parse(); cmd {
	case "serve":
		c.Serve()
	case "init":
		c.Init()
	case "":
	default:
		fmt.Println("Unrecognized command - \"" + cmd + "\".")
	}
}

func (c *Wing) initLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}
