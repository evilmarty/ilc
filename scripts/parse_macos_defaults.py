import os
import re
import sys


def parse_frontmatter(content):
    title = ""
    description = ""
    if content.startswith("---"):
        parts = content.split("---", 2)
        if len(parts) >= 3:
            yaml_lines = parts[1].strip().splitlines()
            in_desc = False
            desc_lines = []
            for line in yaml_lines:
                if in_desc:
                    if line.startswith(" ") or line.strip() == "":
                        desc_lines.append(line.strip())
                    else:
                        in_desc = False

                if not in_desc:
                    if line.startswith("title:"):
                        t = line.split(":", 1)[1].strip()
                        if (t.startswith('"') and t.endswith('"')) or (
                            t.startswith("'") and t.endswith("'")
                        ):
                            t = t[1:-1]
                        title = t
                    elif line.startswith("description:"):
                        d = line.split(":", 1)[1].strip()
                        if d in ("|", ">"):
                            in_desc = True
                        else:
                            if (d.startswith('"') and d.endswith('"')) or (
                                d.startswith("'") and d.endswith("'")
                            ):
                                d = d[1:-1]
                            description = d
            if desc_lines:
                description = "\n".join(desc_lines)

    if " | " in title:
        title = title.split(" | ", 1)[0].strip()
    return title.strip(), description.strip()


def parse_write_command(cmd_line):
    # Match defaults write <domain> <key> <type> <value_and_more>
    pattern = (
        r'defaults\s+write\s+(\S+)\s+(?:"([^"]+)"|\'([^\']+)\'|(\S+))\s+(-\S+)\s+(.*)'
    )
    m = re.search(pattern, cmd_line)
    if not m:
        return None

    domain = m.group(1)
    key = m.group(2) or m.group(3) or m.group(4)
    param_type = m.group(5)
    rest = m.group(6).strip()

    post_cmd = ""
    value_part = rest
    if "&&" in rest:
        parts = rest.split("&&", 1)
        value_part = parts[0].strip()
        post_cmd = "&& " + parts[1].strip()

    clean_value = value_part
    if (clean_value.startswith('"') and clean_value.endswith('"')) or (
        clean_value.startswith("'") and clean_value.endswith("'")
    ):
        clean_value = clean_value[1:-1]

    if clean_value.startswith('\\"') and clean_value.endswith('\\"'):
        clean_value = clean_value[2:-2]

    return {
        "domain": domain,
        "key": key,
        "type": param_type,
        "raw_value": value_part,
        "clean_value": clean_value,
        "post_cmd": post_cmd,
    }


def clean_heading(text):
    text = text.strip()
    text = text.replace("`", "")
    # Strip (default value) or (default) case-insensitively
    text = re.sub(r"\s*\(\s*default\s*value\s*\)\s*", "", text, flags=re.IGNORECASE)
    text = re.sub(r"\s*\(\s*default\s*\)\s*", "", text, flags=re.IGNORECASE)
    # Strip "Set to " from the beginning of the string case-insensitively
    text = re.sub(r"^(?:Set to\s+)", "", text, flags=re.IGNORECASE)
    return text.strip()


def derive_input_name(cmd_name):
    # If it contains hyphens, take the last part
    if "-" in cmd_name:
        return cmd_name.split("-")[-1]
    # Suffixes check
    suffixes = [
        "size",
        "delay",
        "period",
        "threshold",
        "mode",
        "type",
        "format",
        "location",
        "path",
    ]
    for s in suffixes:
        if cmd_name.lower().endswith(s):
            return s
    # Otherwise fallback to 'value'
    return "value"


def try_parse_number(s):
    if s is None:
        return None
    s_str = str(s).strip()
    try:
        if "." in s_str:
            return float(s_str)
        return int(s_str)
    except ValueError:
        return s_str


def parse_delete_command(content):
    m = re.search(r"defaults\s+delete\s+.*", content)
    if m:
        return m.group(0).strip()
    return None


