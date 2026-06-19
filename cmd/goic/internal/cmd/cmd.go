package cmd

import (
	"fmt"
	"strings"
)

func trimPrefix(s, prefix string) (string, bool) {
	if after, ok := strings.CutPrefix(s, prefix); ok {
		return after, true
	}
	return s, false
}

type Command struct {
	EmptyCommand
	Arguments []string
	Options   []CommandOption
	Method    func(args []string, options map[string]string) error
}

func (c *Command) Call(args ...string) error {
	if isHelp(args) {
		fmt.Println(c.Usage())
		return nil
	}
	arguments, options, err := c.extractArgumentsAndOptions(args)
	if err != nil {
		return err
	}
	return c.Method(arguments, options)
}

func (c *Command) Usage() string {
	var b strings.Builder
	b.WriteString(c.name)
	for _, a := range c.Arguments {
		fmt.Fprintf(&b, " <%s>", a)
	}
	if len(c.Options) != 0 {
		b.WriteString(" [options]")
	}
	fmt.Fprintf(&b, "\n  %s", c.description)
	if len(c.Options) != 0 {
		var pairs [][2]string
		for _, o := range c.Options {
			name := "--" + o.Name
			if o.HasValue {
				name += "=<value>"
			}
			pairs = append(pairs, [2]string{name, o.Description})
		}
		fmt.Fprintf(&b, "\n\nOptions:\n%s", pad(pairs))
	}
	return b.String()
}

func (c *Command) checkArguments(args []string) error {
	l := len(c.Arguments)
	if len(args) != l {
		var s []string
		for _, a := range c.Arguments {
			s = append(s, fmt.Sprintf("<%s>", a))
		}

		switch l {
		case 0:
			return NewInvalidArgumentsError("expected no arguments")
		default:
			return NewInvalidArgumentsError(fmt.Sprintf("expected %d arguments: %s", len(c.Arguments), strings.Join(s, " ")))
		}
	}
	return nil
}

func (c *Command) checkOptions(options map[string]string) error {
	for k, v := range options {
		var found bool
		for _, o := range c.Options {
			if o.Name == k {
				found = true
				if !o.HasValue && v != "" {
					return NewInvalidArgumentsError(fmt.Sprintf("option %s does not take a value", k))
				}
				break
			}
		}
		if !found {
			return NewInvalidArgumentsError(fmt.Sprintf("unknown option %s", k))
		}
	}
	return nil
}

func (c *Command) extractArgumentsAndOptions(args []string) ([]string, map[string]string, error) {
	var (
		arguments []string
		options   = make(map[string]string)
	)
	for _, a := range args {
		if a, ok := trimPrefix(a, "--"); ok {
			parts := strings.SplitN(a, "=", 2)
			if len(parts) == 2 {
				options[parts[0]] = parts[1]
			} else {
				options[parts[0]] = ""
			}
		} else {
			arguments = append(arguments, a)
		}
	}
	if err := c.checkArguments(arguments); err != nil {
		return nil, nil, err
	}
	if err := c.checkOptions(options); err != nil {
		return nil, nil, err
	}
	return arguments, options, nil
}

type CommandFork struct {
	EmptyCommand
	Commands []InternalCommand
}

func (c *CommandFork) Call(args ...string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Println(c.Usage())
		return nil
	}
	var name, args1 = args[0], args[1:]
	var cmd InternalCommand
	for _, c := range c.Commands {
		if c.Name() == name {
			cmd = c
			break
		}
	}
	if cmd == nil {
		return NewErrCommandNotFound(name)
	}
	return cmd.Call(args1...)
}

func (c *CommandFork) Usage() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s <command>\n  %s\n\nCommands:\n", c.name, c.description)
	var pairs [][2]string
	for _, sub := range c.Commands {
		pairs = append(pairs, [2]string{sub.Name(), sub.Description()})
	}
	b.WriteString(pad(pairs))
	return b.String()
}

type CommandOption struct {
	Name        string
	Description string
	HasValue    bool
}

type EmptyCommand struct {
	name        string
	description string
}

func (c *EmptyCommand) Call(_ ...string) error {
	return nil
}

func (c *EmptyCommand) Description() string {
	return c.description
}

func (c *EmptyCommand) Name() string {
	return c.name
}

type ErrCommandNotFound struct {
	Name string
}

func NewErrCommandNotFound(name string) *ErrCommandNotFound {
	return &ErrCommandNotFound{
		Name: name,
	}
}

func (e *ErrCommandNotFound) Error() string {
	return fmt.Sprintf("command %q not found", e.Name)
}

type ErrInvalidArguments struct {
	Expected string
}

func NewInvalidArgumentsError(expected string) *ErrInvalidArguments {
	return &ErrInvalidArguments{
		Expected: expected,
	}
}

func (e *ErrInvalidArguments) Error() string {
	return fmt.Sprintf("invalid arguments: %s", e.Expected)
}

type InternalCommand interface {
	Name() string
	Description() string
	Usage() string

	Call(args ...string) error
}

func isHelp(args []string) bool {
	for _, a := range args {
		if a == "help" || a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

// pad right-pads each name to the widest, returning aligned "  name  desc" lines.
func pad(pairs [][2]string) string {
	var w int
	for _, p := range pairs {
		if len(p[0]) > w {
			w = len(p[0])
		}
	}
	var b strings.Builder
	for _, p := range pairs {
		fmt.Fprintf(&b, "  %-*s  %s\n", w, p[0], p[1])
	}
	return strings.TrimRight(b.String(), "\n")
}

func NewCommand(
	name, description string,
	arguments []string,
	options []CommandOption,
	method func(args []string, options map[string]string) error,
) InternalCommand {
	return &Command{
		EmptyCommand: EmptyCommand{
			name:        name,
			description: description,
		},
		Arguments: arguments,
		Options:   options,
		Method:    method,
	}
}

func NewCommandFork(
	name, description string,
	subCommands ...InternalCommand,
) InternalCommand {
	return &CommandFork{
		EmptyCommand: EmptyCommand{
			name:        name,
			description: description,
		},
		Commands: subCommands,
	}
}
