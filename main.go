package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"gopkg.in/urfave/cli.v2"
)

type Statistics struct {
	Runid     string
	Files     int64
	Directory int64
	Errors    int64
	Start     time.Time
	Stop      time.Time
	Duration  time.Duration
}

func (s *Statistics) String() string {

	start := s.Start.Format("2006-01-01 15:11:12001")
	stop := s.Stop.Format("2006-01-01 15:11:12001")
	return fmt.Sprintf("[start : %s \n ,stop:%s \n, directory: %d, file %d, errors %d,]", start, stop, s.Directory, s.Files, s.Errors)
}

var i int64
var db *bolt.DB
var dbname string
var hasher hash.Hash
var bucketName string
var stats *Statistics

func main() {
	app := &cli.App{}
	app.Version = "0.1"
	app.EnableShellCompletion = true
	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "db", Value: "pa.db", Usage: "the database location (must have writting access)"},
	}
	app.Commands = []*cli.Command{
		{
			Name:        "start",
			Aliases:     []string{"s"},
			Usage:       "start location",
			Description: "will compute a hash for every file below the location and store it into the db",
			Action:      startHash,
		},
		{
			Name:        "display",
			Aliases:     []string{"h"},
			Usage:       "display  runid",
			Description: "will display all the hash and associated file of the run id",
			Action:      display,
		},

		{
			Name:        "delete",
			Aliases:     []string{"h"},
			Usage:       "delete  runid",
			Description: "will delete all information related to a run",
			Action:      delete,
		},

		{
			Name:        "list",
			Aliases:     []string{"a"},
			Usage:       "list",
			Description: "will display all the runids",
			Action:      listRuns,
		},
	}

	app.Run(os.Args)
}

func startHash(c *cli.Context) error {
	dir := c.Args().First()

	stats = &Statistics{Start: time.Now(),
		Errors:    0,
		Files:     0,
		Directory: 0}

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
	stats.Runid = bucketName
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
	stats.Stop = time.Now()
	stats.Duration = stats.Stop.Sub(stats.Start)
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		s, err := b.CreateBucket([]byte("stats"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		fmt.Println(stats)
		bs, err := json.Marshal(stats)
		if err != nil {
			return fmt.Errorf("marshall %s", err)
		}

		s.Put([]byte("stats"), bs)
		return nil
	})

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

		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			//fmt.Println(string(name))
			s := b.Bucket([]byte("stats"))
			if s == nil {
				return nil
			}
			stats = &Statistics{}
			json.Unmarshal(s.Get([]byte("stats")), stats)
			fmt.Printf("stats for %s =  %+v", stats.Runid, stats)
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
		stats.Directory++
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
		return nil
	}

	hasher.Write(data)
	sha := hasher.Sum(nil)

	stats.Files++
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
