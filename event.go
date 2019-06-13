package commander

type event struct {
	handler Handler
	name    string
}

// Event creates a new event handler with the specified name.
// Contrary to commands, there can be multiple event handlers for a single event.
func (g *CommandGroup) Event(name string, handler Handler) {
	c := &event{
		handler: handler,
		name:    name,
	}

	g.events[c.name] = append(g.events[c.name], c)
}
