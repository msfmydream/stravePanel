package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

func TarGzUnzip(zipFile, dest string) error {
	fr, err := os.Open(zipFile)
	if err != nil {
		log.Error().Msg("open zipfile faild, can't find file:" + zipFile)
		return err
	}
	defer fr.Close()
	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	// 读取文件
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fw, err := os.OpenFile(dest+h.Name, os.O_CREATE|os.O_WRONLY, 0666 /*os.FileMode(h.Mode)*/)
		if err != nil {
			return err
		}
		defer fw.Close()
		_, err = io.Copy(fw, tr)
		if err != nil {
			return err
		}
	}
	return nil
}
