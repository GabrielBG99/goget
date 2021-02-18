package filemgr

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	BUF_SIZE              = 1024 * 8
	CONTENT_LENGTH_HEADER = "Content-Length"
	ACCEPT_RANGES_HEADER  = "Accept-Ranges"

	FILE_PART_NAME_TEMPLATE = "%s.part%v"
)

type DownloadClient struct {
	URL         string
	FilePath    string
	Size        uint64
	Parts       uint64
	AcceptRange bool
}

func (d DownloadClient) downloadPart(path string, begin, end int, ch chan<- error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		ch <- err
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		ch <- err
		return
	}

	actualSize := int(stat.Size())
	begin += int(actualSize)
	if begin >= end {
		ch <- nil
		return
	}

	req, err := http.NewRequest(http.MethodGet, d.URL, nil)
	req.Header.Add("Range", fmt.Sprintf("bytes=%v-%v", begin, end))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- err
		return
	}
	defer res.Body.Close()

	buf := make([]byte, BUF_SIZE)
	_, err = io.CopyBuffer(file, res.Body, buf)
	ch <- err
}

func (d DownloadClient) removePartFiles() error {
	dir := path.Dir(d.FilePath)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return ErrRemovingParts
	}

	filePattern := fmt.Sprintf(fmt.Sprintf(`^%v`, FILE_PART_NAME_TEMPLATE), path.Base(d.FilePath), `\d+$`)

	for _, f := range files {
		match, err := regexp.Match(filePattern, []byte(f.Name()))
		if err != nil {
			return ErrRemovingParts
		}
		if match {
			if err := os.Remove(path.Join(dir, f.Name())); err != nil {
				return ErrRemovingParts
			}
		}
	}

	return nil
}

func (d DownloadClient) mergeParts() error {
	file, err := os.OpenFile(d.FilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return ErrUnableToCreateDownloadFile
	}
	defer file.Close()

	buf := make([]byte, BUF_SIZE)
	for i := 0; i < int(d.Parts); i++ {
		partPath := fmt.Sprintf(FILE_PART_NAME_TEMPLATE, d.FilePath, i)
		filePart, err := os.Open(partPath)
		if err != nil {
			return ErrUnableToCreateDownloadFile
		}
		defer filePart.Close()

		if _, err := io.CopyBuffer(file, filePart, buf); err != nil {
			return err
		}
	}

	return d.removePartFiles()
}

func (d *DownloadClient) GetSize() (uint64, error) {
	if d.Size == 0 {
		res, err := http.Head(d.URL)
		if err != nil {
			return 0, ErrUnableToRequest
		}
		if res.StatusCode >= 300 {
			return 0, ErrStatusCodeNotOK
		}

		lenStr := res.Header.Get(CONTENT_LENGTH_HEADER)
		if lenStr == "" {
			log.Println("WARN:", ErrNoContentLength, "- Setting parallelism to 1...")
			d.Parts = 1
			return 0, nil
		}

		if res.Header.Get(ACCEPT_RANGES_HEADER) == "" {
			log.Println("WARN:", ErrNotAcceptRange, "- Setting parallelism to 1...")
			d.Parts = 1
			d.AcceptRange = false
			return 0, nil
		} else {
			d.AcceptRange = true
		}

		size, err := strconv.Atoi(lenStr)
		if err != nil {
			return 0, ErrInvalidContentLength
		}

		d.Size = uint64(size)
	}

	return d.Size, nil
}

func (d DownloadClient) Download() error {
	ch := make(chan error)
	defer close(ch)

	if d.Parts == 1 {
		go d.downloadPart(d.FilePath, 0, int(d.Size), ch)
		return <-ch
	}

	parts := int(d.Parts)
	sizePerPart := int(d.Size / d.Parts)

	for i := 0; i < parts; i++ {
		partPath := fmt.Sprintf(FILE_PART_NAME_TEMPLATE, d.FilePath, i)
		begin := i * sizePerPart
		end := (i+1)*sizePerPart - 1

		if i == parts-1 {
			end = int(d.Size)
		}

		go d.downloadPart(partPath, begin, end, ch)
	}

	errorList := make([]error, 0)
	for i := 0; i < parts; i++ {
		select {
		case err := <-ch:
			if err != nil {
				errorList = append(errorList, err)
			}
		}
	}

	if len(errorList) > 0 {
		msg := ""
		for _, err := range errorList {
			msg += fmt.Sprintln("-", err)
		}

		return errors.New(msg)
	}

	return d.mergeParts()
}

func NewDownloadClient(rawurl, filePath string, parts uint64, force bool) (*DownloadClient, error) {
	if _, err := os.Stat(filePath); err == nil {
		if !force {
			if parts > 1 {
				return nil, ErrDownloadFileAlreadyExists
			}
		} else if err := os.Remove(filePath); err != nil {
			return nil, ErrUnableToForceDownload
		}
	}

	if parts <= 0 {
		return nil, ErrInvalidNumberOfParts
	}

	if _, err := url.Parse(rawurl); err != nil {
		return nil, ErrInvalidURL
	}

	if force {
		fileName := path.Base(filePath)
		dir := path.Dir(filePath)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if strings.HasPrefix(f.Name(), fileName) {
				if err := os.Remove(path.Join(dir, f.Name())); err != nil {
					return nil, ErrUnableToForceDownload
				}
			}
		}
	}

	downloadClient := &DownloadClient{
		URL:      rawurl,
		FilePath: filePath,
		Parts:    parts,
	}

	if _, err := downloadClient.GetSize(); err != nil {
		return nil, err
	}

	return downloadClient, nil
}
