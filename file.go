// file
package main

import (
	"github.com/ginuerzh/weedo"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type File struct {
	Id          bson.ObjectId `bson:"_id,omitempty"`
	Fid         string
	Name        string `bson:"filename"`
	Length      int64  `bson:"length"`
	Md5         string
	Owner       string
	Count       int
	ContentType string    `bson:"contentType"`
	UploadDate  time.Time `bson:"uploadDate"`
}

func (this *File) findOne(query interface{}) (bool, error) {
	var files []File

	err := search(fileColl, query, nil, 0, 1, nil, nil, &files)
	if err != nil {
		return false, err
	}
	if len(files) > 0 {
		*this = files[0]
	}

	return len(files) > 0, nil
}

func (this *File) FindByFid(fid string) (bool, error) {
	return this.findOne(bson.M{"fid": fid})
}

func (this *File) Save() error {
	insert := func(c *mgo.Collection) error {
		this.Id = bson.NewObjectId()
		return c.Insert(this)
	}

	return withCollection(fileColl, insert)
}

func (this *File) Delete() error {
	remove := func(c *mgo.Collection) error {
		err := c.Remove(bson.M{"fid": this.Fid})
		if err == nil {
			weedo.Delete(this.Fid, this.Count) //TODO: fail process
		}
		return err
	}

	if err := withCollection(fileColl, remove); err != nil {
		if err != mgo.ErrNotFound {
			return err
		}
	}
	return nil
}

func (this *File) OwnedBy(userid string) (bool, error) {
	return this.findOne(bson.M{"fid": this.Fid, "owner": userid})
}
