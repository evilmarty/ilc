---
name: Configuration Testing
description: Instructions for compiling, running CLI tests, and verifying config inputs against example yaml targets.
---

# Configuration Testing Skill

This skill guides AI coding assistants on how to compile the `ilc` tool, run automated test suites, and perform manual testing on configurations inside the `examples/` directory.

---

## 1. Sandbox Compilation & Tests
Always use sandboxed-friendly compilation flags (`-buildvcs=false`) to ensure Go tools execute without VCS permission errors inside isolated terminal containers:
```bash
# Run unit tests
go test -buildvcs=false ./...

# Compile local binary
make build
```

---

## 2. Interactive Testing & verification Flows
Since `ilc` uses interactive Bubble Tea TUI components, use the following manual test procedures to verify user interface features against local configuration templates.

### A. Testing Numeric Value Adjustments & Bounds Clamping
Run the rating example to test number input fields, arrow key adjustments, and bounds validation:
```bash
./ilc examples/ilc.yml rate
```
1. **Interactive Commands**: Select the `rate` command, enter a name (e.g. `Laptop`), and navigate to the `rating` input field.
2. **Keyboard Adjustments**: Press the `Up` arrow key to increment the rating, and the `Down` arrow key to decrement.
3. **Verification**:
   - Check that ratings format cleanly (rendering integers like `3` instead of `3.00000`).
   - Confirm that the number clamps exactly at `1` (minimum) and `5` (maximum) and does not go out of bounds.
   - Verify that when the field is invalid or out-of-bounds, the Sage Green checkmark (`✔ `) changes to a Coral Red cross (`✘ `), and the `[Enter] Confirm` guidelines dynamically disappear.

### B. Testing String Validation & Dynamic Prompts
Run the greeting example to test string patterns, empty default placeholders, and progressive input stacking:
```bash
./ilc examples/ilc.yml greet
```
1. **Placeholder Visibility**: Confirm the default placeholder `"World"` is visible in the text field.
2. **Dynamic Validation**: Type valid characters to verify the soft green tick (`✔ `) renders immediately.
3. **Keystroke dry-run validation**: Erase inputs or type invalid non-regex matching characters (e.g. numbers or symbols if prohibited by pattern `^[a-zA-Z]+$`) to verify that the tick changes instantly to a soft red cross (`✘ `).

---

## 3. Help Command Cascades
Verify that the contextual help overrides function correctly for both global configurations and nested subcommand targets:
```bash
# Global help context (prints flags, root inputs, and child commands)
./ilc examples/ilc.yml -help

# Subcommand-specific help context (prints subcommand description, inherited parameters, and child command paths)
./ilc examples/ilc.yml weather -help
```
Confirm that these help commands output cleanly to standard error/output and exit with shell exit status code `0`.
