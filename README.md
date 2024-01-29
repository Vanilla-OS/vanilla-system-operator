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

```md
The Vanilla System Operator is a package manager, a system updater and a task automator.

Usage:
  vso [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Manage the system configuration.
  export      Export an application or binary from the subsystem
  help        Help about any command
  install     Install an application inside the subsystem
  pico-init   Initialize the Pico subsystem, used for package management
  remove      Remove an application from the subsystem
  run         Run an application from the subsystem
  search      Search for an application to install inside the subsystem
  shell       Enter the subsystem environment
  sys-upgrade Upgrade the system
  tasks       Create and manage tasks
  unexport    Unexport an application or binary from the subsystem
  update      Update the subsystem's package repository
  upgrade     Upgrade the packages inside the subsystem

Flags:
  -h, --help      help for vso
  -v, --version   version for vso

Use "vso [command] --help" for more information about a command.
```

## Documentation

The official **documentation and manpage** for `vso` are available at <https://docs.vanillaos.org/docs/en/vso>.

## VSO as system Shell

To use VSO as your system shell, you can copy the `usr/bin/vso-os-shell` script
to your system's `/usr/bin` directory and set it as your default shell. Your
image needs to implement the `usr/bin/os-shell` script, which will expand the
`$SHELL` environment variable, this is much needed for login shells and other
flags, this also ensures that the user's default shell is respected.

Our `vso-image` already implements this script.
