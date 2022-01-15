# Dotbot

*This project is still in active development don't expect everything to be working fully, and there will likely be breaking changes.*

Dotbot is a powerful dotfiles manager and system bootstrapping tool, it is based on the original [Dotbot](https://github.com/anishathalye/dotbot) written in Python by [@anishathalye](https://github.com/anishathalye), but it has been rebuilt from the ground up in Go, its faster and already has a much wider set of features.

This project has a much stronger focus on supporting "profiles" and "groups" allowing you to easily tweak what gets installed and linked on different systems. As well as sudo support, automatically requesting elevated permissions if it detects it needs them.

Dotbot can even be used as a lightweight package manager installing the latest versions of your core utilities irrespective of what package managers or packages are available on your OS.

## Installation

```bash
sh -c "$(curl -fsSL tinyurl.com/dotbot)"
```

Adding `init <owner>` will also clone your dotfiles repo, then adding `--apply` will run dotbot after cloning.

If you only specify your username its assumed the repo is called `dotfiles`.

```bash
sh -c "$(curl -fsSL tinyurl.com/dotbot)" -- init --apply <owner>[/<repo>]
```

## Usage

Proper documentation will arrive soon, for a sneak peek checkout my personal dotfiles [repo](https://github.com/jcwillox/dotfiles).

Once your repo is set up just run `dotbot`, it will self-update if needed, pull your dotfiles repo, and run your specified dotbot configuration.

```bash
$ dotbot
```
