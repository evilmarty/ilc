package main

type Config struct {
	Description string
	Env         map[string]string
	Shell       []string
	Run         string
	Pure        bool
	Inputs      map[string]Input
	Commands    map[string]Command
	FilePath    string
}

type Command struct {
	Description string
	Run         string
	Commands    map[string]Command
	Env         map[string]string
	Pure        bool
	Inputs      map[string]Input
	Aliases     []string
}

type Input struct {
	Type        string
	Description string
	Options     []any
	Pattern     string
	Default     any
	Min         float64
	Max         float64
}
