package commander

import "strings"

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

func (g *CommandGroup) findEvents(name string) []*event {
	cmd, ok := g.events[name]

	var handlers []*event

	if ok {
		handlers = make([]*event, len(cmd))
		copy(handlers, cmd)
	}

	for i := 0; i < len(g.subgroups); i++ {
		handlers = append(handlers, g.subgroups[i].findEvents(name)...)
	}

	return handlers
}

func cleanStringSlice(slice []string) []string {
	j := 0
	for j < len(slice) {
		newt := strings.TrimSpace(slice[j])

		if newt == "" {
			// Remove this one
			slice = append(slice[:j], slice[j+1:]...)
		} else {
			slice[j] = newt
			j++
		}
	}

	return slice
}

func validateArgs(userArgs []string, parsers []argument) error {
	// Too few
	usize := len(userArgs) - 1
	psize := len(parsers)

	numOptional := 0
	infArgs := psize > 0 && (parsers[psize-1].argFlag == argumentStar ||
		parsers[psize-1].argFlag == argumentPlus)

	for _, p := range parsers {
		if p.argFlag == argumentOptional || p.argFlag == argumentStar {
			numOptional++
		}
	}

	if usize < psize-numOptional {
		return ErrNotEnoughArgs
	}

	if !infArgs && usize > psize {
		return ErrTooManyArgs
	}

	return nil
}

// Execute runs the parser on the provided text and executes the found command.
func (g *CommandGroup) Execute(text string) (bool, error) {
	return g.ExecuteWithContext(text, make(map[string]interface{}))
}

// ExecuteWithContext runs the parser on the provided text and executes the found command with the provided context.
func (g *CommandGroup) ExecuteWithContext(text string, context map[string]interface{}) (bool, error) {
	split := strings.Split(text, " ")

	split = cleanStringSlice(split)

	if g.Preprocessor != nil {
		if newargs, ok := g.Preprocessor.Process(split); !ok {
			return false, nil
		} else {
			split = newargs
		}
	}

	if len(split) == 0 || !strings.HasPrefix(split[0], "/") {
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
		params: context,
	}

	if err := validateArgs(split, cmd.parsers); err != nil {
		return true, err
	}

	argi := 1
	for i := range cmd.parsers {
		if len(split) <= argi &&
			(cmd.parsers[i].argFlag == argumentOptional || cmd.parsers[i].argFlag == argumentStar) {
			break
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

		argi++
	}

	handler := cmd.handler
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}

	// Execute command.
	return true, handler(*ctx)
}

// Trigger executes the event on all handlers.
func (g *CommandGroup) Trigger(name string) (bool, []error) {
	return g.TriggerWithContext(name, map[string]interface{}{})
}

// TriggerWithContext executes the event on all handlers with the provided context.
func (g *CommandGroup) TriggerWithContext(name string, context map[string]interface{}) (bool, []error) {
	ev := g.findEvents(name)

	if len(ev) == 0 {
		return false, nil
	}

	// Parse args
	ctx := &Context{
		Name:   name,
		Args:   []string{name},
		params: context,
	}

	var errs []error

	for i := range ev {
		err := ev[i].handler(*ctx)

		if err != nil {
			errs = append(errs, err)
		}
	}

	return true, errs
}
