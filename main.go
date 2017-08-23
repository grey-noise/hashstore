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

	"gopkg.in/urfave/cli.v2"

	"github.com/boltdb/bolt"
)

type statistics struct {
	files     int64
	directory int64
	errors    int64
	start     time.Time
	stop      time.Time
	duration  time.Duration
}

func (s *statistics) String() string {

	start := s.start.Format("2006-01-01 15:11:12001")
	stop := s.stop.Format("2006-01-01 15:11:12001")
	return fmt.Sprintf("[start : %s \n ,stop:%s \n, directory: %d, file %d, errors %d,]", start, stop, s.directory, s.files, s.errors)
}

var i int64
var db *bolt.DB
var hasher hash.Hash
var bucketName string
var stats *statistics

func main() {
	app := &cli.App{}
	app.Version = "19.99.0"
	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "name", Value: "bob", Usage: "a name to say"},
	}
	app.Commands = []*cli.Command{
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
			Name:    "delete",
			Aliases: []string{"h"},
			Usage:   "display  a run",
			Action:  delete,
		},

		{
			Name:    "list",
			Aliases: []string{"a"},
			Usage:   "list all the runs ",
			Action:  listRuns,
		},
	}

	app.Run(os.Args)
}

func startHash(c *cli.Context) error {
	dir := c.Args().First()

	stats = &statistics{start: time.Now(),
		errors:    0,
		files:     0,
		directory: 0}

	var err error
	if err != nil {
		return (err)
	}

	db, err = bolt.Open("pa.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	name, err := os.Hostname()
	if err != nil {
		return (err)
	}

	t := time.Now().Local()

	bucketName = fmt.Sprintf("%s://%s://%s", name, dir, t.Format("2006-01-01"))

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
		return err
	}
	stats.stop = time.Now()
	stats.duration = stats.stop.Sub(stats.start)
	fmt.Printf("statisctic %+v ", stats)
	return nil
}

func display(c *cli.Context) error {
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
	return nil
}

func delete(c *cli.Context) error {
	db, err := bolt.Open("pa.db", 0600, nil)
	if err != nil {
		return (err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(c.Args().First()))
		if err != nil {
			return fmt.Errorf("error deleting runs: %s", err)
		}
		return nil
	})
	return nil
}

func listRuns(c *cli.Context) error {

	db, err := bolt.Open("pa.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {

		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			fmt.Println(string(name))
			return nil
		})
	})
	if err != nil {

		return err
	}
	return nil
}

//-----------------------------------------------------------------------------------/
// utilities                                                                         /
//-----------------------------------------------------------------------------------/
func hashcode(path string, info os.FileInfo, err error) error {
	hasher.Reset()
	if err != nil {
		log.Print(err)
		return nil
	}
	if info.Name() == "pa.db.lock" {
		return nil
	}
	if info.IsDir() {
		stats.directory++
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
		return nil
	}

	hasher.Write(data)
	sha := hasher.Sum(nil)

	stats.files++
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("gettinh bucket: %s", err)
		}
		b.Put([]byte(path), sha)
		return nil
	})
	return nil
}
