// remove
package main

import (
	"fmt"
	//"labix.org/v2/mgo"
	"github.com/ginuerzh/weedo"
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

const (
	Layout = "2006-01-02"
)

func Remove(from, to string) int {
	var query, m bson.M
	var thread Thread

	fromTime, err := time.Parse(Layout, from)
	if err == nil {
		m = bson.M{"$gt": fromTime}
	}
	toTime, err := time.Parse(Layout, to)
	if err == nil {
		if m == nil {
			m = bson.M{}
		}
		m["$lt"] = toTime
	}

	if m != nil {
		query = bson.M{"pub_time": m}
	}

	fmt.Println(query)
	total := 0
	it, _ := iter(threadsColl, query, []string{"pub_time"}, &total)
	fmt.Println("total:", total)

	defer it.Close()

	total = 0
	for it.Next(&thread) {
		delThread(&thread)
	}

	return 0
}

func delThread(thread *Thread) {
	for _, text := range thread.Content {

		if strings.HasPrefix(text, "[img]") && strings.HasSuffix(text, "[img]") {
			continue
		}

		if strings.HasPrefix(text, "[fid]") &&
			strings.HasSuffix(text, "[fid]") {
			fid := strings.TrimSuffix(strings.TrimPrefix(text, "[fid]"), "[fid]")
			a := strings.Split(fid, ",")
			if len(a) < 2 {
				continue
			}

			fid = strings.Join(a[:2], ",")
			if err := weedo.Delete(fid, 3); err != nil {
				fmt.Println("delete:", err)
			}
		}
	}

	if err := thread.Remove(); err != nil {
		fmt.Println("remove:", err)
	}
}
