package core

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
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/vanilla-os/sdk/pkg/v1/net"
	"github.com/vanilla-os/sdk/pkg/v1/notification"
	notificationTypes "github.com/vanilla-os/sdk/pkg/v1/notification/types"
	"github.com/vanilla-os/sdk/pkg/v1/system"
)

var ProcessPath string

func RootCheck(display bool) bool {
	if os.Geteuid() != 0 {
		if display {
			fmt.Println("You must be root to run this command.")
		}
		return false
	}
	return true
}

func CheckConnection() bool {
	return net.CheckInternetConnectivity()
}

func SendNotification(title, body string) error {
	notif := notificationTypes.NewNotification(
		"vso",
		title,
		body,
		"distributor-logo-vanilla",
		5000,
		notificationTypes.NotificationAction{},
	)
	return notification.SendNotification(notif)
}

func ConfirmWindow(title string, body string) bool {
	cmd := exec.Command(
		"/usr/bin/adwdialog",
		"--title", title,
		"--description", body,
		"--type", "question",
	)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DISPLAY=:1")
	out, err := cmd.Output()
	fmt.Println(string(out))
	fmt.Println(err)
	return err == nil
}

// getRealUser returns the real user if the current is root
func getRealUser() (string, error) {
	curUser := os.Getenv("USER")
	if curUser == "" || curUser == "root" {
		curUser = os.Getenv("SUDO_USER")
		if curUser == "" {
			fmt.Println("Could not determine current user")
			return "", fmt.Errorf("cannot determine current user")
		}
	}

	return curUser, nil
}

// processIsRunning checks if a process is running based on its name
func processIsRunning(name string, excludeVsoPid bool) bool {
	processes, err := system.GetProcessList()
	if err != nil {
		return false
	}

	for _, p := range processes {
		if strings.Contains(p.Name, name) {
			if excludeVsoPid {
				if p.PID == os.Getppid() {
					continue
				}
			}
			return true
		}
	}
	return false
}

// slugify returns a slugified string
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Replace(s, " ", "-", -1)
	s = strings.Replace(s, "_", "-", -1)
	s = strings.Replace(s, ":", "-", -1)
	s = strings.Replace(s, "/", "-", -1)
	s = strings.Replace(s, "\\", "-", -1)
	s = strings.Replace(s, ".", "-", -1)
	s = strings.Replace(s, ",", "-", -1)
	s = strings.Replace(s, ";", "-", -1)
	s = strings.Replace(s, "!", "-", -1)
	s = strings.Replace(s, "?", "-", -1)
	s = strings.Replace(s, "(", "-", -1)
	s = strings.Replace(s, ")", "-", -1)
	s = strings.Replace(s, "[", "-", -1)
	s = strings.Replace(s, "]", "-", -1)
	s = strings.Replace(s, "{", "-", -1)
	s = strings.Replace(s, "}", "-", -1)
	s = strings.Replace(s, "'", "-", -1)
	s = strings.Replace(s, "\"", "-", -1)
	s = strings.Replace(s, "`", "-", -1)
	s = strings.Replace(s, "#", "-", -1)
	s = strings.Replace(s, "$", "-", -1)
	s = strings.Replace(s, "%", "-", -1)
	s = strings.Replace(s, "^", "-", -1)
	s = strings.Replace(s, "&", "-", -1)
	s = strings.Replace(s, "*", "-", -1)
	s = strings.Replace(s, "+", "-", -1)
	s = strings.Replace(s, "=", "-", -1)
	s = strings.Replace(s, "|", "-", -1)
	s = strings.Replace(s, ">", "-", -1)
	s = strings.Replace(s, "<", "-", -1)
	s = strings.Replace(s, "~", "-", -1)

	return s
}

func CreateVsoTable(writer io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)
	table.SetColumnSeparator("┊")
	table.SetCenterSeparator("┼")
	table.SetRowSeparator("┄")
	table.SetHeaderLine(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(true)

	return table
}
