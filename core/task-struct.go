package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Task struct {
	Name        string
	Slug        string
	Description string
	NeedConfirm bool
	Command     string

	AfterTask        string
	AfterTaskSuccess string
	AfterTaskFailure string
	Every            string
	At               string
	//OnBoot               bool
	OnNetwork            bool
	OnDisconnect         bool
	OnBattery            bool
	OnLowBattery         bool
	OnCharge             bool
	OnFullBattery        bool
	OnConditionCommand   string
	OnProcess            string
	OnInternetUsage      int
	OnHighInternetUsage  bool
	OnMemoryUsage        int
	OnHighMemoryUsage    bool
	OnCPUUsage           int
	OnHighCPUUsage       bool
	OnCPUTemp            int
	OnDeviceConnected    string
	OnDeviceDisconnected string

	LastExecution       time.Time
	LastExecutionOutput string
}

type TaskJson struct {
	Task
	Dependencies []Task `json:"Dependencies"`
	Relations    []Task `json:"Relations"`
	Target       string `json:"Target"`
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

// ShouldRun checks if a task respect the assigned event/condition
func (t *Task) ShouldRun(cChecks *CommonChecks, event string) bool {
	fmt.Printf("Task: %s\n", t.Name)
	fmt.Printf("| Check if task should run\n")
	res := false
	target := ""

	if t.AfterTask != "" && TaskHasRun(t.AfterTask) {
		res = true
		target = "after task " + t.AfterTask
	} else if t.AfterTaskSuccess != "" && TaskHasRunSuccess(t.AfterTaskSuccess) {
		res = true
		target = "after task success " + t.AfterTaskSuccess
	} else if t.AfterTaskFailure != "" && TaskHasRunFail(t.AfterTaskFailure) {
		res = true
		target = "after task failure " + t.AfterTaskFailure
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
	} else if t.OnInternetUsage > 0 && cChecks.InternetUsage > t.OnInternetUsage {
		res = true
		target = "internet usage: " + strconv.Itoa(cChecks.InternetUsage) + " > " + strconv.Itoa(t.OnInternetUsage)
	} else if t.OnHighInternetUsage && cChecks.HighInternetUsage {
		res = true
		target = "high internet usage"
	} else if t.OnMemoryUsage > 0 && cChecks.MemoryUsage > t.OnMemoryUsage {
		res = true
		target = "memory usage: " + strconv.Itoa(cChecks.MemoryUsage) + " > " + strconv.Itoa(t.OnMemoryUsage)
	} else if t.OnHighMemoryUsage && cChecks.HighMemoryUsage {
		res = true
		target = "high memory usage"
	} else if t.OnCPUUsage > 0 && cChecks.CPUUsage > t.OnCPUUsage {
		res = true
		target = "cpu usage: " + strconv.Itoa(cChecks.CPUUsage) + " > " + strconv.Itoa(t.OnCPUUsage)
	} else if t.OnHighCPUUsage && cChecks.HighCPUUsage {
		res = true
		target = "high cpu usage"
	} else if t.OnCPUTemp > 0 && cChecks.CPUTemp > t.OnCPUTemp {
		res = true
		target = "cpu temp: " + strconv.Itoa(cChecks.CPUTemp) + " > " + strconv.Itoa(t.OnCPUTemp)
	} else if t.OnDeviceConnected != "" {
		if deviceIsConnected(t.OnDeviceConnected) {
			res = true
			target = "device connected: " + t.OnDeviceConnected
		} else {
			t.SaveLastFailure()
		}
	} else if t.OnDeviceDisconnected != "" {
		if !deviceIsConnected(t.OnDeviceDisconnected) {
			res = true
			target = "device disconnected: " + t.OnDeviceDisconnected
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

// deviceIsConnected checks if a device is connected
func deviceIsConnected(device string) bool {
	cmd := exec.Command("sh", "-c", "lsusb | grep -i '"+device+"'")
	cmd.Env = os.Environ()
	err := cmd.Run()
	return err == nil
}

// Target returns the target of the task
func (t *Task) Target() string {
	if t.AfterTask != "" {
		return "afterTask"
	} else if t.AfterTaskSuccess != "" {
		return "afterTaskSuccess"
	} else if t.AfterTaskFailure != "" {
		return "afterTaskFailure"
	} else if t.Every != "" {
		return "every"
	} else if t.At != "" {
		return "at"
	} else if t.OnNetwork {
		return "onNetwork"
	} else if t.OnDisconnect {
		return "onDisconnect"
	} else if t.OnBattery {
		return "onBattery"
	} else if t.OnLowBattery {
		return "onLowBattery"
	} else if t.OnCharge {
		return "onCharge"
	} else if t.OnFullBattery {
		return "onFullBattery"
	} else if t.OnConditionCommand != "" {
		return "onConditionCommand"
	} else if t.OnProcess != "" {
		return "onProcess"
	} else if t.OnInternetUsage > 0 {
		return "onInternetUsage"
	} else if t.OnHighInternetUsage {
		return "onHighInternetUsage"
	} else if t.OnMemoryUsage > 0 {
		return "onMemoryUsage"
	} else if t.OnHighMemoryUsage {
		return "onHighMemoryUsage"
	} else if t.OnCPUUsage > 0 {
		return "onCPUUsage"
	} else if t.OnHighCPUUsage {
		return "onHighCPUUsage"
	} else if t.OnCPUTemp > 0 {
		return "onCPUTemp"
	}

	return ""
}

// Save saves a task in a vsotask file
func (t *Task) Save() error {
	t.Slug = slugify(t.Name)

	if t.AfterTask != "" {
		t.Slug = t.AfterTask + "-" + t.Slug
	} else if t.AfterTaskSuccess != "" {
		t.Slug = t.AfterTaskSuccess + "-" + t.Slug
	} else if t.AfterTaskFailure != "" {
		t.Slug = t.AfterTaskFailure + "-" + t.Slug
	}

	file, err := os.Create(getUserTasksLocation() + "/" + t.Slug + ".vsotask")
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

	file, err := os.Create("/tmp/" + t.Slug + ".vsotask.success")
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

	err := os.Remove("/tmp/" + t.Slug + ".vsotask.success")
	if err != nil {
		return err
	}

	return nil
}

// WasSuccessful checks if a task was successful
func (t *Task) WasSuccessful() bool {
	_, err := os.Stat("/tmp/" + t.Slug + ".vsotask.success")
	return err == nil
}

// SaveLastFailure saves the last failure of a task in /tmp
func (t *Task) SaveLastFailure() error {
	fmt.Println("| Saving last failure status")

	t.RemoveRunning()
	t.RemoveLastSuccess()

	file, err := os.Create("/tmp/" + t.Slug + ".vsotask.failure")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	os.Remove("/tmp/" + t.Slug + ".vsotask.success")

	return nil
}

// RemoveLastFailure removes the last failure of a task in /tmp
func (t *Task) RemoveLastFailure() error {
	fmt.Println("| Removing last failure status")

	err := os.Remove("/tmp/" + t.Slug + ".vsotask.failure")
	if err != nil {
		return err
	}

	return nil
}

// SaveRunning saves the running state of a task in /tmp
func (t *Task) SaveRunning() error {
	fmt.Println("| Saving running status")

	file, err := os.Create("/tmp/" + t.Slug + ".vsotask.running")
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
	_, err := os.Stat("/tmp/" + t.Slug + ".vsotask.running")
	return err == nil
}

// RemoveRunning removes the running state of a task in /tmp
func (t *Task) RemoveRunning() error {
	fmt.Println("| Removing running status")

	err := os.Remove("/tmp/" + t.Slug + ".vsotask.running")
	if err != nil {
		return err
	}

	return nil
}

// WasFailure checks if a task was a failure
func (t *Task) WasFailure() bool {
	_, err := os.Stat("/tmp/" + t.Slug + ".vsotask.failure")
	return err == nil
}

// Relations returns a list of Task which depends on the current one
func (t *Task) Relations() []Task {
	tasks, err := ListTasksDetailed()
	if err != nil {
		return []Task{}
	}

	var relations []Task
	for _, task := range tasks {
		if task.AfterTask == t.Name || task.AfterTaskSuccess == t.Name || task.AfterTaskFailure == t.Name {
			relations = append(relations, task)
		}
	}

	return relations
}

// Dependencies returns a list of Task which the current one depends on
func (t *Task) Dependencies() []Task {
	if t.AfterTask == "" && t.AfterTaskSuccess == "" && t.AfterTaskFailure == "" {
		return []Task{}
	}

	tasks, err := ListTasksDetailed()
	if err != nil {
		return []Task{}
	}

	var dependencies []Task
	for _, task := range tasks {
		if task.Name == t.AfterTask || task.Name == t.AfterTaskSuccess || task.Name == t.AfterTaskFailure {
			dependencies = append(dependencies, task)
		}
	}

	return dependencies
}