def main():
    docs_dir = next(iter(sys.argv), "docs")
    categories = sorted(
        [
            d
            for d in os.listdir(docs_dir)
            if os.path.isdir(os.path.join(docs_dir, d))
            and not d.startswith(".")
            and d not in ("public",)
        ]
    )

    yaml_lines = []
    yaml_lines.append("#!/usr/bin/env ilc")
    yaml_lines.append("description: |")
    yaml_lines.append("  macOS-defaults")
    yaml_lines.append("  https://macos-defaults.com/")
    yaml_lines.append("commands:")

    for cat in categories:
        cat_path = os.path.join(docs_dir, cat)
        index_path = os.path.join(cat_path, "index.md")

        cat_desc = cat.replace("-", " ").capitalize()
        if os.path.exists(index_path):
            with open(index_path, "r", encoding="utf-8") as f:
                _, idx_desc = parse_frontmatter(f.read())
                if idx_desc:
                    cat_desc = idx_desc

        yaml_lines.append(f"  {cat}:")
        yaml_lines.append("    description: |")
        for line in cat_desc.splitlines():
            yaml_lines.append(f"      {line}")
        yaml_lines.append("    commands:")

        # Get all settings md files in this category
        files = sorted(
            [f for f in os.listdir(cat_path) if f.endswith(".md") and f != "index.md"]
        )

        for file in files:
            cmd_name = file[:-3]  # Remove .md
            # Clean command name: strip leading underscores as they are not valid in command names
            cmd_name = cmd_name.lstrip("_")
            file_path = os.path.join(cat_path, file)

            with open(file_path, "r", encoding="utf-8") as f:
                content = f.read()

            title, desc = parse_frontmatter(content)
            if not desc:
                desc = title or cmd_name.replace("-", " ").capitalize()

            # Extract parameter type and any indented values
            p_type = "n/a"
            accepted_values = []
            type_match = re.search(
                r"(?:-\s*)?\*\*Parameter type\*\*:\s*([^\n]+)", content
            )
            if type_match:
                p_type = type_match.group(1).lower()
                # Find the position of the match
                pos = type_match.end()
                # Read subsequent lines to see if they are indented options
                rest_content = content[pos:]
                for line in rest_content.splitlines():
                    if line.strip() == "":
                        continue
                    # If it starts with indentation (at least two spaces or a tab) and a bullet point
                    m_bullet = re.match(r"^(?:\s{2,}|\t)[-*]\s+(.*)", line)
                    if m_bullet:
                        val = m_bullet.group(1).strip()
                        # Remove any backticks/quotes
                        val = val.replace("`", "").replace('"', "").replace("'", "")
                        accepted_values.append(val)
                    else:
                        # If a line has no indentation, we stop scanning for accepted values
                        if not line.startswith(" ") and not line.startswith("\t"):
                            break

            # Split by headings
            subsections = content.split("\n## ")

            write_commands = []
            delete_cmd = None

            # Find the reset command
            for sub in subsections:
                if sub.strip().startswith("Reset to default value"):
                    delete_cmd = parse_delete_command(sub)
                    break

            # Parse subsections for write commands
            for sub in subsections:
                lines = sub.splitlines()
                if not lines:
                    continue
                heading = lines[0].strip()

                # Check if it is a Set to or Add subsection
                if heading.startswith("Set to") or heading.startswith("Add"):
                    # Find code block
                    code_block_match = re.search(
                        r"```(?:bash|shell|sh)\s*\n(.*?)\n\s*```", sub, re.DOTALL
                    )
                    if code_block_match:
                        block_content = code_block_match.group(1)
                        for line in block_content.splitlines():
                            if "defaults write" in line:
                                parsed = parse_write_command(line)
                                if parsed:
                                    write_commands.append((heading, parsed))
                                    break

            # Determine the delete command fallback if not explicitly found
            if not delete_cmd and write_commands:
                # Fallback based on the first write command
                _, pw = write_commands[0]
                delete_cmd = f'defaults delete {pw["domain"]} "{pw["key"]}" {pw["post_cmd"]}'.strip()

            # Find the default value from headings
            default_val = None
            for sub in subsections:
                lines = sub.splitlines()
                if not lines:
                    continue
                heading = lines[0].strip()
                if (
                    "(default value)" in heading.lower()
                    or "(default)" in heading.lower()
                ):
                    if p_type == "bool":
                        if "true" in heading.lower():
                            default_val = "true"
                        elif "false" in heading.lower():
                            default_val = "false"
                    else:
                        # Find defaults write in this sub
                        code_block_match = re.search(
                            r"```(?:bash|shell|sh)\s*\n(.*?)\n\s*```", sub, re.DOTALL
                        )
                        if code_block_match:
                            block_content = code_block_match.group(1)
                            for line in block_content.splitlines():
                                if "defaults write" in line:
                                    parsed = parse_write_command(line)
                                    if parsed:
                                        default_val = (
                                            parsed["raw_value"]
                                            if parsed["type"] == "-array-add"
                                            else parsed["clean_value"]
                                        )
                                        break

            # Format the output for this command
            yaml_lines.append(f"      {cmd_name}:")
            yaml_lines.append("        description: |")
            for line in desc.splitlines():
                yaml_lines.append(f"          {line}")

            if p_type == "n/a" or not write_commands:
                # No inputs needed, simple run command
                # If we have a delete command, that's what it runs (e.g. clear timezones)
                # Otherwise if there is a write command, just run that
                run_cmd = delete_cmd
                if not run_cmd and write_commands:
                    # just use the raw defaults write command of the first option
                    _, pw = write_commands[0]
                    # rebuild the command line
                    run_cmd = f'defaults write {pw["domain"]} "{pw["key"]}" {pw["type"]} {pw["raw_value"]} {pw["post_cmd"]}'.strip()

                if run_cmd:
                    yaml_lines.append("        run: |")
                    yaml_lines.append(f"          {run_cmd}")
                else:
                    # Fallback in case there is absolutely no command found
                    yaml_lines.append("        run: \"echo 'No command documented'\"")
            elif p_type in ("int", "float"):
                # Number based parameter
                input_name = derive_input_name(cmd_name)
                default_num = try_parse_number(default_val)

                yaml_lines.append("        inputs:")
                yaml_lines.append(f"          {input_name}:")
                yaml_lines.append("            type: number")
                if default_num is not None:
                    yaml_lines.append(f"            default: {default_num}")

                # Build run template
                _, first_parsed = write_commands[0]
                domain = first_parsed["domain"]
                key = first_parsed["key"]
                type_flag = first_parsed["type"]
                post_cmd = first_parsed["post_cmd"]
                post_cmd_str = f" {post_cmd.strip()}" if post_cmd.strip() else ""

                yaml_lines.append("        run: |")
                if default_num is not None:
                    yaml_lines.append(
                        f"          {{{{ if (eq .Input.{input_name} {default_num}) }}}}"
                    )
                    yaml_lines.append(f"            {delete_cmd}")
                    yaml_lines.append("          {{ else }}")
                    yaml_lines.append(
                        f'            defaults write {domain} "{key}" {type_flag} "{{{{ .Input.{input_name} }}}}"{post_cmd_str}'
                    )
                    yaml_lines.append("          {{ end }}")
                else:
                    yaml_lines.append(
                        f'            defaults write {domain} "{key}" {type_flag} "{{{{ .Input.{input_name} }}}}"{post_cmd_str}'
                    )
            elif p_type.startswith("string") and not accepted_values:
                # Arbitrary string based parameter (e.g. screenshot save location)
                input_name = derive_input_name(cmd_name)

                yaml_lines.append("        inputs:")
                yaml_lines.append(f"          {input_name}:")
                yaml_lines.append("            type: string")
                if default_val is not None:
                    default_escaped = default_val.replace('"', '\\"')
                    yaml_lines.append(f'            default: "{default_escaped}"')

                # Build run template
                _, first_parsed = write_commands[0]
                domain = first_parsed["domain"]
                key = first_parsed["key"]
                type_flag = first_parsed["type"]
                post_cmd = first_parsed["post_cmd"]
                post_cmd_str = f" {post_cmd.strip()}" if post_cmd.strip() else ""

                yaml_lines.append("        run: |")
                if default_val is not None:
                    default_escaped = default_val.replace('"', '\\"')
                    yaml_lines.append(
                        f'          {{{{ if (eq .Input.{input_name} "{default_escaped}") }}}}'
                    )
                    yaml_lines.append(f"            {delete_cmd}")
                    yaml_lines.append("          {{ else }}")
                    yaml_lines.append(
                        f'            defaults write {domain} "{key}" {type_flag} "{{{{ .Input.{input_name} }}}}"{post_cmd_str}'
                    )
                    yaml_lines.append("          {{ end }}")
                else:
                    yaml_lines.append(
                        f'            defaults write {domain} "{key}" {type_flag} "{{{{ .Input.{input_name} }}}}"{post_cmd_str}'
                    )
            else:
                # We have write commands and options (string, array, bool) with specific options
                yaml_lines.append("        inputs:")
                yaml_lines.append("          mode:")
                yaml_lines.append("            type: string")
                if default_val is not None:
                    default_escaped = default_val.replace('"', '\\"')
                    yaml_lines.append(f'            default: "{default_escaped}"')
                yaml_lines.append("            options:")

                # If it's a bool, we use the standard Enable/Disable options
                if p_type == "bool":
                    yaml_lines.append('              Enable: "true"')
                    yaml_lines.append('              Disable: "false"')
                    yaml_lines.append('              Reset: "reset"')
                else:
                    # Dynamically build options
                    for heading, parsed in write_commands:
                        label = clean_heading(heading)
                        val = (
                            parsed["raw_value"]
                            if parsed["type"] == "-array-add"
                            else parsed["clean_value"]
                        )
                        # escape quotes in label and value if any
                        label_escaped = label.replace('"', '\\"')
                        val_escaped = val.replace('"', '\\"')
                        yaml_lines.append(
                            f'              "{label_escaped}": "{val_escaped}"'
                        )
                    yaml_lines.append('              Reset: "reset"')

                # Build run template
                _, first_parsed = write_commands[0]
                domain = first_parsed["domain"]
                key = first_parsed["key"]
                type_flag = first_parsed["type"]
                post_cmd = first_parsed["post_cmd"]

                yaml_lines.append("        run: |")
                yaml_lines.append('          {{ if (eq .Input.mode "reset") }}')
                yaml_lines.append(f"            {delete_cmd}")
                yaml_lines.append("          {{ else }}")

                # strip trailing and leading space from post_cmd and ensure single spacing
                post_cmd_str = f" {post_cmd.strip()}" if post_cmd.strip() else ""
                if type_flag == "-array-add":
                    # array-add doesn't use surrounding quotes in the template
                    yaml_lines.append(
                        f'            defaults write {domain} "{key}" {type_flag} {{{{ .Input.mode }}}}{post_cmd_str}'
                    )
                else:
                    yaml_lines.append(
                        f'            defaults write {domain} "{key}" {type_flag} "{{{{ .Input.mode }}}}"{post_cmd_str}'
                    )

                yaml_lines.append("          {{ end }}")

    # Write to macos.yml in the parent/workspace directory
    output_path = "examples/macos.yml"
    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(yaml_lines) + "\n")

    print(f"Successfully generated {output_path}")


if __name__ == "__main__":
    main()
