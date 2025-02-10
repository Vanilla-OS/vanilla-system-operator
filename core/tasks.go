package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2024
	Description: The Vanilla System Operator is a package manager,
	a system updater and a task automator.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	err := os.Remove(getUserTasksLocation() + "/" + strings.ToLower(name) + ".vsotask")
	if err != nil {
		return fmt.Errorf("task does not exist with unit file name: %s", name)
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
func RotateTasks(event string, silent bool) error {
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
		if !silent {
			fmt.Println("---")
		}
		cChecks := GetCommonChecks()
		rotatedCount, err := runTasksRotator(cChecks, event, silent)
		if err != nil {
			fmt.Println(err)
		} else if rotatedCount == 0 {
			// avoid running constantly if no tasks are installed
			time.Sleep(60 * time.Second)
		}

		time.Sleep(5 * time.Second)
	}
}

// runTasksRotator checks which tasks should be run and starts rotating
func runTasksRotator(cChecks *CommonChecks, event string, silent bool) (int, error) {
	var err error
	CurrentQueue, err = ListTasksDetailed()
	if err != nil {
		return 0, err
	}

	rotatedN := 0
	rotatedNFails := 0
	rotatedNSuccess := 0

	for _, t := range CurrentQueue {
		if t.ShouldRun(cChecks, event) {
			err = t.Run()
			if err != nil {
				rotatedNFails++
				return 0, err
			}
		}

		rotatedN++
		rotatedNSuccess++
	}

	CurrentQueue = []Task{}

	if !silent {
		fmt.Printf("Rotated %d tasks, %d failed, %d success\n", rotatedN, rotatedNFails, rotatedNSuccess)
	}

	return rotatedN, nil
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

	return nil
}

// getUserTasksLocation returns the location of the user tasks
func getUserTasksLocation() string {
	curUser, err := getRealUser()
	if err != nil {
		return ""
	}

	location := filepath.Join("/home", curUser, TasksLocation)
	if _, err := os.Stat(location); os.IsNotExist(err) {
		os.MkdirAll(location, 0755)
	}

	return location
}
