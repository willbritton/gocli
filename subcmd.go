package gocli

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

type SubCmdBuilder interface {
	Add(name string, description string, onParse func(*pflag.FlagSet))
	AddWithFlags(name string, description string, cfg func(*pflag.FlagSet), onParse func(*pflag.FlagSet))
	Enter(name string) error
	Parse() (*pflag.FlagSet, error)
	Exit() error
	Cancel()
	Set(flag *pflag.Flag, value string) error
	Active() bool
	Usage(name string) error
	Description(name string) (string, error)
}

type subCmd struct {
	*pflag.FlagSet
	onParse     func(*pflag.FlagSet)
	hiddenFlags map[string]bool
	description string
}

type subCmdBuilder struct {
	host     *pflag.FlagSet
	cmds     map[string]subCmd
	active   *subCmd
	builder  []string
	lastArg  int
	subFlags map[*pflag.Flag]bool
}

func (s *subCmdBuilder) AddWithFlags(name string, description string, cfg func(*pflag.FlagSet), onParse func(*pflag.FlagSet)) {
	s.cmds[name] = subCmd{
		FlagSet:     pflag.NewFlagSet(name, pflag.ContinueOnError),
		onParse:     onParse,
		hiddenFlags: map[string]bool{},
		description: description,
	}
	if cfg != nil {
		cfg(s.cmds[name].FlagSet)
	}
	s.cmds[name].VisitAll(func(f *pflag.Flag) {
		s.host.AddFlag(f)
		s.cmds[name].hiddenFlags[f.Name] = f.Hidden
		f.Hidden = true
		s.subFlags[f] = true
	})
}

func (s *subCmdBuilder) Add(name string, description string, onParse func(*pflag.FlagSet)) {
	s.AddWithFlags(name, description, nil, onParse)
}

func (s *subCmdBuilder) Enter(name string) error {
	if s.active != nil {
		if err := s.Exit(); err != nil {
			return err
		}
	}
	var err error
	if s.active, err = s.get(name); err != nil {
		return err
	}
	s.updateArgs()
	s.builder = []string{}
	s.active.VisitAll(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue)
		f.Hidden = s.active.hiddenFlags[f.Name]
	})
	return nil
}

func (s *subCmdBuilder) Parse() (*pflag.FlagSet, error) {
	active := s.active
	if active.FlagSet == nil {
		return nil, fmt.Errorf("no command active")
	}
	if err := active.Parse(s.builder); err != nil {
		return nil, err
	}
	s.active.onParse(active.FlagSet)
	return active.FlagSet, nil
}

func (s *subCmdBuilder) Cancel() {
	if s.active == nil {
		return
	}
	s.active.VisitAll(func(f *pflag.Flag) {
		f.Hidden = true
	})
	s.active = nil
	s.builder = []string{}
}

func (s *subCmdBuilder) Exit() error {
	if s.active == nil {
		return nil
	}
	s.updateArgs()

	if _, err := s.Parse(); err != nil {
		return err
	}
	s.Cancel()
	return nil
}

func (s *subCmdBuilder) Usage(name string) error {
	if err := s.Enter(name); err != nil {
		return err
	}
	defer s.Exit()
	if s.active.FlagSet.Usage != nil {
		s.active.FlagSet.Usage()
		return nil
	}
	if s.active.description != "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n\n", s.active.Name(), s.active.description)
	}
	s.active.FlagSet.PrintDefaults()
	return nil
}

func (s *subCmdBuilder) Set(flag *pflag.Flag, value string) error {
	active := s.active
	if err := flag.Value.Set(value); err != nil {
		return err
	}
	if s.active == nil {
		// if there is no active sub-command, reject any flags which are sub-command specific
		if s.subFlags[flag] {
			return fmt.Errorf("flag used outside a subcommand: --%s", flag.Name)
		}
		return nil
	}
	// if the active command changed during the last flag set, it must have been because this command is
	// a command delimiter - do not add to the builder
	if active != s.active {
		return nil
	}
	s.updateArgs()
	s.builder = append(s.builder, fmt.Sprintf("--%s=%v", flag.Name, value))
	return nil
}

func (s *subCmdBuilder) Active() bool {
	return s.active != nil
}

func (s *subCmdBuilder) Description(name string) (string, error) {
	cmd, err := s.get(name)
	if err != nil {
		return "", err
	}
	return cmd.description, nil
}

func NewSubCmdBuilder(host *pflag.FlagSet) SubCmdBuilder {
	return &subCmdBuilder{
		host:     host,
		cmds:     map[string]subCmd{},
		builder:  []string{},
		subFlags: map[*pflag.Flag]bool{},
	}
}

func (s *subCmdBuilder) get(name string) (*subCmd, error) {
	ok := false
	cmd, ok := s.cmds[name]
	if !ok {
		return nil, fmt.Errorf("command not recognized: %s", name)
	}
	return &cmd, nil
}

func (s *subCmdBuilder) updateArgs() {
	for _, a := range s.host.Args()[s.lastArg:] {
		s.builder = append(s.builder, a)
	}
	s.lastArg = len(s.host.Args())
}
