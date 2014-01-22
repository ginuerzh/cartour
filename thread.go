// thread
package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"time"
)

var (
	randMaxInt            = 1 << 31
	random     *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type Thread struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Title      string
	From       string
	Tid        string
	Url        string
	Author     string
	AuthorPage string    `bson:"author_page"`
	PubTime    time.Time `bson:"pub_time"`
	Content    []string
	Image      string
	Random     int
	Publish    bool
}

func (this *Thread) Save() error {
	insert := func(c *mgo.Collection) error {
		this.Id = bson.NewObjectId()
		this.Random = random.Intn(randMaxInt)
		return c.Insert(this)
	}

	return withCollection(threadsColl, insert)
}

func (this *Thread) Exists() (bool, error) {
	count := 0
	err := search(threadsColl, bson.M{"tid": this.Tid}, nil, 0, 0, nil, &count, nil)

	return count > 0, err
}

func (this *Thread) findOne(query interface{}) (bool, error) {
	var threads []Thread

	if err := search(threadsColl, query, nil, 0, 1, nil, nil, &threads); err != nil {
		return false, err
	}
	if len(threads) > 0 {
		*this = threads[0]
	}
	return len(threads) > 0, nil
}

func (this *Thread) FindByTid(tid string) (bool, error) {
	return this.findOne(bson.M{"tid": tid})
}

func GetThreadList(source string, skip, limit int) (total int, threads []Thread, err error) {
	var query bson.M

	if source != "" {
		query = bson.M{"from": source}
	}

	err = search(threadsColl, query, nil, skip, limit, []string{"-pub_time"}, &total, &threads)

	return
}

func (this *Thread) UpdateContent() error {
	change := bson.M{
		"$set": bson.M{
			"content": this.Content,
			"image":   this.Image,
		},
	}

	return updateId(threadsColl, this.Id, change)
}
