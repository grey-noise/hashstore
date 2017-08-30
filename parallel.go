package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/boltdb/bolt"
	ui "github.com/gosuri/uiprogress"
)

// A result is the product of reading and summing a file using MD5.
type result struct {
	path string
	sum  []byte
	err  error
}

// walkFiles starts a goroutine to walk the directory tree at root and send the
// path of each regular file on the string channel.  It sends the result of the
// walk on the error channel.  If done is closed, walkFiles abandons its work.
func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error) {
	paths := make(chan string)
	errc := make(chan error, 1)
	go func() { // HL
		// Close the paths channel after Walk returns.
		defer close(paths) // HL
		// No select needed for this send, since errc is buffered.
		errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error { // HL
			if err != nil {
				log.Println("eoor")
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}

			select {
			case paths <- path: // HL
			case <-done: // HL
				return errors.New("walk canceled")
			}
			return nil
		})
	}()
	return paths, errc
}

// digester reads path names from paths and sends digests of the corresponding
// files on c until either paths or done is closed.
func digester(done <-chan struct{}, paths <-chan string, c chan<- result) {
	for path := range paths { // HLpaths
		res := result{path: path}
		f, err := os.Open(path)
		if err != nil {
			res.err = err
		} else {
			h := md5.New()
			if _, err := io.Copy(h, f); err != nil {
				res.err = err
			}
			res.sum = h.Sum(nil)
		}
		select {
		case c <- res:
		case <-done:
			return

		}

	}
}

// MD5All reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, MD5All returns an error.  In that case,
// MD5All does not wait for inflight read operations to complete.
func MD5All(root string, dbName string, bucketName string, stats *Statistics, bar *ui.Bar) error {

	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	// MD5All closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.
	done := make(chan struct{})
	defer close(done)

	paths, errc := walkFiles(done, root)

	// Start a fixed number of goroutines to read and digest files.
	c := make(chan result) // HLc
	var wg sync.WaitGroup

	wg.Add(digesterNr)
	log.Printf("launching %d digester", digesterNr)
	for i := 0; i < digesterNr; i++ {
		go func() {
			digester(done, paths, c) // HLc
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c) // HLc
	}()
	// End of pipeline. OMIT
	err = db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		e := b.Bucket([]byte("errors"))
		if b == nil {
			return fmt.Errorf("gettings bucket")
		}
		for r := range c { // HLrange
			bar.Incr()
			stats.Files++
			if r.err != nil {
				stats.Errors++
				if err = e.Put([]byte(r.path), []byte(r.err.Error())); err != nil {
					log.Println(err)
				}
			} else {
				if err = b.Put([]byte(r.path), r.sum[:]); err != nil {
					log.Println(err)
				}

			}

		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := <-errc; err != nil {
		log.Println(err)
		return err
	}
	return nil
}
