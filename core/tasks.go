package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	TasksLocation = "/.config/vso/tasks"
	CurrentQueue  = []Task{}
)

// ListUnitTasks lists all tasks files
func ListUnitTasks() ([]string, error) {
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

// ListTasksDetailed lists all tasks with detailed information
func ListTasksDetailed() ([]Task, error) {
	list, err := ListUnitTasks()
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for _, name := range list {
		name = name[:len(name)-len(".vsotask")]
		task, err := LoadTaskByUnitName(name)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	return tasks, nil
}

// ListTasksJson lists all tasks with detailed information in JSON format
func ListTasksJson() (string, error) {
	tasks, err := ListTasksDetailed()
	if err != nil {
		return "", err
	}

	var tasksJson []TaskJson
	for _, task := range tasks {
		tasksJson = append(tasksJson, TaskJson{
			Task:         task,
			Dependencies: task.Dependencies(),
			Relations:    task.Relations(),
			Target:       task.Target(),
		})
	}

	json, err := json.Marshal(tasksJson)
	if err != nil {
		return "", err
	}

	return string(json), nil
}

// DeleteTaskByUnitName deletes a task
func DeleteTaskByUnitName(name string) error {
	err := os.Remove(getUserTasksLocation() + "/" + name + ".vsotask")
	if err != nil {
		return err
	}

	return nil
}

// RunTaskByUnitName runs a task
func RunTaskByUnitName(name string) error {
	t, err := LoadTaskByUnitName(name)
	if err != nil {
		return err
	}

	err = t.Run()
	if err != nil {
		return err
	}

	return nil
}

// RotateTasks checks if no other rotators are running, then performs initial
// checks and starts rotating every 5 seconds
func RotateTasks(event string) error {
	if isTasksRotatorIsRunning() {
		fmt.Println("Rotator is already running")
		return nil
	}

	err := TasksInit()
	if err != nil {
		return err
	}

	err = saveTasksRotatorRunning()
	if err != nil {
		return err
	}

	for {
		fmt.Println("---")
		cChecks := GetCommonChecks()
		err := runTasksRotator(cChecks, event)
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(5 * time.Second)
	}
}

// runTasksRotator checks which tasks should be run and starts rotating
func runTasksRotator(cChecks *CommonChecks, event string) error {
	CurrentQueue, err := ListTasksDetailed()
	if err != nil {
		return err
	}

	rotatedN := 0
	rotatedNFails := 0
	rotatedNSuccess := 0

	for _, t := range CurrentQueue {
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

	CurrentQueue = []Task{}

	fmt.Printf("Rotated %d tasks, %d failed, %d success\n", rotatedN, rotatedNFails, rotatedNSuccess)

	return nil
}

// saveTasksRotatorRunning saves that the rotator is running
func saveTasksRotatorRunning() error {
	err := os.WriteFile("/tmp/vso-rotator-running", []byte("true"), 0644)
	if err != nil {
		return err
	}

	return nil
}

// removeTasksRotatorRunning removes that the rotator is running
func removeTasksRotatorRunning() error {
	err := os.Remove("/tmp/vso-rotator-running")
	if err != nil {
		return err
	}

	return nil
}

// isTasksRotatorIsRunning checks if the rotator is running
func isTasksRotatorIsRunning() bool {
	_, err := os.Stat("/tmp/vso-rotator-running")
	running := !os.IsNotExist(err)

	if running {
		if !processIsRunning("rotate-tasks", true) {
			running = false
		}
	}

	if !running {
		err := removeTasksRotatorRunning()
		if err != nil {
			fmt.Println(err)
		}
	}

	return running
}

// TaskHasRun checks if a task has run in the current queue
func TaskHasRun(name string) bool {
	for _, task := range CurrentQueue {
		if task.Name == name {
			return true
		}
	}

	return false
}

// TaskHasRunSuccess checks if a task has run successfully in the current queue
func TaskHasRunSuccess(name string) bool {
	for _, task := range CurrentQueue {
		if task.Name == name && task.WasSuccessful() {
			return true
		}
	}

	return false
}

// TaskHasRunFail checks if a task has run unsuccessfully in the current queue
func TaskHasRunFail(name string) bool {
	for _, task := range CurrentQueue {
		if task.Name == name && task.WasFailure() {
			return true
		}
	}

	return false
}

// LoadTaskByUnitName loads a task
func LoadTaskByUnitName(name string) (*Task, error) {
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

// makeTasksRotatorAutostart creates an autostart file for vso if it doesn't exist
func makeTasksRotatorAutostart() error {
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

// TasksInit calls makeTasksLocation and makeTasksRotatorAutostart in one call
func TasksInit() error {
	err := makeTasksLocation()
	if err != nil {
		return err
	}

	err = makeTasksRotatorAutostart()
	if err != nil {
		return err
	}

	return nil
}

// getUserTasksLocation returns the location of the user tasks
func getUserTasksLocation() string {
	curUser, err := getRealUser()
	if err != nil {
		return ""
	}

	return "/home/" + curUser + TasksLocation
}
