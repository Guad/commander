package commander

import (
	"strings"
)

// CommandGroup is a collection of commands and middlewares that only apply to this path.
type CommandGroup struct {
	subgroups      []*CommandGroup
	commands       map[string]*command
	middleware     []CommandMiddleware
	Contextualizer func(*Context)
}

// CommandMiddleware is a simple middleware for commands.
type CommandMiddleware func(Handler) Handler

// New returns a new initialized Command Group
func New() *CommandGroup {
	return &CommandGroup{
		commands: make(map[string]*command),
	}
}

// Group creates a new subgroup on this command group path. This is useful when using middlewares, as they
// will only apply to this path.
func (g *CommandGroup) Group() *CommandGroup {
	ng := New()

	g.subgroups = append(g.subgroups, ng)

	return ng
}

// Use adds a new middleware to this command group path.
func (g *CommandGroup) Use(mw CommandMiddleware) {
	g.middleware = append(g.middleware, mw)
}

func (g *CommandGroup) findCommand(prefix string) (*command, []CommandMiddleware, bool) {
	cmd, ok := g.commands[prefix]

	var mw []CommandMiddleware

	i := 0
	for !ok && i < len(g.subgroups) {
		cmd, mw, ok = g.subgroups[i].findCommand(prefix)
		i++
	}

	if ok {
		mw = append(g.middleware, mw...)
	}

	return cmd, mw, ok
}

// Execute runs the parser on the provided text and executes the found command.
func (g *CommandGroup) Execute(text string) (bool, error) {
	split := strings.Split(text, " ")

	j := 0
	for j < len(split) {
		newt := strings.TrimSpace(split[j])

		if newt == "" {
			// Remove this one
			split = append(split[:j], split[j+1:]...)
		} else {
			split[j] = newt
			j++
		}
	}

	if len(split) == 0 || !strings.HasPrefix(text, "/") {
		return false, nil
	}

	// Find the command
	cmd, mw, ok := g.findCommand(split[0])

	if !ok {
		return false, nil
	}

	// Parse args
	ctx := &Context{
		Name:   cmd.prefix,
		Args:   split[1:],
		params: make(map[string]interface{}),
	}

	argi := 1
	for i := range cmd.parsers {
		if len(split) <= argi &&
			(cmd.parsers[i].argFlag == argumentOptional || cmd.parsers[i].argFlag == argumentStar) {
			break
		}

		if len(split) <= argi {
			return true, ErrNotEnoughArgs
		}

		if cmd.parsers[i].argFlag == argumentPlus || cmd.parsers[i].argFlag == argumentStar {
			val, err := cmd.parsers[i].parser.Parse(strings.Join(split[argi:], " "))
			if err != nil {
				return true, ErrArgSyntaxError
			}

			ctx.params[cmd.parsers[i].name] = val
		} else {
			val, err := cmd.parsers[i].parser.Parse(split[argi])
			if err != nil {
				return true, ErrArgSyntaxError
			}
			ctx.params[cmd.parsers[i].name] = val
		}

		// last parser
		if i == len(cmd.parsers)-1 &&
			(cmd.parsers[i].argFlag == argumentSimple || cmd.parsers[i].argFlag == argumentOptional) &&
			len(split) > argi+1 {
			return true, ErrTooManyArgs
		}

		argi++
	}

	if g.Contextualizer != nil {
		g.Contextualizer(ctx)
	}

	handler := cmd.handler
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}

	// Execute command.
	return true, handler(*ctx)
}
