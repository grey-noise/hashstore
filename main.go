package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/urfave/cli.v2"
)

// Statistics is the struc to save statistics data
type Statistics struct {
	HostName  string
	Location  string
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
	return fmt.Sprintf("[hostname: %s location: %s start : [%s] stop:%s, directory: [%d], file [%d], errors [%d]]", s.HostName, s.Location, start, stop, s.Directory, s.Files, s.Errors)
}

//var bar *pb.Bar
var dbname string
var digesterNr int

func main() {
	app := &cli.App{}
	app.Version = "0.1"
	app.EnableShellCompletion = true
	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "db", Value: "pa.db", Usage: "the database location (must have writing access)", Destination: &dbname},
		&cli.IntFlag{Name: "pr", Value: 3, Usage: "the number of workers", Destination: &digesterNr},
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
			Name:        "error",
			Aliases:     []string{"r"},
			Usage:       "error runid",
			Description: "will display all the errors and associated file of the run id",
			Action:      derror,
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
		{
			Name:        "compare",
			Aliases:     []string{"a"},
			Usage:       "compare runid1 runid2",
			Description: "will display all the runids",
			Action:      compareRuns,
		},
	}

	if err := (app.Run(os.Args)); err != nil {
		log.Fatal(err)
	}
}
