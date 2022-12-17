package core

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

func AskConfirmation(s string) bool {
	var response string
	fmt.Print(s + " [y/N]: ")
	fmt.Scanln(&response)
	if response == "y" || response == "Y" {
		return true
	}
	return false
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
