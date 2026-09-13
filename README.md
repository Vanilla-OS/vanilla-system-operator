<div align="center">
  <img src="vso-logo.svg" height="120">
  <h1 align="center">Vanilla System Operator</h1>
	
[![Translation Status][weblate-image]][weblate-url]

[weblate-url]: https://hosted.weblate.org/engage/vanilla-os/
[weblate-image]: https://hosted.weblate.org/widget/vanilla-os/vanilla-system-operator/svg-badge.svg
 
  <p align="center">VSO is a utility which allows you to perform maintenance tasks on your Vanilla OS installation.</p>
</div>

<br/>

## Help

```text
Usage: vso [flags] [command]

Commands:
  config          Manage the system configuration
  man             Generate the manual page
  native          Manage the VSO package subsystem
  tasks           Create and manage tasks
  upgrade         Check for or apply a system image update
```

> [!NOTE]
> Use `vso <COMMAND> --help` for command-specific options. The VSO 3 command groups
> replace the VSO 2 top-level package commands and `sys-upgrade` command.

## Documentation

The official **documentation and manpage** for `vso` are available at <https://docs.vanillaos.org/docs/en/vso>.

## VSO as system Shell

To use VSO as your system shell, you can copy the `usr/bin/vso-os-shell` script
to your system's `/usr/bin` directory and set it as your default shell. Your
image needs to implement the `usr/bin/os-shell` script, which will expand the
`$SHELL` environment variable, this is much needed for login shells and other
flags, this also ensures that the user's default shell is respected.

Our `vso-image` already implements this script.

## Use of Generative AI

Maintainers may use generative AI tools as assistants while working on vanilla-system-operator. Non-trivial assisted commits disclose the tool, model, and scope of the work.

AI tools may assist with code comments, documentation, repetitive code, and issue triage. Maintainers make project decisions and review every assisted change before it is merged.

Use these trailers for non-trivial assisted commits:

```plain
Assisted-by: <tool>:<model-version>
AI-Scope: <what the tool generated and the prompt or a short prompt summary>
```

Single-line completions, renames, and formatting changes do not need trailers.

Coding agents must also follow [AGENTS.md](AGENTS.md) before changing files,
creating commits, or opening pull requests.
