package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	ui "github.com/gosuri/uiprogress"
	"github.com/oklog/ulid"
	"gopkg.in/urfave/cli.v2"
)

const (
	errorBucket = "errors"
	statsBucket = "stats"
	keyStats    = "stats"
)

//List : Will display all the runids
func listRuns(c *cli.Context) error {
	db := openDb(dbname)
	defer closeDb(db)

	statsFromOnBucket := func(name []byte, b *bolt.Bucket) error {
		s := getBucketFromBucket(b, statsBucket)
		stats := &Statistics{}
		if err := json.Unmarshal(s.Get([]byte(keyStats)), stats); err != nil {
			log.Printf("Unable to unmarshall statistic")
			return err
		}
		fmt.Printf("Stats for %s \n \t %+v \n -----------------------------------\n", stats.Runid, stats)
		return nil
	}

	statsFromAllBucket := func(tx *bolt.Tx) error {
		return tx.ForEach(statsFromOnBucket)
	}
	return db.View(statsFromAllBucket)
}

//Display : Will display all the hash and associated file of the run id
func display(c *cli.Context) error {
	db := openDb(dbname)
	defer closeDb(db)

	displayBucker := func(tx *bolt.Tx) error {
		b := getBucketFromTransaction(tx, getBucketName(c))
		err := b.ForEach(func(k, v []byte) error {
			if v != nil {
				fmt.Printf("file=%s, hash=%s\n", k, hex.EncodeToString(v))
			}
			return nil
		})
		return err
	}
	return db.View(displayBucker)
}

//Error : Will display all the errors and associated file of the run id
func derror(c *cli.Context) error {
	db := openDb(dbname)
	defer closeDb(db)

	displayErrorFromOneBucket := func(tx *bolt.Tx) error {
		b := getBucketFromTransaction(tx, getBucketName(c))
		ber := getBucketFromBucket(b, errorBucket)
		err := ber.ForEach(func(k, v []byte) error {
			if v != nil {
				fmt.Printf("file=%s, hash=%s\n", k, v)
			}
			return nil
		})
		return err
	}
	return db.Update(displayErrorFromOneBucket)
}

//Compare : Will display all the runids
func compareRuns(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return fmt.Errorf("Please provice two runids")
	}
	return compare(c.Args().Get(0), c.Args().Get(1), dbname)
}

//Call by compareRuns
func compare(src string, dest string, dbname string) error {
	log.Printf("Comparing src %s with dest %s in db : %s", src, dest, dbname)
	db := openDb(dbname)
	defer closeDb(db)

	compareTwoBucket := func(tx *bolt.Tx) error {
		log.Println("starting analysis")
		bsrc := getBucketFromTransaction(tx, src)
		bdest := getBucketFromTransaction(tx, dest)

		if bsrc.Stats().KeyN != bdest.Stats().KeyN {
			fmt.Println("different size")
		}

		// Iterating over src
		c := bsrc.Cursor()
		d := bdest.Cursor()
		i, w := d.First()

		for k, v := c.First(); k != nil; {
			// key-value as empty value.
			if v == nil {
				k, v = c.Next()
			}
			if w == nil {
				i, w = d.Next()
			}
			if v != nil {
				comparaison := bytes.Compare(k, i)

				switch {
				case comparaison == 0:
					analysehash(k, v, w)
					k, v = c.Next()
					i, w = d.Next()
				case comparaison < 0:
					log.Printf("%s is missing in destination run ", k)
					k, v = c.Next()
				case comparaison > 0:
					log.Printf("%s is missing in the destirnation src ", i)
					i, w = d.Next()
				}
			}
		}
		return nil
	}

	return db.View(compareTwoBucket)
}

// delete : Will delete all information related to a run
func delete(c *cli.Context) error {
	if len(c.Args().First()) <= 0 {
		return fmt.Errorf("Please provice a runId")
	}
	db := openDb(dbname)
	defer closeDb(db)

	deleteOneBucket := func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(c.Args().First())); err != nil {
			return fmt.Errorf("error deleting runs: %s", err)
		}
		return nil
	}
	return db.Update(deleteOneBucket)
}

func analysehash(key []byte, srcvalue []byte, destvalue []byte) bool {
	if bytes.Equal(srcvalue, destvalue) {
		return true
	}
	log.Printf("%s is not ok", key)
	return false
}

//start : Will compute a hash for every file below the location and store it into the db
func startHash(c *cli.Context) error {

	fmt.Println("computing...")
	if len(c.Args().First()) <= 0 {
		return fmt.Errorf("Please provice a path or '.'")
	}
	dir := c.Args().First()

	name, err := os.Hostname()
	if err != nil {
		return err
	}

	db := openDb(dbname)
	defer closeDb(db)

	bucketName, err := createRunBucket(db)
	if err != nil {
		return err
	}

	fmt.Println("computing the number of files to be processed ")
	count := 0
	countDirectory := 0
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("eoor")
			return err
		}
		if info.IsDir() {
			countDirectory++
			return nil
		}
		count++
		return nil
	})
	fmt.Println("count ", count)
	if err != nil {
		return err
	}

	ui.Start()
	bar := ui.AddBar(count)
	bar.AppendCompleted()
	bar.PrependElapsed()

	stats := &Statistics{
		HostName:  name,
		Location:  dir,
		Start:     time.Now(),
		Errors:    0,
		Files:     0,
		Directory: countDirectory,
		Runid:     bucketName,
	}

	if err := MD5All(dir, db, bucketName, stats, bar); err != nil {
		return err
	}

	stats.Stop = time.Now()
	stats.Duration = stats.Stop.Sub(stats.Start)

	log.Printf("\n statisctic %+v \n", stats)
	return saveStat(db, bucketName, *stats)
}

func createRunBucket(db *bolt.DB) (string, error) {
	t := time.Now()
	entropy := rand.New(rand.NewSource(t.UnixNano()))

	bucketName := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	createBucket := func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = b.CreateBucketIfNotExists([]byte(errorBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = b.CreateBucket([]byte(statsBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	}
	return bucketName, db.Update(createBucket)
}

func saveStat(db *bolt.DB, runID string, stats Statistics) error {
	updateStat := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(runID))
		if b == nil {
			return fmt.Errorf("get bucket")
		}
		s := b.Bucket([]byte(statsBucket))
		if s == nil {
			return fmt.Errorf("get bucket")
		}
		bs, err := json.Marshal(stats)
		if err != nil {
			return fmt.Errorf("marshall %s", err)
		}
		if err := s.Put([]byte(keyStats), bs); err != nil {
			return err
		}
		return nil
	}
	return db.Update(updateStat)
}

func openDb(dbname string) *bolt.DB {
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Open Db %s \n -----------------------------------\n", dbname)
	return db
}

func closeDb(db *bolt.DB) {
	if err := db.Close(); err != nil {
		fmt.Println(err)
	}
}

func getBucketFromBucket(fb *bolt.Bucket, bucketName string) *bolt.Bucket {
	b := fb.Bucket([]byte(bucketName))
	if b == nil {
		log.Fatalf("Could not retrieve Bucket %s from Bucket", bucketName)
	}
	return b
}

func getBucketFromTransaction(tx *bolt.Tx, bucketName string) *bolt.Bucket {
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		log.Fatalf("Could not retrieve Bucket %s from Transaction", bucketName)
	}
	return b
}

func getBucketName(c *cli.Context) string {
	bucketName := c.Args().First()
	if bucketName == "" {
		log.Fatalf("Don't forget to pass the run ID in argument, you can get it by the command 'list'")
	}
	fmt.Printf("Run ID : %v\n", bucketName)
	return bucketName
}
