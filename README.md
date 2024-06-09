## ILC - The simple way to create a command-line utility

[![CI](https://github.com/evilmarty/ilc/actions/workflows/ci.yml/badge.svg)](https://github.com/evilmarty/ilc/actions/workflows/ci.yml)

Create an easy to use interactive CLI to simplify your workflow with a single YAML file.

## Installation

### Homebrew

To install via [Homebrew](https://brew.sh) just run the following command:

```shell
brew install --cask evilmarty/ilc/ilc
```

### Golang

Ensure you have [Go](https://go.dev) installed then run the follow:

```shell
go install github.com/evilmarty/ilc
```

### Manual

Binaries are available to download. Get the [latest release](https://github.com/evilmarty/ilc/releases/latest) binary for your platform.

## Usage

The usage is as followed:

```shell
ilc [--version] [--debug] CONFIG [COMMAND ...] [INPUT ...]
```

`CONFIG` is the path to your config file.

`COMMAND` is one or a cascade of subcommands defined in the config file.

`INPUT` is one or many inputs inherited by the command.

The best way to use `ilc` is to include it in the shebang of your config, like so:

```yaml
#!/usr/bin/env ilc
```

### Commands

If the configuration has defined `commands` they can either be passed as arguments or an interactive prompt will allow you to choose a command. If the command specified in the arguments has itself subcommands the interactive prompt will appear to complete the selection process.

### Inputs

After a command is specified or selected an interactive prompt will ask for input before the command will be executed. Inputs can be passed as arguments or as environment variables that are prefixed with `ILC_INPUT_`. Inputs that have been passed as arguments will not be asked, only for the inputs that have yet a value.

#### Example of passing inputs as arguments

```shell
ilc example/ilc.yaml calendar -month feb
```

#### Example of passing inputs via environment variables

```shell
export ILC_INPUT_month=feb
ilc example/ilc.yaml calendar
```

## Config

### `description`

The overall description of what is the config's purpose. Is optional.

### `env`

Optionally set environment variables for the command. Cascades to descending
commands and subcommands. Expressions can be used in values.

### `shell`

The shell to run the command in. Must be in JSON array format. Defaults to `["/bin/sh"]`.

### `run`

Runs command-line programs using the specified shell. See [[Templating]] for more information.

### `pure`

Setting `pure` to `true` to not pass through environment variables and only use environment variables that have been specified or inherited. Subcommands do not inherit this option and must be set for each command.

### `inputs`

Optionally specify inputs to be used in `run` and `env` values. Inputs can be passed as arguments or will be asked when invoking a command. Nested commands inherit inputs and cascade down.

### `inputs.<input_name>`

The key `input_name` is a string and its value is a map of the input's
configuration. The name can be used as an argument in the form `-<input_name>`
or `--<input_name>` followed by a value. The input's value is a string.

### `inputs.<input_name>.description`

Optionally describe the input's purpose or outcome.

### `inputs.<input_name>.options`

Limit the value to a list of acceptable values. Options can be a list of values
or a map, with the keys presented as labels and the corresponding values the
resulting value.

#### Example

- A list of options:

```yaml
inputs:
  month:
    options:
      - January
      - February
      - March
      - April
      - May
      - June
      - July
      - August
      - September
      - October
      - November
      - December
```

- A map of options:

```yaml
inputs:
  city:
    options:
      Brisbane: bne
      Melbourne: mlb
      Sydney: syd
```

### `inputs.<input_name>.pattern`

A regex pattern to validate the input's value. Default is to allow any input.

#### Example

```yaml
inputs:
  year:
    pattern: "(19|20)[0-9]{2}"
```

### `inputs.<input_name>.default`

Set the default value for the input. It is overwritten when a value is given as
an argument or changed when prompted. If a default value is not defined then a
value is required.

### `run`

Runs command-line programs using the specified `shell`. See [[Templating]] for more information.

#### Example

- A single-line command:

```yaml
run: cal
```

- A multi-line command:

```yaml
run: |
  cal
  date
```

### `commands`

The commands defined are then available to be invoked from the command line
either by passing arguments or by interactive selection. Must define at least
one command.

### `commands.<command_name>`

Use `commands.<command_name>` to give your command a unique name. The key
`command_name` is a string and its value is a map of the command's configuration
data. A string value can be used as a shorthand for the `run` attribute.

#### Example

```yaml
commands:
  calendar: cal
```

### `commands.<command_name>.description`

Optionally describe the command's purpose or outcome.

### `commands.<command_name>.run`

See [`run`] for more information.

### `commands.<command_name>.commands`

Define sub-commands in the same structure as `commands`. All `inputs` or `env`
defined cascade to all sub-commands. Cannot be used in conjunction with `run`.

### `commands.<command_name>.env`

Optionally set environment variables for the command. Cascades to descending
commands and subcommands. See [[Templating]] for more information.

#### Example

```yaml
commands:
  greet:
    env:
      NAME: "{{ .Input.name }}"
      GREETING: Hello
```

### `commands.<command_name>.pure`

Setting `pure` to `true` to not pass through environment variables and only use environment variables that have been specified or inherited. Subcommands do not inherit this option and must be set for each command.

### `commands.<command_name>.inputs`

Optionally specify inputs to be used in `run` and `env` values. Inputs can be passed as arguments or will be asked when invoking a command. Nested commands inherit inputs and cascade down. See [`inputs`] for more information.

## Templating

Go's [templating](https://pkg.go.dev/text/template) is available for `run` and `env` values to construct complex entries. Templates are evaluated after inputs are collected but before script execution. Along with inputs, templates can access environment variables that are present and regardless whether `pure` is enabled or not.

### .Input.<input_name>

The expression to reference an input value. ie. '{{ .Input.my_input }}'

### .Env.<variable_name>

The expression to reference an environment variable. ie. '{{ .Env.HOME }}'

## Example config with single command

```yaml
description: Display a calendar for the month
inputs:
  month:
    options:
      - January
      - February
      - March
      - April
      - May
      - June
      - July
      - August
      - September
      - October
      - November
      - December
run: cal -m {{ .Input.month }}
```

## Example config with commands

```yaml
description: My awesome CLI
commands:
  weather:
    description: Show the current weather forecast
    run: curl wttr.in?0QF
  calendar:
    description: Display a calendar for the month
    inputs:
      month:
        options:
          - January
          - February
          - March
          - April
          - May
          - June
          - July
          - August
          - September
          - October
          - November
          - December
    run: cal -m {{ .Input.month }}
  greet:
    description: Give a greeting
    inputs:
      name:
        default: World
      greeting:
        options:
          - Hello
          - Hi
          - G'day
    run: echo $GREETING $NAME
    env:
      NAME: "{{ .Input.name }}"
      GREETING: "{{ .Input.greeting }}"
```

## TODO

- [ ] Better help output
- [ ] Support dynamic options
