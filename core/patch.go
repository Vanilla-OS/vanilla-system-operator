package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// The root JSON object that holds all the patches
type RootPatch struct {
	PatchVersions []PatchVersion `json:"patchVersions"`
}

type PatchVersion struct {
	Version  float32 `json:"version"`
	Name     string  `json:"name"`
	Patches  []Patch `json:"patches"`
	Optional bool    `json:"optional"`
}

type Patch struct {
	// The format that every patch must follow, regardless of type
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`

	// Shell Patches
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`

	// Flatpak Install Patch & Flatpak Remove Patch
	Flatpak string `json:"flatpak"`

	// Flatpak Replace Patch
	OldFlatpak string `json:"old-flatpak"`
	NewFlatpak string `json:"new-flatpak"`
}

var versionedPatches []PatchVersion
var requiredPatchList []Patch
var optionalPatchList []Patch
var patchListInitialised bool
var newestVersion float32

func rawPatchVersionFilePath() (string, error) {
	currentUser, err := getRealUser()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/home/%s/.config/vso/patchVersion", currentUser), nil
}

func getPatchVersionFilePath() (string, error) {

	patchVersionFilePath, err := rawPatchVersionFilePath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(patchVersionFilePath); err != nil {
		return "", nil
	}

	return patchVersionFilePath, nil
}

func loadPatches() ([]PatchVersion, error) {
	patchFilePath := "/usr/share/vso/patch.json"

	f, err := os.Open(patchFilePath)
	if err != nil {
		return []PatchVersion{}, err
	}
	defer f.Close()

	var loadedFile RootPatch
	err = json.NewDecoder(f).Decode(&loadedFile)
	if err != nil {
		return []PatchVersion{}, err
	}

	return loadedFile.PatchVersions, nil
}

func loadPatchVersion() (float32, error) {
	var version float32

	patchVersionFilePath, err := getPatchVersionFilePath()
	if err != nil {
		return 0, err
	}

	// Load the current patch version
	if patchVersionFilePath == "" { // The file does not exist
		version = 0
	} else {
		content, err := os.ReadFile(patchVersionFilePath)
		if err != nil {
			return 0, err
		}

		version64, err := strconv.ParseFloat(string(content), 32)
		if err != nil {
			return 0, err
		}
		version = float32(version64)
	}

	return version, nil
}

func setPatchVersion() error {
	patchVersionFilePath, err := rawPatchVersionFilePath()
	if err != nil {
		return err
	}

	newestVersionString := strconv.FormatFloat(float64(newestVersion), 'f', -1, 32)
	return os.WriteFile(patchVersionFilePath, []byte(newestVersionString), 0644)
}

// GetPatches aggregates all the patches that have to be applied into the final patch lists, and returns the first and last patches to be applied
func GetPatches() (PatchVersion, PatchVersion, error) {
	fullPatchList, err := loadPatches()
	if err != nil {
		return PatchVersion{}, PatchVersion{}, err
	}

	currentPatchVersion, err := loadPatchVersion()
	if err != nil {
		return PatchVersion{}, PatchVersion{}, err
	}

	newestVersion = fullPatchList[0].Version

	versionedPatches = fullPatchList
	for i, v := range fullPatchList {
		if v.Version <= currentPatchVersion {
			versionedPatches = fullPatchList[0:i]
			break
		}
	}

	for _, v := range versionedPatches {
		for _, w := range v.Patches {
			if !validatePatch(w) {
				continue
			}
			if w.Optional {
				optionalPatchList = append(optionalPatchList, w)
			} else {
				requiredPatchList = append(requiredPatchList, w)
			}
		}
	}

	patchListInitialised = true
	return fullPatchList[0], fullPatchList[len(fullPatchList)-1], nil
}

func validatePatch(patch Patch) bool {
	switch patch.Type {
	case "shell":
		return true
	case "flatpak-install":
		return true
	case "flatpak-remove":
		return validateFlatpakRemove(patch)
	case "flatpak-replace":
		return validateFlatpakReplace(patch)
	default:
		return false
	}
}

func validateFlatpakRemove(patch Patch) bool {
	err := exec.Command("flatpak", "info", patch.Flatpak).Run()
	return err == nil
}

func validateFlatpakReplace(patch Patch) bool {
	err := exec.Command("flatpak", "info", patch.OldFlatpak).Run()
	if err != nil {
		return false
	}
	err = exec.Command("flatpak", "info", patch.NewFlatpak).Run()
	if err == nil {
		return false
	}
	return true
}

func GetOptionalPatches() ([]Patch, error) {
	if !patchListInitialised {
		return []Patch{}, fmt.Errorf("Patch list is requested before it is initialised")
	}

	return optionalPatchList, nil
}

func GetRequiredPatches() ([]Patch, error) {
	if !patchListInitialised {
		return []Patch{}, fmt.Errorf("Patch list is requested before it is initialised")
	}

	return requiredPatchList, nil
}

//TODO: Restore previous application state in case of failed patch application

func AsyncApplyPatches(statusChannel chan int, completeChannel chan bool, errorChannel chan error, optionalPatchConsent ...bool) {
	if len(optionalPatchConsent) != len(optionalPatchList) {
		errorChannel <- fmt.Errorf("Optional Patch Consent list is malformed")
		return
	}

	totalTransactions := 0

	for _, v := range requiredPatchList {
		totalTransactions += 1
		statusChannel <- totalTransactions

		switch v.Type {
		case "shell":
			err := applyShellPatch(v)
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-install":
			err := applyFlatpakInstallPatch(v)
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-remove":
			err := applyFlatpakRemovePatch(v)
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-replace":
			err := applyFlatpakReplacePatch(v)
			if err != nil {
				errorChannel <- err
			}
		default:
			errorChannel <- fmt.Errorf("Invalid Patch Type: %s", v.Type)
		}

		statusChannel <- -1
	}

	for i, v := range optionalPatchList {
		if !optionalPatchConsent[i] {
			continue
		}
		totalTransactions += 1

		statusChannel <- totalTransactions

		switch v.Type {
		case "shell":
			err := applyShellPatch(v)
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-install":
			err := applyFlatpakInstallPatch(v)
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-remove":
			err := applyFlatpakRemovePatch(v)
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-replace":
			err := applyFlatpakReplacePatch(v)
			if err != nil {
				errorChannel <- err
			}
		default:
			errorChannel <- fmt.Errorf("Invalid Patch Type: %s", v.Type)
		}
		statusChannel <- -1
	}
	err := setPatchVersion()
	if err != nil {
		errorChannel <- err
	}
	completeChannel <- true
}

func applyShellPatch(patch Patch) error {
	err := exec.Command(patch.Command, patch.Arguments...).Run()
	return err
}

func applyFlatpakInstallPatch(patch Patch) error {
	err := exec.Command("flatpak", "install", "-y", patch.Flatpak).Run()
	return err
}

func applyFlatpakRemovePatch(patch Patch) error {
	err := exec.Command("flatpak", "remove", "-y", patch.Flatpak).Run()
	return err
}

func applyFlatpakReplacePatch(patch Patch) error {
	err := exec.Command("flatpak", "remove", "-y", patch.OldFlatpak).Run()
	if err != nil {
		return err
	}
	err = exec.Command("flatpak", "install", "-y", patch.NewFlatpak).Run()
	return err
}
