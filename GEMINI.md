# Gemini Project Overview: ilc

This document provides a high-level overview of the `ilc` command-line interface (CLI) project. It is intended to be a guide for developers and AI assistants to quickly understand the project's architecture, key components, and development patterns.

## Project Purpose

`ilc` is a tool that creates interactive command-line utilities from a single YAML file. It allows users to define commands, subcommands, inputs, and environment variables in a declarative way, and then uses an interactive TUI to guide the user through the execution of the commands.

## Directory Structure

```
.
├── cmd/ilc/
│   ├── command.go
│   ├── command_test.go
│   ├── config.go
│   ├── config_test.go
│   ├── execute.go
│   ├── execute_test.go
│   ├── help.go
│   ├── help_test.go
│   ├── input.go
│   ├── input_test.go
│   ├── main.go
│   ├── tui.go
│   └── types.go
├── examples/
│   ├── ilc.yml
│   ├── myecs.yml
│   └── single.yml
├── go.mod
├── go.sum
└── README.md
```

## Key Files and Their Purpose

### `cmd/ilc/main.go`

The entry point of the application. It is responsible for:

-   Parsing command-line flags (`--version`, `--help`, `--debug`).
-   Loading the configuration file.
-   Selecting the command to execute.
-   Collecting inputs from the user.
-   Executing the command.

### `cmd/ilc/config.go`

Handles loading and parsing the YAML configuration file. It defines the `Config` struct, which maps to the structure of the YAML file.

### `cmd/ilc/command.go`

Responsible for selecting the command to be executed based on the command-line arguments. It also includes the logic for the interactive command selection TUI.

### `cmd/ilc/input.go`

Handles the collection of inputs from the user. It supports collecting inputs from command-line arguments, environment variables, and an interactive TUI.

### `cmd/ilc/execute.go`

Responsible for executing the selected command. It uses Go's `text/template` package to create and execute the command's `run` script. It also handles setting environment variables.

### `cmd/ilc/tui.go`

Contains the implementation of the interactive terminal user interface (TUI) using the `charmbracelet/bubbletea` library. It defines the models for the command and input selection screens.

### `cmd/ilc/help.go`

Contains the logic for generating and printing the help screens. It provides context-aware help based on the loaded configuration and the selected command.

### `cmd/ilc/types.go`

Defines the core data structures used throughout the application, including `Config`, `Command`, and `Input`.

## Key Libraries

-   **`gopkg.in/yaml.v3`**: Used for parsing the YAML configuration files.
-   **`github.com/charmbracelet/bubbletea`**: Used for creating the interactive terminal user interface.

## Application Flow

1.  **Initialization**: The `main` function in `main.go` is the entry point. It parses command-line flags and loads the configuration.
2.  **Configuration Loading**: `loadConfig` in `config.go` finds and parses the YAML configuration file.
3.  **Command Selection**: `selectCommand` in `command.go` determines which command to run based on the command-line arguments. If no command is specified, it launches an interactive TUI to let the user choose.
4.  **Input Collection**: `collectInputs` in `input.go` gathers the necessary inputs for the selected command. It can get inputs from command-line arguments, environment variables, or an interactive TUI.
5.  **Execution**: `executeCommand` in `execute.go` runs the command. It uses Go's templating engine to construct the command string and then executes it in the specified shell.

## Testing

The project has a suite of unit tests for the core components. The tests are located in `_test.go` files alongside the code they are testing. To run the tests, use the following command:

```shell
go test ./cmd/ilc
```
