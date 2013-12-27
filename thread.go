// thread
package main

import (
	"time"
)

type Thread struct {
	Title      string
	From       string
	Tid        string
	Url        string
	Author     string
	AuthorPage string    `bson:"author_page"`
	PubTime    time.Time `bson:"pub_time"`
	Content    []string
}
