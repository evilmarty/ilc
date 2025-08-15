# Copilot Instructions

## Overview

This repository contains the `ilc` project, a tool for creating interactive command-line utilities using a simple YAML configuration. The following instructions are designed to help contributors and maintainers leverage GitHub Copilot effectively while working on this project.

### Application Purpose and Features

`ilc` simplifies workflows by allowing users to define commands and inputs in a YAML configuration file. It provides an interactive CLI utility that executes these commands seamlessly. Key features include:
- YAML-based configuration for defining commands and inputs.
- Interactive prompts for selecting commands and providing inputs.
- Environment variable support for passing inputs.

### High-Level Architecture

The application consists of the following key components:
- **Config**: Handles YAML parsing and configuration validation.
- **Runner**: Executes commands defined in the configuration.
- **Input**: Manages user inputs and environment variables.
- **Prompt**: Provides an interactive interface for selecting commands and inputs.

#### Data Flow
1. YAML configuration is parsed by `config.go`.
2. Inputs are collected interactively or via environment variables using `prompt.go`.
3. Commands are executed by `runner.go`, leveraging the collected inputs and environment variables.

## Best Practices

### Writing Code

1. **Follow the Project Structure**: Ensure that new code adheres to the existing structure and conventions. For example, tests should be placed in `*_test.go` files, and new features should be added to the appropriate `.go` files.

2. **Use Copilot for Boilerplate**: Leverage Copilot to generate repetitive or boilerplate code, such as struct definitions, interface implementations, or test cases. Always review and refine the generated code to ensure it meets project standards.

3. **Adhere to Go Standards**: Ensure that all code follows Go best practices, including proper naming conventions, idiomatic Go patterns, and effective use of Go modules.

### Writing Tests

1. **Test Coverage**: Aim for high test coverage. Use Copilot to generate test cases for new functions and methods. Place tests in the corresponding `*_test.go` files.

2. **Review Generated Tests**: While Copilot can generate test cases, always review them to ensure they cover edge cases and align with the intended functionality.

3. **Run Tests Locally**: Before committing, run all tests locally to ensure they pass. Use the `go test` command for this purpose.

### Documentation

1. **Update README**: If you add new features or make significant changes, update the `README.md` file to reflect these changes.

2. **Inline Comments**: Use Copilot to generate inline comments for complex code. Ensure comments are clear and concise.

3. **YAML Examples**: When adding new features related to YAML configuration, include examples in the `examples/` directory and update the `README.md` accordingly.

### GitHub Workflows

1. **CI/CD**: Ensure that all changes pass the CI checks defined in `.github/workflows/ci.yml`.

2. **Pull Requests**: Use Copilot to draft pull request descriptions. Clearly explain the changes and link to relevant issues.

3. **Issue Templates**: When creating new issues, follow the template in `.github/issue_template.md`.

### Using Copilot

1. **Code Suggestions**: Use Copilot to assist with code suggestions, but always validate the generated code.

2. **Prompt Engineering**: Provide clear and specific prompts to Copilot to get the best results. For example, "Generate a function to parse YAML configuration."

3. **Iterative Refinement**: Use Copilot iteratively. Start with a broad prompt, review the output, and refine the prompt as needed.

## Critical Developer Workflows

1. **Building the Application**:
   - Use the `make build` command to compile the application.

2. **Running Tests**:
   - Use the `make test` command or `go test ./...` to run all tests.

3. **Debugging**:
   - Use the `--debug` flag when running the application to enable detailed logging.

## YAML Configuration

### Overview

YAML configuration is central to `ilc`. It defines commands, inputs, and environment variables. The application processes these configurations to execute commands interactively.

### Examples

- **Basic Configuration**:
  ```yaml
  description: Display a calendar for the month
  inputs:
    month:
      options:
        - January
        - February
        - March
  run: cal -m {{ .Input.month }}
  ```

- **Advanced Configuration**:
  ```yaml
  description: My awesome CLI
  commands:
    greet:
      description: Give a greeting
      inputs:
        name:
          default: World
        greeting:
          options:
            - Hello
            - Hi
      run: echo $GREETING $NAME
      env:
        NAME: "{{ .Input.name }}"
        GREETING: "{{ .Input.greeting }}"
  ```

## Debugging and Troubleshooting

1. **Enable Debug Mode**: Use the `--debug` flag to get detailed logs.
2. **Check Logs**: Review logs for errors and warnings.
3. **Validate YAML**: Use a YAML linter to ensure your configuration is valid.

## Extensibility

1. **Adding New Input Types**: Extend the `Input` component to support additional input types (e.g., `date`).
2. **Enhancing Prompts**: Modify the `prompt.go` file to add new features to the interactive prompt.

## Codebase Navigation

- `config.go`: Handles YAML parsing and configuration validation.
- `runner.go`: Executes commands defined in the configuration.
- `input.go`: Manages user inputs and environment variables.
- `prompt.go`: Handles interactive prompts for user input.

## Integration Points

- **External Dependencies**:
  - `bubbletea`: Used for interactive prompts.
  - `yaml.v3`: Used for YAML parsing.
- **Environment Variables**:
  - Inputs can be passed via environment variables prefixed with `ILC_INPUT_`.

## Resources

- [Go Documentation](https://golang.org/doc/)
- [GitHub Copilot Documentation](https://docs.github.com/en/copilot)
- [YAML Specification](https://yaml.org/spec/)

By following these instructions, you can make the most of GitHub Copilot while contributing to the `ilc` project.
