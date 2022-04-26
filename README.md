ILC - The simple way to create a command-line utility
-----------------------------------------------------

[![CI](https://github.com/evilmarty/ilc/actions/workflows/ci.yml/badge.svg)](https://github.com/evilmarty/ilc/actions/workflows/ci.yml)

Create an easy to use interactive CLI to simplify your workflow with a single YAML file.

## Installation

Ensure you have [Go](https://go.dev) installed then run the follow:

```shell
go install github.com/evilmarty/ilc
```

## Usage

Run `ilc` for it to load `ilc.yml` in the current directory, or to specify a
config file pass `-f` with the path. A config file is required.

## Config

### `description`

The overall description of what is the config's purpose. Is optional.

### `commands`

The commands defined are then available to be invoked from the command line
either by passing arguments or by interactive selection. Must define at least
one command.

### `commands.<command_name>`

Use `commands.<command_name>` to give your command a unique name. The key
`command_name` is a string and its value is a map of the command's configuration
data.

### `commands.<command_name>.description`

Optionally describe the command's purpose or outcome.

### `commands.<command_name>.run`

Runs command-line programs using the operating system's shell. Inputs defined
are available to use via expression. Go's
[templating](https://pkg.go.dev/text/template) syntax is fully supported here.

#### Example

* A single-line command:

```yaml
commands:
  calendar:
    run: cal
```

* A multi-line command:

```yaml
commands:
  today:
    run: |
      cal
      date
```

### `commands.<command_name>.env`

Optionally set environment variables for the command. Expressions can be used
in values.

#### Example

```yaml
commands:
  greet:
    env:
      NAME: "{{ .name }}"
      GREETING: Hello
```

### `commands.<command_name>.inputs`

Optionally specify inputs to be used in `run` and `env` values. Inputs can be
passed as arguments or will be asked when invoking a command.

### `commands.<command_name>.inputs.<input_name>`

The key `input_name` is a string and its value is a map of the input's
configuration. The name can be used as an argument in the form `-<input_name>`
or `--<input_name>` followed by a value. The input's value is a string.

### `commands.<command_name>.inputs.<input_name>.options`

Limit the value to a list of acceptable values. Options can be a list of values
or a map, with the keys presented as labels and the corresponding values the
resulting value.

#### Example

* A list of options:

```yaml
commands:
  calendar:
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
* A map of options:

```yaml
commands:
  weater:
    inputs:
      city:
        options:
          Brisbane: bne
          Melbourne: mlb
          Sydney: syd
```


### `commands.<command_name>.inputs.<input_name>.pattern`

A regex pattern to validate the input's value. Default is to allow any input.

#### Example

```yaml
commands:
  calendar:
    inputs:
      year:
        pattern: "(19|20)[0-9]{2}"
```

### `commands.<command_name>.inputs.<input_name>.default`

Set the default value for the input. It is overwritten when a value is given as
an argument or changed when prompted. If a default value is not defined then a
value is required.

## Example config

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
        type: choice
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
    run: cal -m {{ .month }}
  greet:
    description: Give a greeting
    inputs:
      name:
        type: text
        default: World
      greeting:
        type: choice
        options:
          - Hello
          - Hi
          - G'day
    run: echo $GREETING $NAME
    env:
      NAME: "{{ .name }}"
      GREETING: "{{ .greeting }}"
```

## TODO

* [ ] Add tests
* [ ] Better help output
* [ ] Support dynamic options
* [ ] Sub-commands
