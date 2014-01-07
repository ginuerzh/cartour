// util
package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var (
	mgoSession   *mgo.Session
	databaseName = "cartour"
	threadsColl  = "threads"
	fileColl     = "files"
)

func getSession() *mgo.Session {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial("localhost")
		if err != nil {
			panic(err) // no, not really
		}
	}
	return mgoSession.Clone()
}

func withCollection(collection string, s func(*mgo.Collection) error) error {
	session := getSession()
	defer session.Close()
	c := session.DB(databaseName).C(collection)
	return s(c)
}

func search(collection string, query interface{}, selector interface{},
	skip, limit int, sortFields []string, total *int, result interface{}) error {

	q := func(c *mgo.Collection) error {
		qy := c.Find(query)
		var err error

		if selector != nil {
			qy = qy.Select(selector)
		}

		if total != nil {
			if *total, err = qy.Count(); err != nil {
				return err
			}
		}

		if limit > 0 {
			qy = qy.Limit(limit)
		}
		if skip > 0 {
			qy = qy.Skip(skip)
		}
		if len(sortFields) > 0 {
			qy = qy.Sort(sortFields...)
		}

		if result != nil {
			err = qy.All(result)
		}
		return err
	}

	return withCollection(collection, q)
}

func updateId(collection string, id bson.ObjectId, change interface{}) error {
	update := func(c *mgo.Collection) error {
		return c.UpdateId(id, change)
	}

	return withCollection(collection, update)
}

func FileMd5(file io.Reader) string {
	h := md5.New()
	io.Copy(h, file)
	return fmt.Sprintf("%x", h.Sum(nil))
}
