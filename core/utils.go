package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
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

func AskConfirmation(s string, norm bool) bool {
	var response string
	var defResponse string
	if norm {
		fmt.Print(s + " [Y/n]: ")
		defResponse = "y"
	} else {
		fmt.Print(s + " [y/N]: ")
		defResponse = "n"
	}
	fmt.Scanln(&response)
	if strings.ToLower(response) != strings.ToLower(defResponse) && len(strings.TrimSpace(response)) != 0 {
		return !norm
	}
	return norm
}

func PickOption(s string, a []string, def int) int {
	var response int
	selected := -1
	for selected > len(a) || selected < 0 {
		selected = def
		for i, opt := range a {
			fmt.Printf("%d) %s\n", i+1, opt)
		}
		fmt.Printf("\n%s", s)
		if def > -1 {
			fmt.Printf(" (%d): ", def)
		}
		fmt.Scanln(&response)
		if response != 0 {
			selected = response
		}
	}
	return selected - 1
}

func CheckConnection() bool {
	_, err := http.Get("https://google.com") // TODO: use a better way to check connection
	return err == nil
}

func SendNotification(title, body string) error {
	cmd := exec.Command(
		"notify-send",
		title, body,
		"-i", "distributor-logo-vanilla",
	)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
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
	cmd := exec.Command("ps", "-A", "-o", "pid,ppid,cmd")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, name) {
			if excludeVsoPid {
				vsoPid := strconv.Itoa(os.Getppid())
				if strings.Contains(line, vsoPid) {
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
