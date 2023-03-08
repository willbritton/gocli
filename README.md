# Description
gocli is an opinionated framework for building multi-command CLIs in go.

It builds on the excellent https://github.com/spf13/pflag library and adds first class support for multiple sub commands as well as top level flags.

# Installation

`go get github.com/willbritton/gocli`

# Usage

## Write sub commands

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
	cli := gocli.NewCli(name)

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

	// register some sub commands
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

Using `--version` or `-v` will report the version information if provided.

If implemented, the banner will appear after the version information, but can be suppressed with `--no-banner` or `--silent`.

```
./greeter -v
version 0.0.0
Please help support the Retcon project at https://www.undeveloper.com/retcon#support
```

### Banner support

If a function is provided for the `Cli.Banner` field, this function will be called following almost every operation on the CLI, including when sub commands run.

The `Cli.Banner` function does not return a string, it simply runs whatever you tell it to, so it can be used to do some advanced stuff.

The banner invocation can be suppressed with either `--no-banner` or `--silent`.

```
./greeter -v
version 0.0.0
Please help support the Retcon project at https://www.undeveloper.com/retcon#support
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

### Using flags in sub commands

You can use the https://github.com/spf13/pflag library in your sub commands, but you may need to be careful about the top level flags you want to pass down to sub commands.

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

However, the sub command throws an error if one of the global commands are set, so the exit code is 1 and the command doesn't even run properly:

```
./greeter greet --name Will --no-banner || echo '(command failed)'


Please help support the Retcon project at https://www.undeveloper.com/retcon#support

(command failed)
```

To fix this you need to either manually register the global options to your sub command, or for convenience you can use the `cli.IgnoreGlobalOptions` method. The `cli.IgnoreGlobalOptions` method registers all the global options except those you explictly want to exclude, and by default it will mark them as hidden. Since you have access to the entire pflag API, you are free to subsequently modify the `Hidden` field if you like.

The second argument of `cli.IgnoreGlobalOptions` is an exclusion list of type `[]string`. If a global command appears in the exclusion list it will not be added to the sub command. This allows, for example, the sub command to provide its own implementation of the flag.

It's also useful to support the default help behavior of pflag, so often you'll want to exclude `"help"` from the global options and allow it to pass through, as in the following example:

```
var greeter = gocli.Command{
	Description: "greets a user",
	Handler: func(cli *gocli.Cli, cmd string, arguments []string) error {
		flags := flag.NewFlagSet(fmt.Sprintf("%s %s", cli.Name, cmd), flag.ContinueOnError)

		// adds all the global options as hidden flags to the sub command flag set
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

Note that while redefining a global flag in a sub command is possible, if you change the type or default behavior, for example, you might end up with mutually exclusive parse failures if the global and local versions of the flags are sufficiently incompatible.

### Displaying help text for global commands

If you are customizing your help text for sub commands, the `PrintGlobalOptions()` method can be used to print the global options help text as seen at the top level:

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

			// append the global options to the sub command help text
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

Note that the `Parse` will only return an error if the default pflag help function is activated (in which case it will return `flag.ErrHelp`) or if the command is not recognized. Unrecognized global options do not result in an error, because the only two ways the top level Cli is used is to pass through to sub commands or to display global help text.

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
