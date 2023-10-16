package core

import (
	"encoding/json"
	"fmt"
	"github.com/vanilla-os/apx/core"
	bolt "go.etcd.io/bbolt"
	"os"
	"strings"
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
		true,
		true,
		true,
		true,
	)

	if err != nil {
		return err
	}

	err = subsystem.Create()
	if err != nil {
		return err
	}

	_, err = subsystem.Exec(false, "ewaydroid", "--version")
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
