package core

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/buger/jsonparser"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type FdroidRepo struct {
	Name       string
	Key        string
	IndexURL   string
	RepoURL    string
	PackageURL string
}

var Repositories []FdroidRepo
var Indexes []os.DirEntry
var CacheDir = fmt.Sprintf("%s/.cache/vso/indexes/", os.Getenv("HOME"))

func getRepos() error {
	configFiles, err := os.ReadDir("/etc/vso/fdroid.repos.d/")
	if err != nil {
		return err
	}
	for _, config := range configFiles {
		if !strings.Contains(config.Name(), ".toml") {
			continue
		}
		parseRepoFile("/etc/vso/fdroid.repos.d/" + config.Name())
	}
	return nil
}

func parseRepoFile(file string) {
	var repo FdroidRepo
	repoData, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	_, err = toml.Decode(string(repoData), &repo)
	if err != nil {
		_ = fmt.Errorf("ERROR: Failed to parse configuration file %s", file)
		fmt.Printf("err: %v\n", err)
	}
	Repositories = append(Repositories, repo)
}

// Download the index of a repo
// the index argument specifies which index of the Repositories variable to download
// if index is set to -1, then the indexes from all Repositories will be downloaded
func downloadIndex(index int) error {
	if index == -1 {
		fmt.Println("syncing everything")
		for i := 0; i < len(Repositories); i++ {
			err := downloadIndex(i)
			if err != nil {
				return err
			}
		}
	} else if len(Repositories)-1 < index {
		return fmt.Errorf("index out of range")
	} else {
		fmt.Printf("Downloading repo %s\n", Repositories[index].Name)
		currTime := time.Now().Unix()
		out, err := os.Create(fmt.Sprintf("%s/index-%s-%d.json", fmt.Sprintf("%s/.cache/vso/indexes/", os.Getenv("HOME")), Repositories[index].Name, currTime))
		defer out.Close()
		if err != nil {
			return err
		}
		resp, err := http.Get(Repositories[index].IndexURL)
		defer resp.Body.Close()
		if err != nil {
			return err
		}
		_, err = io.Copy(out, resp.Body)
		return err
	}
	return nil
}

func syncIndex() {
	_, err := os.Stat(CacheDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(CacheDir, 0755)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
		err = downloadIndex(-1)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		return
	}
	Indexes, err = os.ReadDir(CacheDir)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	if len(Indexes) <= 0 {
		err := downloadIndex(-1)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
		return
	}
	var indexRepos []string
	for _, index := range Indexes {
		fileName := strings.ReplaceAll(index.Name(), ".json", "")
		indexFile := strings.Split(fileName, "-")
		indexRepos = append(indexRepos, indexFile[1])
		timestamp, err := strconv.ParseInt(indexFile[len(indexFile)-1], 10, 64)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		currentTime := time.Now().Unix()
		if currentTime-timestamp >= 604800 { // TODO: Consider if a different period is more suited (currently 1 week)
			fmt.Printf("Index for repo %s outdated! Syncing...\n", indexFile[1])
			for i, repository := range Repositories {
				if repository.Name == indexFile[1] {
					err := downloadIndex(i)
					if err != nil {
						fmt.Printf("err: %v\n", err)
					}
				}
			}
		} else {
			fmt.Printf("Index for repo %s not outdated!\n", indexFile[1])
		}
	}

	for index, repository := range Repositories {
		if !slices.Contains(indexRepos, repository.Name) {
			fmt.Printf("Index for repo %s not synced! Syncing now...\n", repository.Name)
			err := downloadIndex(index)
			if err != nil {
				fmt.Printf("err: %v\n", err)
			}
		}
	}
}

func searchIndex(search string) ([][]string, error) {
	err := getRepos()
	if err != nil {
		return nil, err
	}
	syncIndex()
	var matches [][]string
	for _, repository := range Repositories {
		for _, index := range Indexes {
			if strings.Contains(index.Name(), repository.Name) {
				indexContent, err := os.ReadFile(fmt.Sprintf("%s/%s", CacheDir, index.Name()))
				if err != nil {
					return nil, err
				}
				err = jsonparser.ObjectEach(indexContent, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
					name, err := jsonparser.GetString(value, "metadata", "name", "en-US")
					if err != nil {
						fmt.Printf("err: %v\n", err)
						return err
					}
					//fmt.Println(name)
					if strings.Contains(strings.ToLower(name), strings.ToLower(search)) {
						match := []string{string(key), name}
						matches = append(matches, match)
					}
					return nil
				}, "packages")
			}
		}
	}
	return matches, nil
}

func SearchPackage(search string) error {
	matches, err := searchIndex(search)
	if err != nil {
		return err
	}
	for _, match := range matches {
		fmt.Printf("%s - %s\n", match[1], match[0])
	}
	return nil
}

/*
func FetchPackage(package string) error {
	bodyBytes, err := searchApi(package)
	jsonparser.GetString(bodyBytes, "")
	return nil
}*/
