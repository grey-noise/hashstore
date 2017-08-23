package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	ui "github.com/gosuri/uiprogress"

	"gopkg.in/urfave/cli.v2"

	"github.com/boltdb/bolt"
)

func listRuns(c *cli.Context) error {

	db, err := bolt.Open("pa.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {

		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			fmt.Println(string(name))
			s := b.Bucket([]byte("stats"))
			if s == nil {

				return fmt.Errorf("could not retrieve stats")
			}
			stats = &Statistics{}
			if err := json.Unmarshal(s.Get([]byte("stats")), stats); err != nil {
				log.Printf("unabale to unmarshall statistic")
			}
			fmt.Printf("stats for %s =  %+v", stats.Runid, stats)
			return nil
		})
	})
	if err != nil {
		return err
	}
	return nil
}
func display(c *cli.Context) error {
	bucketName := c.Args().First()

	db, err := bolt.Open("pa.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		err = b.ForEach(func(k, v []byte) error {
			if v != nil {
				fmt.Printf("file=%s, hash=%s\n", k, hex.EncodeToString(v))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func delete(c *cli.Context) error {
	db, err := bolt.Open("pa.db", 0600, nil)
	if err != nil {
		return (err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(c.Args().First()))
		if err != nil {
			log.Println("error Deleting", c.Args().First())
			return fmt.Errorf("error deleting runs: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}
func startHash(c *cli.Context) error {
	fmt.Println("computing...")
	dir := c.Args().First()

	stats = &Statistics{Start: time.Now(),
		Errors:    0,
		Files:     0,
		Directory: 0}

	db, err := bolt.Open("pa.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	name, err := os.Hostname()
	if err != nil {
		return err
	}

	t := time.Now().Local()

	bucketName = fmt.Sprintf("%s://%s://%s", name, dir, t.Format("2006-01-01"))
	stats.Runid = bucketName
	fmt.Println("computing the number of files to be processed ")

	if err := filepath.Walk(dir, fcount); err != nil {
		return err
	}

	ui.Start()
	bar = ui.AddBar(count)
	bar.AppendCompleted()
	bar.PrependElapsed()

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = b.CreateBucketIfNotExists([]byte("errors"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})
	if err != nil {
		log.Println(err)
	}
	if err := MD5All(dir, db); err != nil {
		return err
	}
	stats.Stop = time.Now()
	stats.Duration = stats.Stop.Sub(stats.Start)
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		s, err := b.CreateBucket([]byte("stats"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		bs, err := json.Marshal(stats)
		if err != nil {
			return fmt.Errorf("marshall %s", err)
		}

		if err := s.Put([]byte("stats"), bs); err != nil {
			log.Println(err)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("statisctic %+v ", stats)
	return nil
}
func fcount(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Print(err)
		//	return err
	}
	if info.Name() == "pa.db.lock" {
		return nil
	}
	if info.IsDir() {
		stats.Directory++
		return nil
	}

	count++
	return nil
}
