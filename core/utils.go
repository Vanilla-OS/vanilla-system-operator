package core

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
