// bitauto
package main

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BitAuto struct {
	CarTour
}

func NewBitAuto() *BitAuto {
	ba := &BitAuto{}
	ba.Name = "bitauto"
	ba.Charset = "utf-8"
	ba.Domain = "http://baa.bitauto.com"
	ba.userPhotos = regexp.MustCompile(`http://img\d\.baa\.bitautotech\.com/img/(.*)\.jpg`)
	ba.timeLayout = "2006-01-02 15:04"
	return ba
}

func (this *BitAuto) Fetch(maxPages int, maxThreads int) []*Thread {
	if maxPages <= 0 {
		maxPages = DefaultPages
	}
	if maxThreads <= 0 {
		maxThreads = DefaultThreads
	}
	tids := make([]string, 0, maxThreads)
	threads := make([]*Thread, 0, maxThreads)

	for i := 0; i < maxPages; i++ {
		list := this.GetTids(this.ForumPageUrl(i+1), maxThreads)
		if len(list) > 0 {
			tids = append(tids, list...)
		}

		maxThreads -= len(list)
		if maxThreads == 0 {
			break
		}
	}

	for _, tid := range tids {
		if t := this.FetchThread(tid); t != nil {
			threads = append(threads, t)
		}
	}

	this.Threads = threads

	return threads
}

func (this *BitAuto) GetTids(pageUrl string, max int) []string {
	tids := make([]string, 0, 100)

	doc, err := GetQueryDoc(pageUrl, this.Charset)
	if err != nil {
		log.Println(err)
		return nil
	}

	doc.Find(".postslist_xh").EachWithBreak(func(i int, entry *goquery.Selection) bool {
		link := entry.Find("ul > li.bt > a")
		if link.Size() == 0 {
			//log.Println("not exists!")
			return true
		}
		url, exist := link.Attr("href")
		if !exist {
			return true
		}

		tids = append(tids, this.parseThreadId(url))
		if len(tids) >= max {
			return false
		}
		return true
	})

	return tids
}

func (this *BitAuto) FetchThread(tid string) *Thread {
	t := &Thread{}
	t.Content = make([]string, 0, 100)

	t.From = this.Name
	t.Tid = tid
	url := this.ThreadPageUrl(tid, 1)
	t.Url = url

	i := 1
	for {
		more := this.perPageContent(url, t)
		if !more {
			break
		}
		i++
		url = this.ThreadPageUrl(tid, i)
	}

	return t
}

func (this *BitAuto) perPageContent(pageUrl string, t *Thread) bool {
	doc, err := GetQueryDoc(pageUrl, "utf-8")
	if err != nil {
		log.Println(err)
		return false
	}

	t.Title = doc.Find(".bbsnamebox > h2").Text()

	loadInfo := func(t *Thread, post *goquery.Selection) {
		link := post.Find("a.mingzi")
		url, exist := link.Attr("href")
		if exist {
			t.Author = link.Text()
			t.AuthorPage = url
		}

		timeInfo := post.Find(".postright > .fabiaoyubox > .time").Text()
		timeInfo = strings.TrimSpace(timeInfo)
		start := strings.Index(timeInfo, "2")
		//log.Println(timeInfo[start:])

		pubtime, err := time.Parse(this.timeLayout, timeInfo[start:])
		if err != nil {
			log.Fatal(err)
		}
		t.PubTime = pubtime
	}

	doc.Find(".postcontbox > .postcont_list").Each(func(i int, post *goquery.Selection) {
		if len(t.Author) == 0 {
			loadInfo(t, post)
		}

		if post.Find(".yinyongbox").Size() > 0 {
			//log.Println("just reply, ignore!")
			return
		}

		base := post.Find(".post_width")
		if base.Size() > 1 {
			base = base.Last()
		}
		content := this.parseContent(base)
		if len(content) > 0 {
			t.Content = append(t.Content, content...)
		}
	})

	return doc.Find(".next_on").Size() > 0
}

func (this *BitAuto) parseContent(base *goquery.Selection) []string {
	var content []string
	exist := false

	base.Children().Each(func(i int, child *goquery.Selection) {

		text := strings.TrimSpace(child.Text())
		if len(text) > 0 {
			content = append(content, text)
		}

		child.Find("img").Each(func(i int, img *goquery.Selection) {
			src, _ := img.Attr("_src")
			sourcesrc, _ := img.Attr("_sourcesrc")
			if this.userPhotos.MatchString(src) {
				content = append(content, src)
				exist = true
			} else if this.userPhotos.MatchString(sourcesrc) {
				content = append(content, sourcesrc)
				exist = true
			}
		})
	})

	if !exist {
		return nil
	}
	return content
}

func (this *BitAuto) parseThreadId(url string) string {
	s := strings.Split(url, "/")
	s = strings.Split(s[len(s)-1], ".")
	return s[0]
}

func (this *BitAuto) ForumPageUrl(pageIndex int) string {
	return this.Domain + "/drive/index-0-all-" +
		strconv.FormatInt(int64(pageIndex), 10) + "-1.html"
}

func (this *BitAuto) ThreadPageUrl(tid string, pageIndex int) string {
	if pageIndex == 1 {
		return this.Domain + "/drive/" + tid + "-building.html"
	}
	return this.Domain + "/drive/" + tid + "-building-" + strconv.FormatInt(int64(pageIndex), 10) + ".html"
}
