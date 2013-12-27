// cartour
package main

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

const (
	DefaultPages   = 65535
	DefaultThreads = 65535
)

type Fetcher interface {
	Fetch(pages int, threads int) []*Thread
}

type CarTour struct {
	Name       string
	Domain     string
	Charset    string
	Threads    []*Thread
	userPhotos *regexp.Regexp
	timeLayout string
}

func GetQueryDoc(url string, charset string) (*goquery.Document, error) {
	log.Println("GET", url)

	if len(charset) == 0 {
		charset = "utf-8"
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	in, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(in)*2)
	r, w, err := iconv.Convert(in, out, charset, "utf-8//IGNORE")
	if err != nil {
		log.Println(r, w, err, string(out))
		return nil, err
	}

	return goquery.NewDocumentFromReader(bytes.NewBuffer(out))
}
