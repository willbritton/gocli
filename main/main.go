package main

import (
	"os"
	"path"
	"strings"

	"github.com/spf13/pflag"
	"github.com/willbritton/gocli"
)

func main() {
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
}
