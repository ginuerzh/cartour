package main

import (
	"github.com/codegangsta/cli"
	"log"
	"os"
	"strings"
)

const (
	autoHome       = "autohome"
	autoHomeTravel = "travel"
	bitAuto        = "bitauto"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	app := cli.NewApp()
	app.Name = "cartour"
	app.Usage = "fetch, update, query, delete threads and more"
	app.Version = "0.1"

	app.Commands = []cli.Command{
		{
			Name:        "fetch",
			ShortName:   "f",
			Usage:       "fetch threads from bbs",
			Description: "Note: this command will overwrite data, if you want to update, please use update command",
			Flags: []cli.Flag{
				cli.StringFlag{"sources, s", "", "the source we fetch from, may autohome or bitauto, separated by comma"},
				//cli.StringSliceFlag{"sources, s", &cli.StringSlice{}, "the source we fetch from, may autohome or bitauto"},
				cli.StringFlag{"db, d", "cartour", "mongodb database to use"},
				cli.StringFlag{"collection, c", "threads", "mongodb collection to use"},
				cli.IntFlag{"threads, t", 0, "threads to fetch"},
				cli.IntFlag{"pages, p", 1, "pages to fetch"},
			},
			Action: func(c *cli.Context) {
				Fetch(c.String("sources"), c.Int("pages"), c.Int("threads"))
			},
		},

		{
			Name:      "image",
			ShortName: "i",
			Usage:     "fetch images in threads",
			Flags: []cli.Flag{
				cli.StringFlag{"tid, t", "", "thread id"},
				cli.IntFlag{"threads, n", 0, "threads to update"},
				cli.StringFlag{"source, s", "", "autohome or bitauto"},
			},
			Action: func(c *cli.Context) {
				UpdateImages(c.String("source"), c.String("tid"), c.Int("threads"))
			},
		},

		{
			Name:  "remove",
			Usage: "remove threads and all related resource(e.g. images).",
			Flags: []cli.Flag{
				cli.StringFlag{"from", "", "date from, e.g. 2014-01-01"},
				cli.StringFlag{"to", "", "date to, same as from"},
			},
			Action: func(c *cli.Context) {
				Remove(c.String("from"), c.String("to"))
			},
		},
	}

	app.Run(os.Args)
}

func Fetch(sources string, maxPages, maxThreads int) {
	for _, name := range strings.Split(sources, ",") {
		if len(name) == 0 {
			continue
		}
		count := 0
		//threads := []*Thread{}
		if name == autoHome {
			autohome := NewAutoHome()
			count = autohome.Fetch(maxPages, maxThreads)
		} else if name == bitAuto {
			bitauto := NewBitAuto()
			count = bitauto.Fetch(maxPages, maxThreads)
		} else if name == autoHomeTravel {
			autohome := NewAutoHome()
			count = autohome.FetchTravel(maxPages, maxThreads)
		}
		log.Println("fetch threads done, total", count)
	}
}
