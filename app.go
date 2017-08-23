package main

import (
	"fmt"
	"hash"
	"os"
	"time"

	"github.com/boltdb/bolt"
	pb "github.com/gosuri/uiprogress"
	"gopkg.in/urfave/cli.v2"
)

type Statistics struct {
	Runid     string
	Files     int
	Directory int
	Errors    int
	Start     time.Time
	Stop      time.Time
	Duration  time.Duration
}

func (s *Statistics) String() string {

	start := s.Start.Format(time.RFC3339Nano)
	stop := s.Stop.Format(time.RFC3339Nano)
	return fmt.Sprintf("[start : %s \n ,stop:%s \n, directory: %d, file %d, errors %d,]", start, stop, s.Directory, s.Files, s.Errors)
}

var i int
var count int

var db *bolt.DB
var bar *pb.Bar
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

	fmt.Println(app.Run(os.Args))
}
