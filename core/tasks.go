package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	Name                string
	Description         string
	NeedConfirm         bool
	Command             string
	Every               string
	At                  string
	OnBoot              bool
	OnNetwork           bool
	OnDisconnect        bool
	OnBattery           bool
	OnLowBattery        bool
	OnCharge            bool
	OnFullBattery       bool
	OnConditionCommand  string
	OnProcess           string
	LastExecution       time.Time
	LastExecutionOutput string
}

var (
	tasksLocation = "/.config/vso/tasks"
)

// List lists all tasks
func List() ([]string, error) {
	files, err := os.ReadDir(getUserTasksLocation())
	if err != nil {
		return nil, err
	}

	var list []string
	for _, file := range files {
		list = append(list, file.Name())
	}

	return list, nil
}

// ListDetailed lists all tasks with detailed information
func ListDetailed() ([]Task, error) {
	list, err := List()
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for _, name := range list {
		name = name[:len(name)-len(".vsotask")]
		task, err := Load(name)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	return tasks, nil
}

// Delete deletes a task
func Delete(name string) error {
	err := os.Remove(getUserTasksLocation() + "/" + name + ".vsotask")
	if err != nil {
		return err
	}

	return nil
}

// Run runs a task
func Run(name string) error {
	t, err := Load(name)
	if err != nil {
		return err
	}

	err = t.Run()
	if err != nil {
		return err
	}

	return nil
}

// Run runs a task
func (t *Task) Run() error {
	if t.IsRunning() {
		fmt.Println("| Not running since it is already running")
		return nil
	}

	err := t.SaveRunning()
	if err != nil {
		return err
	}

	if t.NeedConfirm {
		if !ConfirmWindow("Task '"+t.Name+"' wants to run", t.Description) {
			err = t.RemoveRunning()
			if err != nil {
				return err
			}
			return nil
		}
	}

	cmd := exec.Command("bash", "-c", t.Command)
	output, err := cmd.CombinedOutput()
	fmt.Printf("| Command: %s\n", t.Command)
	fmt.Printf("| Output: %s\n", output)
	if err != nil {
		t.SaveLastFailure()
	}

	t.LastExecution = time.Now()
	t.LastExecutionOutput = string(output)

	err = t.Save()
	if err != nil {
		return err
	}

	t.SaveLastSuccess()

	return nil
}

// Rotate checks which tasks should be run and runs them
func Rotate(event string) error {
	if rotatorIsRunning() {
		fmt.Println("Rotator is already running")
		return nil
	}

	err := TasksInit()
	if err != nil {
		return err
	}

	err = saveRotatorRunning()
	if err != nil {
		return err
	}

	for {
		fmt.Println("---")
		cChecks := GetCommonChecks()
		err := runRotator(cChecks, event)
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(5 * time.Second)
	}
}

func runRotator(cChecks *CommonChecks, event string) error {
	list, err := List()
	if err != nil {
		return err
	}

	rotatedN := 0
	rotatedNFails := 0
	rotatedNSuccess := 0

	for _, name := range list {
		t, err := Load(name)
		if err != nil {
			rotatedNFails++
			return err
		}

		if t.ShouldRun(cChecks, event) {
			err = t.Run()
			if err != nil {
				rotatedNFails++
				return err
			}
		}

		rotatedN++
		rotatedNSuccess++
	}

	fmt.Printf("Rotated %d tasks, %d failed, %d success\n", rotatedN, rotatedNFails, rotatedNSuccess)

	return nil
}

// saveRotatorRunning saves that the rotator is running
func saveRotatorRunning() error {
	err := os.WriteFile("/tmp/vso-rotator-running", []byte("true"), 0644)
	if err != nil {
		return err
	}

	return nil
}

// removeRotatorRunning removes that the rotator is running
func removeRotatorRunning() error {
	err := os.Remove("/tmp/vso-rotator-running")
	if err != nil {
		return err
	}

	return nil
}

// rotatorIsRunning checks if the rotator is running
func rotatorIsRunning() bool {
	_, err := os.Stat("/tmp/vso-rotator-running")
	running := !os.IsNotExist(err)

	if running {
		if !processIsRunning("rotate-tasks", true) {
			running = false
		}
	}

	if !running {
		err := removeRotatorRunning()
		if err != nil {
			fmt.Println(err)
		}
	}

	return running
}

// ShouldRun checks if a task respect the assigned event/condition
func (t *Task) ShouldRun(cChecks *CommonChecks, event string) bool {
	fmt.Printf("Task: %s\n", t.Name)
	fmt.Printf("| Check if task should run\n")
	res := false
	target := ""

	if t.OnBoot && event == "boot" {
		res = true
		target = "boot"
	} else if t.Every != "" && ItsBeen(t.LastExecution, t.Every) {
		res = true
		target = "every " + t.Every
	} else if t.At != "" && ItsTime(t.At) {
		res = true
		target = "at " + t.At
	} else if t.OnNetwork && cChecks.Network {
		res = true
		target = "network"
	} else if t.OnDisconnect && !cChecks.Network {
		res = true
		target = "disconnect"
	} else if t.OnBattery && cChecks.Battery {
		res = true
		target = "battery"
	} else if t.OnLowBattery && cChecks.LowBattery {
		res = true
		target = "low battery"
	} else if t.OnCharge && !cChecks.Battery {
		res = true
		target = "charge"
	} else if t.OnFullBattery && cChecks.FullBattery {
		res = true
		target = "full battery"
	} else if t.OnConditionCommand != "" {
		cmd := exec.Command("sh", "-c", t.OnConditionCommand)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "VSO_TASK_NAME="+t.Name)
		err := cmd.Run()
		if err == nil {
			res = true
			target = "condition command: " + t.OnConditionCommand
		}
	} else if t.OnProcess != "" {
		if processIsRunning(t.OnProcess, true) {
			res = true
			target = "process: " + t.OnProcess
		} else {
			t.SaveLastFailure()
		}
	}
	if res {
		fmt.Printf("| Condition reached: %s\n", target)

		if t.WasSuccessful() {
			res = false
			fmt.Printf("| Not running due to last success\n")
		}
	} else {
		fmt.Printf("| Condition not reached, task will not run\n")
	}

	return res
}

