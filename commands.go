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

	ui "github.com/gosuri/uiprogress"
	"github.com/oklog/ulid"

	"gopkg.in/urfave/cli.v2"

	"github.com/boltdb/bolt"
)

func listRuns(c *cli.Context) error {

	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {

		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			//fmt.Println(string(name))
			s := b.Bucket([]byte("stats"))
			if s == nil {

				return fmt.Errorf("could not retrieve stats")
			}
			stats = &Statistics{}
			if err := json.Unmarshal(s.Get([]byte("stats")), stats); err != nil {
				log.Printf("unabale to unmarshall statistic")
			}
			fmt.Printf("stats for %s \n \t %+v \n", stats.Runid, stats)
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
	db, err := bolt.Open(dbname, 0600, nil)
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

func derror(c *cli.Context) error {
	bucketName := c.Args().First()
	db, err := bolt.Open(dbname, 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		if b == nil {
			return fmt.Errorf("unable to get run info: %s", err)
		}
		ber := b.Bucket([]byte("errors"))
		err = ber.ForEach(func(k, v []byte) error {
			if v != nil {
				fmt.Printf("file=%s, hash=%s\n", k, v)
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

func compareRuns(c *cli.Context) error {
	src := c.Args().Get(0)
	dest := c.Args().Get(1)

	compare(src, dest, dbname)
	return nil
}

func delete(c *cli.Context) error {
	db, err := bolt.Open(dbname, 0600, nil)
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
	name, err := os.Hostname()
	if err != nil {
		return err
	}

	stats = &Statistics{
		HostName:  name,
		Location:  dir,
		Start:     time.Now(),
		Errors:    0,
		Files:     0,
		Directory: 0}

	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	t := time.Now()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	fmt.Println()
	bucketName = ulid.MustNew(ulid.Timestamp(t), entropy).String()
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

	if info.IsDir() {
		stats.Directory++
		return nil
	}

	count++
	return nil
}

func compare(src string, dest string, dbname string) error {
	//ok := 0
	//nok := 0$
	log.Printf("Comparing src %s with dest %s in db : %s", src, dest, dbname)
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		log.Println(err)
		return (err)
	}
	err = db.View(func(tx *bolt.Tx) error {

		log.Println("starting analysis")
		bsrc := tx.Bucket([]byte(src))

		if bsrc == nil {
			log.Println("src is not a runid")
			return fmt.Errorf("unable to get info from src")
		}

		bdest := tx.Bucket([]byte(dest))
		if bdest == nil {
			log.Println("dest is not a runid")
			return fmt.Errorf("unable to get info from dest")
		}

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
				//	w := bdest.Get(k)
				comparaison := bytes.Compare(k, i)

				if comparaison == 0 {
					analysehash(k, v, w)
					k, v = c.Next()
					i, w = d.Next()
				}

				if comparaison < 0 {
					log.Printf("%s is missing in destination run ", k)
					k, v = c.Next()
				}

				if comparaison > 0 {

					log.Printf("%s is missing in the destirnation src ", i)
					i, w = d.Next()
				}
			}

		}
		return nil
	})
	return nil
}

func analysehash(key []byte, srcvalue []byte, destvalue []byte) bool {
	if bytes.Equal(srcvalue, destvalue) {
		return true
	}
	log.Printf("%s is not ok", key)
	return false

}