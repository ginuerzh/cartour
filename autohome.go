// autohome
package main

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AutoHome struct {
	CarTour
}

func NewAutoHome() *AutoHome {
	ah := &AutoHome{}
	ah.Name = "autohome"
	ah.Charset = "gbk"
	ah.Domain = "http://club.autohome.com.cn"
	ah.userPhotos = regexp.MustCompile(`club\d\.autoimg\.cn/album/userphotos/(.*)`)
	ah.timeLayout = "2006-1-2 15:04:05"
	return ah
}

func (this *AutoHome) Fetch(maxPages int, maxThreads int) (total int) {
	if maxPages <= 0 {
		maxPages = DefaultPages
	}
	if maxThreads <= 0 {
		maxThreads = DefaultThreads
	}
	tids := make([]string, 0, maxThreads)
	//threads := make([]*Thread, 0, maxThreads)

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
		t := &Thread{}
		t.Tid = tid
		if exists, _ := t.Exists(); exists {
			log.Println("Ignore exists thread", tid)
			continue
		}
		if t = this.FetchThread(tid); t != nil {
			if err := t.Save(); err != nil {
				log.Println("save thread", tid, "failed:", err)
			} else {
				total++
				log.Println("save thread", tid, "ok", t.Id.Hex())
			}
		} else {
			log.Println("get thread", tid, "failed:")
		}
	}

	return
}

func (this *AutoHome) FetchThread(tid string) *Thread {
	t := &Thread{}
	t.Publish = false
	t.From = this.Name
	t.Tid = tid
	t.Url = this.ThreadPageUrl(tid, 1)

	log.Println("fetch autohome", t.Url)

	doc, err := GetQueryDoc(t.Url, this.Charset)
	if err != nil {
		log.Println(err)
		return nil
	}

	topic := doc.Find("#maxwrap-maintopic")
	owner := topic.Find("ul.maxw li:nth-child(1) a")
	t.Author = owner.Text()
	t.AuthorPage, _ = owner.Attr("href")

	pubtime, err := time.Parse(this.timeLayout, topic.Find("span[xname='date']").Text())
	if err != nil {
		log.Println(err)
	} else {
		d, _ := time.ParseDuration("-8h")
		t.PubTime = pubtime.Add(d)
	}

	t.Title = topic.Find(".maxtitle").Text()

	buffer := NewContentBuffer()
	this.parseContent(topic.Find(".conttxt > .w740"), buffer)
	this.replyPerPage(doc, buffer)

	pages, _ := doc.Find("div#x-pages1").Attr("maxindex")
	//start from page 2
	pageNum, _ := strconv.ParseInt(pages, 10, 32)
	for i := 2; i <= int(pageNum); i++ {
		//log.Println("page", i)
		doc, err := GetQueryDoc(this.ThreadPageUrl(tid, i), this.Charset)
		if err != nil {
			log.Println(err)
			break
		}

		this.replyPerPage(doc, buffer)
	}

	if !buffer.IsValid() {
		log.Println("invalid thread", t.Url)
		return nil
	}

	t.Content = buffer.Content()
	//for _, line := range t.Content {
	//	log.Println(line)
	//}

	return t
}

func (this *AutoHome) replyPerPage(doc *goquery.Document, buffer *ContentBuffer) {
	doc.Find("#maxwrap-reply > div.contstxt").Each(func(i int, child *goquery.Selection) {
		w740 := child.Find(".w740")
		if w740.Find(".relyhf").Size() > 0 {
			//log.Println("just reply, ignore!")
			return
		}
		this.parseContent(w740, buffer)
	})
}

/*
func (this *AutoHome) replyPerPage(doc *goquery.Document) []string {
	var reply []string

	doc.Find("#maxwrap-reply > div.contstxt").Each(func(i int, child *goquery.Selection) {
		w740 := child.Find(".w740")
		if w740.Find(".relyhf").Size() > 0 {
			//log.Println("just reply, ignore!")
			return
		}

		if content := this.parseContent(w740); len(content) > 0 {
			reply = append(reply, content...)
		}
	})

	return reply
}
*/

func (this *AutoHome) parseContent(selector *goquery.Selection, buffer *ContentBuffer) {
	contents := selector.Contents()

	if selector.Is("p") || selector.Is("br") {
		buffer.Newline()
	}

	if contents.Length() == 0 {
		if !selector.Is("img") {
			buffer.Append(selector.Text())
		} else {
			src1, _ := selector.Attr("src")
			src2, _ := selector.Attr("src9")
			exist := true
			if this.userPhotos.MatchString(src1) {
				buffer.Newline()
				buffer.Append("[img]" + src1 + "[img]")
			} else if this.userPhotos.MatchString(src2) {
				buffer.Newline()
				buffer.Append("[img]" + src2 + "[img]")
			} else {
				exist = false
			}

			if exist {
				buffer.Newline()
				buffer.Valid(true)
			}
		}
		return
	}

	contents.Each(func(i int, child *goquery.Selection) {
		this.parseContent(child, buffer)
	})

	if selector.Is("p") {
		buffer.Newline()
	}
}

/*
func (this *AutoHome) parseContent(base *goquery.Selection) []string {
	var content []string
	exist := false

	base.Children().Each(func(i int, child *goquery.Selection) {
		child.Find("img").Each(func(i int, img *goquery.Selection) {
			src, _ := img.Attr("src")
			src9, _ := img.Attr("src9")
			if this.userPhotos.MatchString(src) {
				content = append(content, "[img]"+src+"[img]")
				exist = true
			} else if this.userPhotos.MatchString(src9) {
				content = append(content, "[img]"+src9+"[img]")
				exist = true
			}
		})
		text := strings.TrimSpace(child.Text())
		if len(text) > 0 {
			content = append(content, text)
		}
	})

	if !exist {
		return nil
	}
	return content
}
*/

func (this *AutoHome) GetTids(pageUrl string, max int) []string {
	log.Println("autohome thread list", pageUrl)

	tids := make([]string, 0, 100)

	doc, err := GetQueryDoc(pageUrl, this.Charset)
	if err != nil {
		log.Println(err)
		return nil
	}

	doc.Find("dl.list_dl").EachWithBreak(func(i int, dl *goquery.Selection) bool {
		link := dl.Find("dt > a")
		if link.Size() == 0 {
			//log.Println("not exists!")
			return true
		}
		titleLink := link.First()
		uri, exist := titleLink.Attr("href")
		if !exist {
			return true
		}
		tids = append(tids, this.parseThreadId(uri))

		if len(tids) >= max {
			return false
		}
		return true
	})
	return tids
}

func (this *AutoHome) parseThreadId(url string) string {
	list := strings.Split(url, "-")
	return list[2] + "-" + list[3]
}

func (this *AutoHome) ForumPageUrl(pageIndex int) string {
	return this.Domain + "/bbs/forum-o-200042-" +
		strconv.FormatInt(int64(pageIndex), 10) +
		".html?orderby=dateline&type=refine#pvareaid=101061"
}

func (this *AutoHome) ThreadPageUrl(tid string, pageIndex int) string {
	return this.Domain + "/bbs/threadowner-o-" + tid + "-" +
		strconv.FormatInt(int64(pageIndex), 10) + ".html#pvareaid=101435"
}
