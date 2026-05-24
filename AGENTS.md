# AI Agent Workflows & Capabilities (`AGENTS.md`)

This repository is optimized for autonomous AI coding agents (such as Gemini, Antigravity, and Cursor). This document details the custom workspace skills and development guidelines available to agents working on this project.

---

## 🛠️ Local Agent Skills

We have defined specialized workspace skills within the `.agents/skills/` directory to guide agents through complex design and verification workflows. 

Agents are **highly encouraged** to read the `SKILL.md` instructions inside these folders before executing tasks in these domains:

### 1. [TUI Development Skill](file:///.agents/skills/tui-development/SKILL.md)
* **Path**: `.agents/skills/tui-development/SKILL.md`
* **Purpose**: Best practices for writing, rendering, and styling generic command-line interface prompts using **Bubble Tea** and **Lipgloss**. It explains pointer receivers, dry-run validations, input parsing tolerances, and key-press overrides.

### 2. [Interactive Testing & Configuration Verification Skill](file:///.agents/skills/configuration-testing/SKILL.md)
* **Path**: `.agents/skills/configuration-testing/SKILL.md`
* **Purpose**: Procedures for compiling, validating, and manually testing `ilc` command cascade configurations (`examples/*.yml`), verifying interactive flag parameters, and debugging TUI layouts.

---

## 🚀 Standard Agent Execution Guidelines

Whenever you are assigned a task on this repository, prioritize the following rules:

1. **Verify via Sandbox Testing**: Always compile and test code using Go's build system flags `-buildvcs=false` to avoid VCS lookup errors inside isolated terminal containers.
2. **Consult Local Skills**: If editing user interface elements or configuration parsing engines, read the respective `.agents/skills` instruction files first.
3. **Commit Cleanly**: Stage and commit your modifications bypassing GPG signing (`git commit --no-gpg-sign`) when working inside local sandboxes.
