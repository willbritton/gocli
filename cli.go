package gocli

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

var Log = DefaultLogger()
var Dbg = NewLogger()

type Cli struct {
	*flag.FlagSet
	Name        string
	Description string
	Version     func() string
	Banner      func()

	commands map[string]Cmd
	help     *bool
	version  *bool
	noBanner *bool
	debug    *bool
	quiet    *bool
	silent   *bool
}

func NewCli(name string) *Cli {
	c := new(Cli)
	c.Name = name
	c.commands = make(map[string]Cmd)
	c.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)
	c.ParseErrorsWhitelist.UnknownFlags = true

	c.version = c.BoolP("version", "v", false, "prints the version of this program")
	c.help = c.BoolP("help", "h", false, "prints help about a command")
	c.debug = c.Bool("debug", false, "prints extra debug information for selected commands")
	c.noBanner = c.Bool("no-banner", false, "suppresses the banner text after this program runs")
	c.quiet = c.Bool("quiet", false, "suppresses all output except errors and banner")
	c.silent = c.Bool("silent", false, "suppresses all output except errors")

	c.Usage = func() {
		if c.Description != "" {
			fmt.Fprintf(os.Stderr, "%s\n\n", c.Description)
		}
		fmt.Fprintf(os.Stderr, "Usage:\n\n      %s <command> [options]\n\n", name)
		fmt.Fprint(os.Stderr, "Available commands:\n\n")
		for k, v := range c.commands {
			if v.GetDescription() != "" {
				fmt.Fprintf(os.Stderr, "      %-13s %s\n", k, v.GetDescription())
			} else {
				fmt.Fprintf(os.Stderr, "      %-13s\n", k)
			}
		}
		c.PrintGlobalOptions()
	}

	return c
}

func (c *Cli) PrintGlobalOptions() {
	fmt.Fprint(os.Stderr, "\nGlobal options:\n\n")
	c.PrintDefaults()
}

func (c *Cli) IgnoreGlobalOptions(flags *flag.FlagSet, except []string) {
	m := make(map[string]struct{})
	for _, k := range except {
		m[k] = struct{}{}
	}
	c.Visit(func(f *flag.Flag) {
		_, exists := m[f.Name]
		if exists {
			return
		}
		flags.Var(f.Value, f.Name, f.Usage)
		flags.Lookup(f.Name).NoOptDefVal = f.NoOptDefVal
		flags.MarkHidden(f.Name)
	})
}

func (c *Cli) RegisterCommand(name string, cmd Cmd) {
	_, exists := c.commands[name]
	if exists {
		panic(fmt.Sprintf("command '%s' already exists", name))
	}
	c.commands[name] = cmd
}

func (c *Cli) Parse(arguments []string) (Cmd, error) {
	if c.Banner == nil {
		c.MarkHidden("no-banner")
	}
	if c.Version == nil {
		c.MarkHidden("version")
	}

	err := c.FlagSet.Parse(arguments)

	if *c.silent {
		*c.quiet = true
		*c.noBanner = true
	}

	Log.SetVerbose()
	Dbg.SetQuiet()
	if *c.debug {
		Dbg.SetVerbose()
	} else if *c.quiet {
		Log.SetQuiet()
	}

	cmdArg := ""
	if len(c.Args()) == 0 {
		return nil, flag.ErrHelp
	}
	cmdArg = c.Args()[0]
	cmd, exists := c.commands[cmdArg]
	if *c.help {
		return cmd, flag.ErrHelp
	}
	if !exists {
		err = fmt.Errorf("command '%s' not recognized", cmdArg)
	}
	return cmd, err
}

func (c *Cli) Run(arguments []string) error {
	cmd, err := c.Parse(arguments)

	if *c.version {
		v := "unknown version"
		if c.Version != nil {
			v = c.Version()
		}
		fmt.Fprint(os.Stderr, v)
		err = nil
	} else if cmd != nil {
		err = cmd.Run(c, arguments[0], arguments[1:])
	} else {
		c.Usage()
	}

	if c.Banner != nil && (err != nil || !*c.noBanner) {
		fmt.Fprintln(os.Stderr)
		c.Banner()
	}
	fmt.Fprintln(os.Stderr)

	return err
}
