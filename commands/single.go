package commands

import (
	"errors"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/GabrielBG99/goget/filemgr"
	"github.com/urfave/cli/v2"
)

const (
	URL_FLAG_NAME         = "url"
	CHUNKS_FLAG_NAME      = "chunks"
	OUTPUT_FILE_FLAG_NAME = "output"
	OUTPUT_DIR_FLAG_NAME  = "dir"
	OVERWRITE_OUTPUT      = "overwrite"
)

func single(url, output, dir string, nChunks int, overwrite bool) error {
	filePath := path.Join(dir, output)
	client, err := filemgr.NewDownloadClient(url, filePath, uint64(nChunks), overwrite)
	if err != nil {
		if errors.Is(err, filemgr.ErrDownloadFileAlreadyExists) {
			log.Println("WARN:", err)
			return nil
		}
		return err
	}

	return client.Download()
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
				Usage:    "Content `URL`",
				Required: true,
			},
			&cli.IntFlag{
				Name:    CHUNKS_FLAG_NAME,
				Aliases: []string{"n"},
				Value:   runtime.NumCPU() * 2,
				Usage:   "Number of parts to split the file to download",
			},
			&cli.StringFlag{
				Name:    OUTPUT_FILE_FLAG_NAME,
				Aliases: []string{"o"},
				Usage:   "The output file name",
			},
			&cli.PathFlag{
				Name:    OUTPUT_DIR_FLAG_NAME,
				Aliases: []string{"d"},
				Usage:   "The output dir `PATH`",
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

			urlStr := c.String(URL_FLAG_NAME)

			fileName := c.String(OUTPUT_FILE_FLAG_NAME)
			if fileName == "" {
				fileName = path.Base(urlStr)
			}

			dir := c.Path(OUTPUT_DIR_FLAG_NAME)
			if dir == "" {
				dir, err = os.Getwd()
				if err != nil {
					return ErrUnableToDetectFolder
				}
			} else if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return ErrUnableToCreateFolder
			}

			chunks := c.Int(CHUNKS_FLAG_NAME)
			overwrite := c.Bool(OVERWRITE_OUTPUT)
			return single(urlStr, fileName, dir, chunks, overwrite)
		},
	}
}
