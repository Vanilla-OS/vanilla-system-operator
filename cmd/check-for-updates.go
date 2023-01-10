package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2022
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/vso/core"
)

func checkUpdateUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	Check for system updates

Usage:
  	vso update-check [options]

Options:
	--help/-h		show this message

Examples:
	vso update-check
`)
	return nil
}

func NewCheckUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-check",
		Short: "Check for system updates",
		RunE:  checkUpdate,
	}
	cmd.SetUsageFunc(checkUpdateUsage)
	return cmd
}

func checkUpdate(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	fmt.Println("Checking for updates...")

	update_cmd := exec.Command("___apt___", "update")
	update_cmd.Stdin = os.Stdin
	update_cmd.Stdout = os.Stdout
	update_cmd.Stderr = os.Stderr
	if err := update_cmd.Run(); err != nil {
		return err
	}

	list_cmd := exec.Command("___apt___", "list", "--upgradable")
	list_cmd.Env = os.Environ()
	list_cmd.Env = append(list_cmd.Env, "LANG=en_US.UTF-8")
	output, err := list_cmd.Output()
	if err != nil {
		return err
	}

	packages := strings.Split(string(output), "\n")
	if len(packages) <= 2 {
		fmt.Println("Your system is up-to-date.")
		return nil
	}

	fmt.Println("--------------------------------------------")
	fmt.Println("The following packages have pending updates:")
	for _, pkg := range packages[1 : len(packages)-1] { // First and last lines are not packages
		cols := strings.Split(pkg, " ")
		pkg_name, pkg_newver, pkg_oldver := cols[0], cols[1], cols[5]
		pkg_name = strings.Split(pkg_name, "/")[0]  // Remove source info
		pkg_oldver = pkg_oldver[:len(pkg_oldver)-1] // Remove trailing "]"

		fmt.Printf("  - %s\t%s -> %s\n", pkg_name, pkg_oldver, pkg_newver)
	}
	fmt.Println("\nRun `sudo vso trigger-update --now` to upgrade your system manually.")

	return nil
}
