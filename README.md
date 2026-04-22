# Description
gocli is an opinionated framework for building multi-command CLIs in go.

It builds on the excellent https://github.com/spf13/pflag library and adds first class support for multiple commands as well as global flags.

There is also a separate standalone interface to help implement "sub-command" scenarios, where sequences of the command line are executed as discrete sub-commands.

# Installation

`go get github.com/willbritton/gocli`

# Usage

## Write commands

### By implementing the `gocli.Cmd` interface

```
type greeterCmd struct{}

var greeter = greeterCmd{}

func (c greeterCmd) GetDescription() string { return "greets a user" }

func (c greeterCmd) Run(cli *gocli.Cli, cmd string, arguments []string) error {
	println("Hello, world!")
	return nil
}
```

### Or by using the `gocli.Command` implementation

```
var greeter = gocli.Command{
	Description: "greets a user",
	Handler: func(cli *gocli.Cli, cmd string, arguments []string) error {
		println("Hello, world!")
		return nil
	},
}
```

## Set up the CLI

```
func main() {
	name := path.Base(os.Args[0])
	// create a new Cli instance
	cli := gocli.New(name)

	// optionally - add a description
	cli.Description = "A friendly CLI"

	// optionally - add a function to provide version information
	cli.Version = func() string {
		return "version 0.0.0"
	}

	// optionally - add a function to display a banner
	cli.Banner = func() {
		fmt.Fprintln(os.Stderr, "\033[1;97mPlease help support the Retcon project at https://www.undeveloper.com/retcon#support\033[0m")
	}

	// register some commands
	cli.RegisterCommand("greet", greeter)

	// when you have finished setting up your Cli, call Run passing the command line arguments
	err := cli.Run(os.Args[1:])
	if err != nil {
		os.Exit(1)
		return
	}
	os.Exit(0)
}
```

## Top-level Features

### Automatic help

Either if no commands specified, or if the `--help` or `-h` flags are used.

If implemented, the banner will be printed after the help text – the `--no-banner` flag is ignored for the help text.

```
./greeter --help
A friendly CLI

Usage:

      greeter <command> [options]

Available commands:

      greet         greets a user

Global options:

      --debug       prints extra debug information for selected commands
  -h, --help        prints help about a command
      --no-banner   suppresses the banner text after this program runs
      --quiet       suppresses all output except errors and banner
      --silent      suppresses all output except errors
  -v, --version     prints the version of this program

Please help support the Retcon project at https://www.undeveloper.com/retcon#support
```

### Version reporting

Using `--version` or `-v` will report the version information if provided. Set the `Cli.Version` field to a function which returns the version string.

If implemented, the banner will appear after the version information, but can be suppressed with `--no-banner` or `--silent`.

```
./greeter --version
version 0.0.0
Please help support the Retcon project at https://www.undeveloper.com/retcon#support
```

### Banner support

If a function is provided for the `Cli.Banner` field, this function will be called following almost every operation on the CLI, including when commands run.

The `Cli.Banner` function does not return a string, it simply runs whatever you tell it to, so it can be used to do some advanced stuff.

The banner invocation can be suppressed with either `--no-banner` or `--silent`.

```
./greeter --version --no-banner
version 0.0.0
```

### Default loggers

The `gocli` module provides two default loggers – `gocli.Log` and `gocli.Dbg`. These are lightweight wrappers around the built in `log.Logger` interface.

You can call the logging methods on `gocli.Log` or `gocli.Dbg` from your code. Both loggers will honor the `--quiet`, `--debug` and `--silent` flags in the following way:

| Flags | gocli.Log messages | gocli.Dbg messages | Banner | Error / CLI help messsages |
| --- | --- | --- | --- | --- |
| (none) | Active | Silent | Active | Active |
| `--no-banner` | Active | Silent | Silent | Active |
| `--debug` | Active | Active | Active | Active |
| `--quiet` | Silent | Silent | Active | Active |
| `--silent` | Silent | Silent | Silent | Active |

