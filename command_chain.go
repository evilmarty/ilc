package main

type CommandChain []*Command

func (commands CommandChain) Last() *Command {
	if l := len(commands); l > 0 {
		return commands[l-1]
	} else {
		return nil
	}
}

func (commands CommandChain) Pure() bool {
	if c := commands.Last(); c != nil {
		return c.Pure
	} else {
		return false
	}
}

func (commands CommandChain) Inputs() Inputs {
	inputs := make(Inputs, 0)
	for _, command := range commands {
		inputs = append(inputs, command.Inputs...)
	}
	return inputs
}
