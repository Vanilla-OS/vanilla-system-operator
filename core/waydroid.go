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
	"strings"

	"github.com/vanilla-os/apx/v2/core"
	bolt "go.etcd.io/bbolt"
)

type BucketNotFoundError struct {
	BucketName string
	Err        error
}

func (b *BucketNotFoundError) Error() string {
	return fmt.Sprintf("Buckent %s not found", b.BucketName)
}

var DatabasePath = fmt.Sprintf("%s/.local/share/vso/waydroid/droidapps.db", os.Getenv("HOME"))

func GetWay() (*core.SubSystem, error) {
	subsystem, err := core.LoadSubSystem("vso-waydroid", true)
	if err != nil {
		return nil, err
	}

	return subsystem, nil
}

func EnsureWayStarted() error {
	subsystem, err := GetWay()
	if err != nil {
		return err
	}

	finalArgs := []string{
		"ewaydroid",
		"status",
	}
	subsystem.Exec(true, false, finalArgs...)
	out, err := subsystem.Exec(true, false, finalArgs...)
	if err != nil {
		return err
	}

	if strings.Contains(out, "STOPPED") {
		fmt.Println("Waydroid session is stopped, starting...")
		WayLXCStart()
		WaySessionStart()
	}

	return nil
}

func GetWayDatabase() (*bolt.DB, error) {
	_, exist := os.Stat(DatabasePath)
	if os.IsNotExist(exist) {
		os.MkdirAll(strings.Replace(DatabasePath, "/droidapps.db", "", 1), 0777)
		os.Create(DatabasePath)
	}
	db, err := bolt.Open(DatabasePath, 0777, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func WayGetAppFromDatabase(rdnsname string, db *bolt.DB) (FdroidPackage, error) {
	var err error
	var deferDB = false
	if db == nil {
		db, err = GetWayDatabase()
		deferDB = true
	}
	if err != nil {
		return FdroidPackage{}, err
	}
	var pkg FdroidPackage
	err = db.View(func(tx *bolt.Tx) error {
		apps := tx.Bucket([]byte("Apps"))
		app := apps.Get([]byte(rdnsname))
		err := json.Unmarshal(app, &pkg)
		return err
	})
	if deferDB {
		defer db.Close()
	}
	return pkg, err
}

func WayPutAppIntoDatabase(pkg FdroidPackage, db *bolt.DB) error {
	var err error
	var deferDB = false
	if db == nil {
		db, err = GetWayDatabase()
		deferDB = true
	}
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		apps, err := tx.CreateBucketIfNotExists([]byte("Apps"))
		if err != nil {
			return err
		}
		if buf, err := json.Marshal(pkg); err != nil {
			return err
		} else if err := apps.Put([]byte(pkg.RDNSName), buf); err != nil {
			return err
		}
		return nil
	})
	if deferDB {
		defer db.Close()
	}
	return err
}

func WayRemoveAppFromDatabase(rdnsname string, db *bolt.DB) error {
	fmt.Printf("removing %s\n", rdnsname)
	var err error
	var deferDB = false
	if db == nil {
		db, err = GetWayDatabase()
		deferDB = true
	}
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		apps := tx.Bucket([]byte("Apps"))
		if apps == nil {
			return &BucketNotFoundError{
				BucketName: "Apps",
			}
		}
		fmt.Println("here2")
		return apps.Delete([]byte(rdnsname))
	})
	if deferDB {
		defer db.Close()
	}
	return err
}

func WayExists() bool {
	_, err := GetWay()
	return err == nil
}

func WayInit() error {
	stack, _ := core.LoadStack("vso-waydroid")
	subsystem, err := core.NewSubSystem(
		"vso-waydroid",
		stack,
		"",
		true,
		true,
		true,
		false,
		false,
	)

	if err != nil {
		return err
	}

	err = subsystem.Create()
	if err != nil {
		return err
	}

	way, err := GetWay()
	if err != nil {
		return err
	}

	finalArgs := []string{"sudo", "ewaydroid", "init"}
	way.Exec(false, false, finalArgs...)
	_, err = way.Exec(false, false, finalArgs...)
	// oh, the above? yeah, it's a hack. I don't know why it doesn't work the
	// first time, but it doesn't. I guess it's fine for now

	// we need to manage the waydroid session ourselves and since the waydroid
	// method to stop the session is not reliable, we need to stop the whole
	// container
	way.Stop()

	return err
}

func WayLXCStart() error {
	WayLXCStop() // ensure even systemd service is stopped

	way, err := GetWay()
	if err != nil {
		return err
	}

	finalArgs := []string{"sudo", "ewaydroid", "container", "start"}
	_, err = way.Exec(false, true, finalArgs...)
	return err
}

func WayLXCStop() error {
	way, err := GetWay()
	if err != nil {
		return err
	}

	finalArgs := []string{"sudo", "systemctl", "stop", "waydroid-container.service"}
	way.Exec(false, false, finalArgs...) // we do not really care if it fails

	finalArgs = []string{"sudo", "ewaydroid", "container", "stop"}
	_, err = way.Exec(false, false, finalArgs...)
	return err
}

func WaySessionStart() error {
	way, err := GetWay()
	if err != nil {
		return err
	}

	finalArgs := []string{"ewaydroid", "session", "start"}
	_, err = way.Exec(false, true, finalArgs...)
	return err
}

func WayDelete() error {
	subsystem, err := GetWay()
	if err != nil {
		return err
	}

	err = subsystem.Remove()
	if err != nil {
		return err
	}

	return nil
}
