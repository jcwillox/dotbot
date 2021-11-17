import argparse
import os.path

parser = argparse.ArgumentParser()
parser.add_argument("name")
parser.add_argument(
    "--windows", action="store_true", help="create a separate file for windows"
)
parser.add_argument("--force", action="store_true")

args = parser.parse_args()

name: str = args.name
path = f"plugins/{name}"
file_path = os.path.join(path, f"{name}.go")
windows_file_path = os.path.join(path, f"{name}_windows.go")

if not os.path.isdir(path):
    os.mkdir(path)

with open("scripts/plugin_template.go") as file:
    plugin_template = (
        file.read()
        .replace("//go:build ignore\n", "")
        .replace("// +build ignore\n\n", "")
        .replace("plugin", name)
        .replace("Plugin", name.title())
        .replace("PLUGIN", name.upper())
    )

if os.path.exists(file_path) and args.force:
    os.rename(file_path, os.path.join(path, f"{name}.bak.go"))

if not os.path.exists(file_path):
    template = plugin_template
    if args.windows:
        template = "//go:build !windows\n// +build !windows\n\n" + plugin_template
    with open(file_path, "w+") as file:
        file.write(template)

if os.path.exists(windows_file_path) and args.force:
    os.rename(windows_file_path, os.path.join(path, f"{name}_windows.bak.go"))

if args.windows and not os.path.exists(windows_file_path):
    template = plugin_template
    with open(windows_file_path, "w+") as file:
        file.write(template)