// Save saves a task in a vsotask file
func (t *Task) Save() error {
	file, err := os.Create(getUserTasksLocation() + "/" + t.Name + ".vsotask")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	err = encoder.Encode(t)
	if err != nil {
		return err
	}

	return nil
}

// Unit returns the unit name of a task
func (t *Task) Unit() string {
	return t.Name + ".vsotask"
}

// Delete deletes a task
func (t *Task) Delete() error {
	err := os.Remove(getUserTasksLocation() + "/" + t.Name + ".vsotask")
	if err != nil {
		return err
	}

	return nil
}

// SaveLastSuccess saves the last success of a task in /tmp
func (t *Task) SaveLastSuccess() error {
	fmt.Println("| Saving last success status")

	t.RemoveRunning()
	t.RemoveLastFailure()

	file, err := os.Create("/tmp/" + t.Name + ".vsotask.success")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	os.Remove("/tmp/" + t.Name + ".vsotask.failure")

	return nil
}

// RemoveLastSuccess removes the last success of a task in /tmp
func (t *Task) RemoveLastSuccess() error {
	fmt.Println("| Removing last success status")

	err := os.Remove("/tmp/" + t.Name + ".vsotask.success")
	if err != nil {
		return err
	}

	return nil
}

// WasSuccessful checks if a task was successful
func (t *Task) WasSuccessful() bool {
	_, err := os.Stat("/tmp/" + t.Name + ".vsotask.success")
	return err == nil
}

// SaveLastFailure saves the last failure of a task in /tmp
func (t *Task) SaveLastFailure() error {
	fmt.Println("| Saving last failure status")

	t.RemoveRunning()
	t.RemoveLastSuccess()

	file, err := os.Create("/tmp/" + t.Name + ".vsotask.failure")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	os.Remove("/tmp/" + t.Name + ".vsotask.success")

	return nil
}

// RemoveLastFailure removes the last failure of a task in /tmp
func (t *Task) RemoveLastFailure() error {
	fmt.Println("| Removing last failure status")

	err := os.Remove("/tmp/" + t.Name + ".vsotask.failure")
	if err != nil {
		return err
	}

	return nil
}

// SaveRunning saves the running state of a task in /tmp
func (t *Task) SaveRunning() error {
	fmt.Println("| Saving running status")

	file, err := os.Create("/tmp/" + t.Name + ".vsotask.running")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	return nil
}

// IsRunning checks if a task is running
func (t *Task) IsRunning() bool {
	_, err := os.Stat("/tmp/" + t.Name + ".vsotask.running")
	return err == nil
}

// RemoveRunning removes the running state of a task in /tmp
func (t *Task) RemoveRunning() error {
	fmt.Println("| Removing running status")

	err := os.Remove("/tmp/" + t.Name + ".vsotask.running")
	if err != nil {
		return err
	}

	return nil
}

// WasFailure checks if a task was a failure
func (t *Task) WasFailure() bool {
	_, err := os.Stat("/tmp/" + t.Name + ".vsotask.failure")
	return err == nil
}

// Load loads a task
func Load(name string) (*Task, error) {
	if !strings.HasSuffix(name, ".vsotask") {
		name = name + ".vsotask"
	}

	file, err := os.Open(getUserTasksLocation() + "/" + name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	var t Task
	err = decoder.Decode(&t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// makeAutostart creates an autostart file for vso if it doesn't exist
func makeAutostart() error {
	curUser, err := getRealUser()
	if err != nil {
		fmt.Println("Cannot determine real user. Autostart will not be created.")
		return err
	}

	autostartFolder := "/home/" + curUser + "/.config/autostart"
	autostartFile := autostartFolder + "/vso.desktop"

	if _, err := os.Stat(autostartFile); os.IsNotExist(err) {
		fmt.Println("Creating vso autostart file for user in " + autostartFile)

		err = os.MkdirAll(autostartFolder, 0755)
		if err != nil {
			return err
		}

		file, err := os.Create(autostartFile)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.WriteString("[Desktop Entry]\nType=Application\nName=vso\nExec=/usr/bin/vso rotate-tasks\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// makeTasksLocation creates the tasks location for the user if it doesn't exist
func makeTasksLocation() error {
	userTasksLocation := getUserTasksLocation()
	if _, err := os.Stat(userTasksLocation); os.IsNotExist(err) {
		fmt.Println("Creating tasks location for user in " + userTasksLocation)

		err = os.MkdirAll(userTasksLocation, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// TasksInit calls makeTasksLocation and makeAutostart in one call
func TasksInit() error {
	err := makeTasksLocation()
	if err != nil {
		return err
	}

	err = makeAutostart()
	if err != nil {
		return err
	}

	return nil
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

// getUserTasksLocation returns the location of the user tasks
func getUserTasksLocation() string {
	curUser, err := getRealUser()
	if err != nil {
		return ""
	}

	return "/home/" + curUser + tasksLocation
}
