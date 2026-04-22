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

	if c.silent != nil && *c.silent {
		if c.quiet != nil {
			*c.quiet = true
		}
		if c.noBanner != nil {
			*c.noBanner = true
		}
	}

	Log.SetVerbose()
	Dbg.SetQuiet()
	if c.debug != nil && *c.debug {
		Dbg.SetVerbose()
	} else if c.quiet != nil && *c.quiet {
		Log.SetQuiet()
	}

	cmdArg := ""
	if len(c.Args()) == 0 {
		return nil, flag.ErrHelp
	}
	cmdArg = c.Args()[0]
	cmd, exists := c.commands[cmdArg]
	if c.help != nil && *c.help {
		return cmd, flag.ErrHelp
	}
	if !exists {
		err = fmt.Errorf("command '%s' not recognized", cmdArg)
	}
	return cmd, err
}

func (c *Cli) Run(arguments []string) error {
	cmd, err := c.Parse(arguments)

	if c.version != nil && *c.version {
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

	if c.Banner != nil && (err != nil || (c.noBanner != nil && !*c.noBanner)) {
		fmt.Fprintln(os.Stderr)
		c.Banner()
	}
	fmt.Fprintln(os.Stderr)

	return err
}

// Deprecated: use [New] instead. This function will be removed in a future release.
func NewCli(name string) *Cli {
	return New(name)
}

type FlagOpt struct {
	Enabled   bool
	Name      string
	Shorthand string
	Usage     string
	Default   bool
}

type CliOpt struct {
	VersionFlag  FlagOpt
	HelpFlag     FlagOpt
	DebugFlag    FlagOpt
	NoBannerFlag FlagOpt
	QuietFlag    FlagOpt
	SilentFlag   FlagOpt
}

var DefaultOpt CliOpt = CliOpt{
	VersionFlag:  FlagOpt{Name: "version", Enabled: true, Shorthand: "v", Default: false, Usage: "prints the version of this program"},
	HelpFlag:     FlagOpt{Name: "help", Enabled: true, Shorthand: "h", Default: false, Usage: "prints help about a command"},
	DebugFlag:    FlagOpt{Name: "debug", Enabled: true, Default: false, Usage: "prints extra debug information for selected commands"},
	NoBannerFlag: FlagOpt{Name: "no-bannerEnabled: true, ", Default: false, Usage: "suppresses the banner text after this program runs"},
	QuietFlag:    FlagOpt{Name: "quiet", Enabled: true, Default: false, Usage: "suppresses all output except errors and banner"},
	SilentFlag:   FlagOpt{Name: "silent", Enabled: true, Default: false, Usage: "suppresses all output except errors"},
}

func NewWithOpt(name string, opt CliOpt) *Cli {
	c := new(Cli)
	c.Name = name
	c.commands = make(map[string]Cmd)
	c.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)
	c.ParseErrorsAllowlist.UnknownFlags = true

	c.version = c.addBuiltinFlag(opt.VersionFlag)
	c.help = c.addBuiltinFlag(opt.HelpFlag)
	c.debug = c.addBuiltinFlag(opt.DebugFlag)
	c.noBanner = c.addBuiltinFlag(opt.NoBannerFlag)
	c.quiet = c.addBuiltinFlag(opt.QuietFlag)
	c.silent = c.addBuiltinFlag(opt.SilentFlag)

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

func New(name string) *Cli {
	return NewWithOpt(name, DefaultOpt)
}

func (c *Cli) addBuiltinFlag(opt FlagOpt) *bool {
	if !opt.Enabled {
		return nil
	}
	if opt.Shorthand != "" {
		return c.BoolP(opt.Name, opt.Shorthand, opt.Default, opt.Usage)
	}
	return c.Bool(opt.Name, opt.Default, opt.Usage)
}
