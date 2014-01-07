// image
package main

import (
	"bytes"
	"errors"
	"github.com/ginuerzh/weedo"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func UpdateImages(source, tid string, maxThreads int) {
	thread := &Thread{}
	if tid != "" {
		if find, err := thread.FindByTid(tid); !find {
			log.Println(err)
			return
		}

		FetchThreadImages(thread)
		return
	}

	if maxThreads <= 0 {
		maxThreads = DefaultThreads
	}

	total, threads, err := GetThreadList(source, 0, maxThreads)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("total threads:", total, "will update", maxThreads)

	for i, _ := range threads {
		FetchThreadImages(&threads[i])
	}
}

func FetchThreadImages(thread *Thread) {
	contents := thread.Content
	for i, _ := range contents {
		if strings.HasPrefix(contents[i], "http") &&
			strings.HasSuffix(contents[i], ".jpg") {
			fid, err := fetchImage(thread.From, contents[i], thread.Url)
			if err != nil {
				log.Println(fid, err)
				continue
			}
			log.Println("fetch image ok", fid)
			contents[i] = "[img]" + fid + "[/img]"
		}
	}
	thread.UpdateContent()
}

func fetchImage(from, url, referer string) (string, error) {
	log.Println("fetch image", url)
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}
	if from == autoHome {
		request.Header.Set("Referer", referer)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	s := strings.Split(url, "/")
	fid, size, err := weedo.AssignUpload(s[len(s)-1], "image/jpeg", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	file := File{}
	file.Fid = fid
	file.ContentType = "image/jpeg"
	file.Name = s[len(s)-1]
	file.Owner = "admin"
	file.Size = size
	file.UploadDate = time.Now()
	file.Md5 = FileMd5(bytes.NewBuffer(data))
	err = file.Save()

	return fid, err
}