If the `--debug` flag is set, the `--quiet` and `--silent` flags will be overridden for logged output. However, the banner will still be suppressed if both `--debug` and `--silent` are set.

The `--silent` flag is equivalent to setting both `--no-banner` and `--quiet` flags at the same time.

For writing errors in your code, do not use either `gocli.Log` or `gocli.Dbg` - just write directly to standard error.

### Using flags in individual commands

You can use the https://github.com/spf13/pflag library in your commands, but you may need to be careful about the global flags you want to pass down to commands.

For example, here is an initial approach:

```
var greeter = gocli.Command{
	Description: "greets a user",
	Handler: func(cli *gocli.Cli, cmd string, arguments []string) error {
		flags := flag.NewFlagSet(fmt.Sprintf("%s %s", cli.Name, cmd), flag.ContinueOnError)
		name := flags.String("name", "world", "the name of the person to greet")
		err := flags.Parse(arguments)
		if err == nil {
			fmt.Printf("Hello, %s!\n", *name)
		}
		return err
	},
}
```

This works fine when supplying the `--name` flag alone:

```
./greeter greet --name Will
Hello, Will!

Please help support the Retcon project at https://www.undeveloper.com/retcon#support
```

However, the command throws an error if one of the global commands are set, so the exit code is 1 and the command doesn't even run properly:

```
./greeter greet --name Will --no-banner || echo '(command failed)'


Please help support the Retcon project at https://www.undeveloper.com/retcon#support

(command failed)
```

To fix this you need to either manually register the global options to your command, or for convenience you can use the `cli.IgnoreGlobalOptions` method. The `cli.IgnoreGlobalOptions` method registers all the global options except those you explictly want to exclude, and by default it will mark them as hidden. Since you have access to the entire pflag API, you are free to subsequently modify the `Hidden` field if you like.

The second argument of `cli.IgnoreGlobalOptions` is an exclusion list of type `[]string`. If a global command appears in the exclusion list it will not be added to the command. This allows, for example, the command to provide its own implementation of the flag.

It's also useful to support the default help behavior of pflag, so often you'll want to exclude `"help"` from the global options and allow it to pass through, as in the following example:

```
var greeter = gocli.Command{
	Description: "greets a user",
	Handler: func(cli *gocli.Cli, cmd string, arguments []string) error {
		flags := flag.NewFlagSet(fmt.Sprintf("%s %s", cli.Name, cmd), flag.ContinueOnError)

		// adds all the global options as hidden flags to the command flag set
		cli.IgnoreGlobalOptions(flags, []string{"help"})

		name := flags.String("name", "world", "the name of the person to greet")
		err := flags.Parse(arguments)
		if err == nil {
			fmt.Printf("Hello, %s!\n", *name)
		}
		return err
	},
}
```

Now our command will honor the `--no-banner` global command:

```
./greeter greet --name Will --no-banner || echo 'command failed'

Hello, Will!
```

And also we will get the default pflag help text for free:

```
./greeter greet --name Will --help                         

Usage of greeter greet:
      --name string   the name of the person to greet (default "world")

Please help support the Retcon project at https://www.undeveloper.com/retcon#support
```

Note that while redefining a global flag in a command is possible, if you change the type or default behavior, for example, you might end up with mutually exclusive parse failures if the global and local versions of the flags are sufficiently incompatible.

### Displaying help text for global commands

If you are customizing your help text for commands, the `PrintGlobalOptions()` method can be used to print the global options help text as seen at the top level:

```
var greeter = gocli.Command{
	Description: "greets a user",
	Handler: func(cli *gocli.Cli, cmd string, arguments []string) error {
		flags := flag.NewFlagSet(fmt.Sprintf("%s %s", cli.Name, cmd), flag.ContinueOnError)
		cli.IgnoreGlobalOptions(flags, []string{"help"})
		name := flags.String("name", "world", "the name of the person to greet")
		err := flags.Parse(arguments)
		if err == nil {
			fmt.Printf("Hello, %s!\n", *name)
		} else if err == flag.ErrHelp {

			// append the global options to the command help text
			cli.PrintGlobalOptions()

		}
		return err
	},
}
```

