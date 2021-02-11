package commands

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

const (
	BUF_SIZE              = 1024 * 8
	URL_FLAG_NAME         = "url"
	CHUNKS_FLAG_NAME      = "chunks"
	OUTPUT_FILE_FLAG_NAME = "output"
	OUTPUT_DIR_FLAG_NAME  = "dir"
	OVERWRITE_OUTPUT      = "overwrite"
)

type FileChunk struct {
	Begin int
	End   int
	Path  string
}

func mergeFiles(output string, parts []string) error {
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	buf := make([]byte, BUF_SIZE)
	for _, path := range parts {
		f, err := os.Open(path)
		if err != nil {
			return err
		}

		io.CopyBuffer(file, f, buf)
	}

	for _, path := range parts {
		os.Remove(path)
	}

	return nil
}

func downloadPart(url string, f FileChunk, ch chan<- error) {
	file, err := os.OpenFile(f.Path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
		ch <- err
		return
	}

	stat, err := file.Stat()
	if err != nil {
		ch <- err
		return
	}

	actualFileSize := int(stat.Size())
	if actualFileSize >= (f.End - f.Begin) {
		ch <- nil
		return
	}
	startFrom := f.Begin + actualFileSize

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		ch <- err
		return
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%v-%v", startFrom, f.End))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- err
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		ch <- ErrStatusCodeNotOK
		return
	}

	buf := make([]byte, BUF_SIZE)
	_, err = io.CopyBuffer(file, resp.Body, buf)
	ch <- err
}

func getFileSize(url string) (int, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	lenStr := resp.Header.Get("Content-Length")
	if lenStr == "" {
		return 0, ErrInvalidURL
	}

	fileSize, err := strconv.Atoi(lenStr)
	if err != nil {
		return 0, ErrInvalidURL
	}

	return fileSize, nil
}

func getFileChunks(size, nChunks int, outputPath string) []*FileChunk {
	chunks := make([]*FileChunk, nChunks)
	sizePerChunk := size / nChunks
	for i := 0; i < nChunks; i++ {
		chunks[i] = &FileChunk{
			Begin: i * sizePerChunk,
			End:   (i+1)*sizePerChunk - 1,
			Path:  fmt.Sprintf("%v.part%v", outputPath, i),
		}
	}

	lastChunk := chunks[len(chunks)-1]
	if lastChunk.End != size {
		lastChunk.End = size
	}

	return chunks
}

func single(url, output, dir string, nChunks int, overwrite bool) error {
	if output == "" {
		output = path.Base(url)
	}
	outputPath := path.Join(dir, output)

	_, err := os.Stat(outputPath)
	if err == nil {
		if !overwrite {
			return nil
		} else {
			os.Remove(outputPath)
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	fileSize, err := getFileSize(url)
	if err != nil {
		return err
	}

	chunks := getFileChunks(fileSize, nChunks, outputPath)

	finishCh := make(chan error)
	for _, chunk := range chunks {
		go downloadPart(url, *chunk, finishCh)
	}

	errorList := make([]error, 0)
	for i := 0; i < nChunks; i++ {
		select {
		case err := <-finishCh:
			if err != nil {
				errorList = append(errorList, err)
			}
		}
	}

	if len(errorList) > 0 {
		elms := make([]string, len(errorList))
		for _, e := range errorList {
			elms = append(elms, fmt.Sprintf("- %v;", e))
		}
		return errors.New(strings.Join(elms, "\n"))
	}

	parts := make([]string, 0)
	for _, p := range chunks {
		parts = append(parts, p.Path)
	}

	return mergeFiles(outputPath, parts)
}

func Single() *cli.Command {
	return &cli.Command{
		Name:    "single",
		Aliases: []string{"s"},
		Usage:   "Download a single file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     URL_FLAG_NAME,
				Aliases:  []string{"u"},
				Usage:    "Content URL",
				Required: true,
			},
			&cli.IntFlag{
				Name:    CHUNKS_FLAG_NAME,
				Aliases: []string{"c"},
				Value:   runtime.NumCPU() * 2,
				Usage:   "Number of parts to split the file to download",
			},
			&cli.StringFlag{
				Name:    OUTPUT_FILE_FLAG_NAME,
				Aliases: []string{"o"},
				Usage:   "The output file path",
			},
			&cli.PathFlag{
				Name:    OUTPUT_DIR_FLAG_NAME,
				Aliases: []string{"d"},
				Usage:   "The output dir path",
			},
			&cli.BoolFlag{
				Name:    OVERWRITE_OUTPUT,
				Aliases: []string{"f"},
				Value:   false,
				Usage:   "Overwrite the output file if already exists",
			},
		},
		Action: func(c *cli.Context) error {
			var err error
			url := c.String(URL_FLAG_NAME)
			fileName := c.String(OUTPUT_FILE_FLAG_NAME)
			dir := c.Path(OUTPUT_DIR_FLAG_NAME)
			if dir == "" {
				dir, err = os.Getwd()
				if err != nil {
					return err
				}
			}
			chunks := c.Int(CHUNKS_FLAG_NAME)
			overwrite := c.Bool(OVERWRITE_OUTPUT)
			return single(url, fileName, dir, chunks, overwrite)
		},
	}
}
