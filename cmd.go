package gocli

type Cmd interface {
	GetDescription() string
	Run(cli *Cli, cmd string, arguments []string) error
}

type Command struct {
	Description string
	Handler     func(cli *Cli, cmd string, arguments []string) error
}

func (c Command) GetDescription() string {
	return c.Description
}

func (c Command) Run(cli *Cli, cmd string, arguments []string) error {
	return c.Handler(cli, cmd, arguments)
}