Results in:

```
./greeter greet --name Will --help

Usage of greeter greet:
      --name string   the name of the person to greet (default "world")

Global options:

      --debug       prints extra debug information for selected commands
  -h, --help        prints help about a command
      --no-banner   suppresses the banner text after this program runs
      --quiet       suppresses all output except errors and banner
      --silent      suppresses all output except errors
  -v, --version     prints the version of this program

Please help support the Retcon project at https://www.undeveloper.com/retcon#support
```

### Parsing without running (advanced usage)

You can perform the parsing of the CLI separately from running by using the `Parse` method instead of the `Run` method. The `Parse` method returns the `error` that pflag would normally return, and also a resolved `gocli.Cmd` interface, or `nil` if no command could be resolved.

Note that the `Parse` will only return an error if the default pflag help function is activated (in which case it will return `flag.ErrHelp`) or if the command is not recognized. Unrecognized global options do not result in an error, because the only two ways the top level Cli is used is to pass through to commands or to display global help text.

e.g.

```
cmd, err := cli.Parse(os.Args[1:])
if cmd != nil {
  println("Found: ", cmd.GetDescription())
} else {
  println("Not found")
}
if err != nil {
  println("Error: ", err.Error())
} else {
  println("No error")
}
```

Outputs:

```
./greeter greet --name Will

Found:  greets a user
No error

./greeter foo --name Will

Not found
Error:  command 'foo' not recognized

./greeter foo --name Will --help

Not found
Error:  pflag: help requested
```

## Configuring global flags

If you want to change the name, shorthand or usage of the global options; or if you don't want them registered at all, you can use the auxillary `NewWithOpt` constructor to `Cli`:

```
name := path.Base(os.Args[0])

// use the default options as a starting point
opt := gocli.DefaultOpt
// disable the debug flag entirely
opt.DebugFlag.Enabled = false
// rename the help flag
opt.HelpFlag.Name = "manual"
// change the help flag's shorthand
opt.HelpFlag.Shorthand = "m"
// change the help flag's usage test
opt.HelpFlag.Usage = "Displays the help manual for a command"
// disable the shorthand for the version flag
opt.VersionFlag.Shorthand = ""

cli := gocli.NewWithOpt(name, opt)
cli.Version = func() string { return "1.2.3" }
err := cli.Run(os.Args[1:])
if err != nil {
	os.Exit(1)
	return
}
os.Exit(0)
```

## The Sub-command Builder

The standalone `SubCmdBuilder` interface can be used to build up sub-commands within a command line which can be parsed as distinct units.

These work directly with `pflag` flags - they don't need to be used in conjunction with `gocli`'s multi-command applications described above.

Create a new instance of `SubCmdBuilder` with the `NewSubCmdBuilder` function, specifying a flagset to host the sub-commands, then add sub-commands with either `Add` or `AddWithFlags`. These take the command name and a callback function which will be called when the command is parsed. New sub-commands are built when the `Enter` method is called, and this can be done conveniently by passing `SubCmdBuilder.Enter` directly in a call to `Func` or `FuncP`.

When a new sub-command enters, it automatically causes any open ones to exit, so whenever your trigger flag is encountered any previous commands in the builder will be parsed and their `onParse` callback called if the sub-command parse was successful.

It's also possible to terminate a sub-command without starting a new one, by calling `SubCmdBuilder.Exit` explicitly. If it's necessary to cancel a sub-command without parsing it, you can call `SubCmdBuilder.Cancel`.

In order to correctly keep track of flags and args on the command line, it's necessary to call the special `Set` method of `SubCmdBuilder` for every flag on the host, rather than the normal flagset `Set` method. This is very easily achieved by calling `ParseAll` on the host flagset, rather than `Parse`, with the handler function being as simple as a call to `SubCmdBuilder.Set` for trivial implementations.

