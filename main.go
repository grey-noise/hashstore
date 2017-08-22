package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/minio/cli"
)

var i int64
var db *bolt.DB
var hasher hash.Hash
var bucketName string

func hashcode(path string, info os.FileInfo, err error) error {
	hasher.Reset()
	if err != nil {
		log.Print(err)
		return nil
	}
	if info.IsDir() || info.Name() == "pa.db.lock" {
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
		return nil
	}

	hasher.Write(data)
	sha := hasher.Sum(nil)

	i++

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		b.Put([]byte(path), sha)
		return nil
	})
	return nil
}

func main() {
	app := cli.NewApp()
	app.Version = "19.99.0"

	app.Commands = []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "start a hash",
			Action:  startHash,
		},
		{
			Name:    "display",
			Aliases: []string{"h"},
			Usage:   "display  a run",
			Action:  display,
		},
		{
			Name:    "list",
			Aliases: []string{"a"},
			Usage:   "list all the runs ",
			Action:  lists_run,
		},
	}

	app.Run(os.Args)
}

func startHash(c *cli.Context) {
	dir := c.Args().First()

	i = 0
	var err error
	if err != nil {
		panic(err)
	}

	db, err := bolt.Open("pa.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	t := time.Now().Local()

	bucketName = fmt.Sprintf("%s://%s://%s", name, dir, t.Format("2006-01-01"))
	fmt.Println(bucketName)
	hasher = sha512.New()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	err = filepath.Walk(dir, hashcode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d files treated", i)
}

func Display(c *cli.Context) {
	bucketName := c.Args().First()

	db, err := bolt.Open("pa.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("file=%s, hash=%s\n", k, hex.EncodeToString(v))
			return nil
		})
		return nil
	})

}

func lists_run(c *cli.Context) {

	db, err := bolt.Open("pa.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {

		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			fmt.Println(string(name))
			return nil
		})
	})
	if err != nil {
		fmt.Println(err)
		return
	}
}
