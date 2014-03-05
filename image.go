// image
package main

import (
	"bytes"
	"errors"
	"github.com/ginuerzh/weedo"
	"github.com/nfnt/resize"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	//"sync"
	"time"
)

func UpdateImages(source, tid string, maxThreads int) {
	log.Println("start update images")

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

	skip := 0
	limit := 10
	for {
		if skip+limit > maxThreads {
			limit = maxThreads - skip
		}
		if limit == 0 {
			break
		}
		_, threads, err := GetThreadList(source, skip, limit)
		if err != nil {
			log.Println(err)
			return
		}
		//log.Println("total threads:", total, "will update", maxThreads)
		if len(threads) == 0 {
			break
		}

		/*
			var wg sync.WaitGroup

			for i, _ := range threads {
				wg.Add(1)
				go func(thread *Thread) {
					defer wg.Done()

					FetchThreadImages(thread)
				}(&threads[i])
			}

			wg.Wait()
		*/
		for i, _ := range threads {
			FetchThreadImages(&threads[i])
		}
		skip += limit
	}

	log.Println("update images done")
}

func FetchThreadImages(thread *Thread) {
	contents := thread.Content
	count := 0
	for i, _ := range contents {
		if strings.HasPrefix(contents[i], "[img]") &&
			strings.HasSuffix(contents[i], "[img]") {

			url := strings.TrimSuffix(strings.TrimPrefix(contents[i], "[img]"), "[img]")
			fid, size, size1, size2, err := fetchImage(thread.From, url, thread.Url)
			if err != nil {
				if len(fid) > 0 {
					weedo.Delete(fid, 3)
				}
				log.Println(url)
				log.Println("fetch image failed", err)
				continue
			}
			contents[i] = "[fid]" + fid + "," + size + "," + size1 + "," + size2 + "[fid]"
			//log.Println("fetch image ok", contents[i])
			if len(thread.Image) == 0 {
				thread.Image = fid
				log.Println("add first image", thread.Image)
			}
			count++
		}
	}
	if count > 0 {
		//log.Println("total images", count)
		if err := thread.UpdateContent(); err != nil {
			log.Println("save thread", thread.Id.Hex(), "images failed:", err)
		} else {
			log.Println("save thread", thread.Id.Hex(), "images ok")
		}
	}
	thread.Pub(true)
}

func fetchImage(from, url, referer string) (fid string, size string, size1 string, size2 string, err error) {
	//log.Println("fetch image", url)
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return
	}
	if from == autoHome {
		request.Header.Set("Referer", referer)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("http error: " + resp.Status)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	//fid, size, err := weedo.AssignUpload(s[len(s)-1], "image/jpeg", bytes.NewBuffer(data))
	fid, err = weedo.AssignN(3)
	if err != nil {
		return
	}
	//log.Println(fid)

	image, err := jpeg.Decode(bytes.NewBuffer(data))
	if err != nil {
		return
	}

	imgResized := &bytes.Buffer{}
	resized := resize.Resize(640, 0, image, resize.MitchellNetravali)
	if err = jpeg.Encode(imgResized, resized, nil); err != nil {
		return
	}
	size = strconv.Itoa(resized.Bounds().Dx()) + "X" + strconv.Itoa(resized.Bounds().Dy())

	filename := strconv.Itoa(time.Now().Nanosecond()) + ".jpg"
	if s := strings.Split(url, "/"); len(s) > 0 {
		filename = s[len(s)-1]
	}
	length, err := weedo.VolumeUpload(fid, 0, filename, "image/jpeg", imgResized)
	if err != nil {
		return
	}

	imgResized.Reset()
	resized = resize.Resize(266, 0, image, resize.MitchellNetravali)
	if err = jpeg.Encode(imgResized, resized, nil); err == nil {
		weedo.VolumeUpload(fid, 1, filename, "image/jpeg", imgResized)
	}
	size1 = strconv.Itoa(resized.Bounds().Dx()) + "X" + strconv.Itoa(resized.Bounds().Dy())

	imgResized.Reset()
	resized = resize.Resize(80, 0, image, resize.MitchellNetravali)
	if err = jpeg.Encode(imgResized, resized, nil); err == nil {
		weedo.VolumeUpload(fid, 2, filename, "image/jpeg", imgResized)
	}
	size2 = strconv.Itoa(resized.Bounds().Dx()) + "X" + strconv.Itoa(resized.Bounds().Dy())

	file := File{}
	file.Fid = fid
	file.ContentType = "image/jpeg"
	file.Name = filename
	file.Owner = "admin"
	file.Length = length
	file.Count = 3
	file.UploadDate = time.Now()
	file.Md5 = FileMd5(bytes.NewBuffer(data))
	err = file.Save()

	return
}