An error will be returned from the host parse if a flag is used which is not recognized by the current sub-command. Flags registered on the host will be ignored by the sub-command and will be parsed by the host as usual. This allows the functionality of entering and exiting sub-commands to work and is convenient generally to support flags which operate across all sub-commands.

An error will also be returned from the host parse if a flag which is designed to be used by a sub-command is run when no sub-command is active. This means that it's not possible to have a flag which works both at the host level and also in one or more sub-commands. In any case, such a usage might be confusing.

Note that all the sub-command flags are registered as flags on the host in order to allow host parsing to work properly, however the flags are hidden so as not to pollute the usage output.
It's possible to print usage for sub-commands using `SubCmdBuilder.Usage(subCmdName)`. This will call the `Usage` function on the flagset if defined, or else will print some basic default usage information. You can use `GetDescription` to return the registered description of the sub-command.

Here's a fully worked example which puts most of these ideas together:

```
name := path.Base(os.Args[0])
cli := gocli.New(name)

var repeat *int
cli.RegisterCommand("Run", gocli.Command{
	Description: "Sub-command runner",
	Handler: func(cli *gocli.Cli, cmd string, arguments []string) error {
		// need a FlagSet to host the sub-commands
		runFlags := pflag.NewFlagSet("Run", pflag.ContinueOnError)

		// create a new SubCmdBuilder
		subs := gocli.NewSubCmdBuilder(runFlags)
		// add a sub-command which prints all its arguments converted to uppercase
		subs.Add("ToUpper", "Converts the argument to uppercase", func(fs *pflag.FlagSet) {
			print(strings.ToUpper(strings.Join(fs.Args(), " ")))
		})
		// add a sub-command which lowercases its arguments
		// and also has an optional flag to repeat its output
		subs.AddWithFlags("ToLower", "Converts the argument to lowercase",
			func(fs *pflag.FlagSet) {
				repeat = fs.IntP("repeat", "r", 0, "Repeats the output")
			},
			func(fs *pflag.FlagSet) {
				s := strings.ToLower(strings.Join(fs.Args(), " "))
				if repeat != nil && *repeat > 0 {
					s = strings.Repeat(s, *repeat)
				}
				print(s)
			},
		)

		// declare a flag (on the host FlagSet) to trigger sub-commands
		runFlags.FuncP("cmd", "c", "Starts a command", subs.Enter)
		// other flags can terminate the sub-command without starting a new one
		runFlags.BoolFuncP("exit", "x", "Exits the latest command", func(s string) error { return subs.Exit() })
		// make a help command which supports sub-command usage
		runFlags.FuncP("help", "h", "Display help for commands", subs.Usage)

		// simplest way to ensure sub-commands keep track of all flags
		// is to call ParseAll with SubCmdBuilder.Set
		if err := runFlags.ParseAll(arguments, subs.Set); err != nil {
			return err
		}
		// need to ensure we close any open sub-commands at the end of the line
		return subs.Exit()
	},
})

err := cli.Run(os.Args[1:])
if err != nil {
	println(err.Error())
	os.Exit(1)
	return
}
os.Exit(0)
```

Outputs:

```
gocli Run -c ToUpper foo bar -c ToLower BAZ -r8
FOO BARbazbazbazbazbazbazbazbaz

gocli Run -c ToUpper -r8 foo bar -c ToLower BAZ
unknown flag: --repeat
exit status 1

gocli Run -c ToUpper foo -x bar -c ToLower BAZ -x QUX
FOObaz

gocli Run -c ToUpper foo bar -c ToLower BAZ -x -r8
FOO BARbaz
flag used outside a subcommand: --repeat
exit status 1

gocli Run -h ToLower
ToLower: Converts the argument to lowercase

  -r, --repeat int   Repeats the output
```
