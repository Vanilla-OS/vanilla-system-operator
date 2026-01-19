//go:build !check_missing_strings

package main

/*	License: GPLv3
	Authors:
		Mirko Brombin <brombin94@gmail.com>
		Pietro di Caprio <pietro@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/vanilla-os/apx/v2/core"
	"github.com/vanilla-os/apx/v2/settings"
	"github.com/vanilla-os/sdk/pkg/v1/app"
	"github.com/vanilla-os/sdk/pkg/v1/app/types"
	cmd "github.com/vanilla-os/vanilla-system-operator/internal/cli"
	vsoSettings "github.com/vanilla-os/vanilla-system-operator/settings"
)

var Version = "development"

//go:embed locales/*
var embeddedLocales embed.FS

func main() {
	var err error

	cnf := settings.NewApxConfig(
		"/usr/share/vso/apx",
		"/usr/share/apx/distrobox/distrobox",
		"btrfs",
	)
	core.NewApx(cnf)

	core.NewApx(cnf)

	err = vsoSettings.LoadConfig()
	if err != nil {
		fmt.Println("Warning: Failed to load config:", err)
	}

	subFS, err := fs.Sub(embeddedLocales, "locales")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd.VSO, err = app.NewApp(types.AppOptions{
		Name:          "vso",
		Version:       Version,
		RDNN:          "org.vanillaos.vso",
		LocalesFS:     subFS,
		DefaultLocale: "en",
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rootCmdStruct := &cmd.RootCmd{
		Version: Version,
	}

	err = cmd.VSO.WithCLI(rootCmdStruct)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd.VSO.CLI.SetTranslator(func(key string) string {
		return cmd.VSO.LC.Get(key)
	})

	err = cmd.VSO.CLI.Execute()
	if err != nil {
		cmd.VSO.Log.Error(err.Error())
		os.Exit(1)
	}
}
