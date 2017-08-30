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

var (
	dbname     string
	digesterNr int
	// Version overwritted by the VERSION file, necessary to use "make build" or "./scripts/travis/compile.sh"
	Version = "Not fixed"
)

//Overloading of String to display information
func (s *Statistics) String() string {
	return fmt.Sprintf("[hostname: %s location: %s start : [%s] stop: [%s], directory: [%d], file [%d], errors [%d]]", s.HostName, s.Location, s.Start.Format(time.RFC3339Nano), s.Stop.Format(time.RFC3339Nano), s.Directory, s.Files, s.Errors)
}

func main() {
	app := &cli.App{
		Version: Version,
		Usage:   "Hashstore application",
		EnableShellCompletion: true,
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "db", Value: "hashstore.db", Usage: "the database location (must have writing access)", Destination: &dbname},
		&cli.IntFlag{Name: "pr", Value: 3, Usage: "the number of workers", Destination: &digesterNr},
	}

	app.Commands = []*cli.Command{
		{
			Name:        "start",
			Aliases:     []string{"s"},
			Usage:       "start location",
			Description: "Will compute a hash for every file below the location and store it into the db",
			Action:      startHash,
		},
		{
			Name:        "display",
			Aliases:     []string{"h"},
			Usage:       "display runid",
			Description: "Will display all the hash and associated file of the run id",
			Action:      display,
		},
		{
			Name:        "error",
			Aliases:     []string{"r"},
			Usage:       "error runid",
			Description: "Will display all the errors and associated file of the run id",
			Action:      derror,
		},
		{
			Name:        "delete",
			Aliases:     []string{"h"},
			Usage:       "delete runid",
			Description: "Will delete all information related to a run",
			Action:      delete,
		},

		{
			Name:        "list",
			Aliases:     []string{"a"},
			Usage:       "list",
			Description: "Will display all the runids",
			Action:      listRuns,
		},
		{
			Name:        "compare",
			Aliases:     []string{"a"},
			Usage:       "compare runid1 runid2",
			Description: "Will display all the runids",
			Action:      compareRuns,
		},
	}

	if err := (app.Run(os.Args)); err != nil {
		log.Fatal(err)
	}
}
