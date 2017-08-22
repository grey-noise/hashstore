package main

import (
	"crypto/sha512"
	"fmt"
	"hash"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/minio/cli"
)

var i int64
var db *bolt.DB
var hasher hash.Hash

func hashcode(path string, info os.FileInfo, err error) error {
	hasher.Reset()
	if err != nil {
		log.Print(err)
		return nil
	}
	if info.IsDir() {
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
	fmt.Println(db)
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("file2"))
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
			Action:  starthash,
		},
		{
			Name:    "terminate",
			Aliases: []string{"h"},
			Usage:   "terminate a task",
			Action: func(c *cli.Context) error {
				if !c.Args().Present() {
					//return errors.New("no arhime")
					os.Exit(2)
				}
				fmt.Println("terminate task ", c.Args().First())
				return nil
			},
		},
		{
			Name:    "list",
			Aliases: []string{"a"},
			Usage:   "list all the tasks",
			Action: func(c *cli.Context) error {
				fmt.Println("list all the task ")
				return nil
			},
		},
	}
	var hostname string
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "hostname",
			Value:       "http://0.0.0.0:12",
			Usage:       "Remote Hostname",
			Destination: &hostname,
		},
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
		},
	}
	app.Run(os.Args)
}

func starthash(c *cli.Context) {
	dir := c.Args().First()

	i = 0
	var err error
	if err != nil {
		panic(err)
	}

	db, err = bolt.Open("pa.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hasher = sha512.New()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("file2"))
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
