package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type Download struct {
	FileId             int
	ContentLength      int
	SliceNum           int
	CurrentSliceNum    int
	CurrentSliceLength int
}

func NewDownloader(sliceNum int) *Download {
	return &Download{SliceNum: sliceNum}
}

func (pThis *Download) Download(url, storepath, filename string) error {

	//1. 获取响应头、判断是否支持分片下载
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	//2.根据响应头，选择分片下载/整个下载

	if resp.StatusCode == http.StatusOK && resp.Header.Get("Accept-Ranges") == "bytes" {
		//分片下载
		return pThis.mutiDownLoad(url, storepath, filename, int(resp.ContentLength))
	} else if resp.StatusCode == http.StatusOK {
		//单独下载
		return pThis.singalDownLoad(url, storepath, filename)
	} else {
		//访问资源失败
		return errors.New("DownLoad Failed from " + url)
	}

}

func (pThis *Download) mutiDownLoad(url, storepath, filename string, contentLength int) error {

	//计算每一分片的大小
	partSize := contentLength / pThis.SliceNum
	filepath := path.Join(storepath, filename)
	//判断目录是否已经存在，如不存在则创建
	if _, err := os.Stat(storepath); os.IsNotExist(err) {
		log.Info().Msg("[download] : Path is Not Exist, Create is. ")
		// 创建部分文件的存放目录
		partDir := pThis.getPartDir(filepath)
		os.Mkdir(partDir, 0777)
	}

	var wg sync.WaitGroup
	wg.Add(pThis.SliceNum)

	rangeStart := 0

	for i := 0; i < pThis.SliceNum; i++ {
		// 并发请求
		go func(i, rangeStart int) {
			defer wg.Done()

			rangeEnd := rangeStart + partSize
			// 最后一部分，总长度不能超过 ContentLength
			if i == pThis.SliceNum-1 {
				rangeEnd = pThis.ContentLength
			}

			pThis.downloadPartial(url, filepath, rangeStart, rangeEnd, i)

		}(i, rangeStart)

		rangeStart += partSize + 1
	}

	wg.Wait()

	// 合并文件
	pThis.merge(storepath)

	return nil
}

func (pThis *Download) singalDownLoad(strURL, storepath, filename string) error {

	filepath := path.Join(storepath, filename)
	//判断目录是否已经存在，如不存在则创建
	if _, err := os.Stat(storepath); os.IsNotExist(err) {
		log.Info().Msg("[download] : Path is Not Exist, Create is. ")
		// 创建部分文件的存放目录
		partDir := pThis.getPartDir(filepath)
		os.Mkdir(partDir, 0777)
	}

	req, err := http.NewRequest("GET", strURL, nil)
	if err != nil {
		log.Error().Msgf("[download] : Create Request Failed:" + err.Error())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		log.Error().Msgf("[download] : Could'd not Send Requset or Server UnAvailable:" + err.Error())
	}
	defer resp.Body.Close()

	flags := os.O_CREATE | os.O_WRONLY
	singleFile, err := os.OpenFile((filepath), flags, 0666)
	if err != nil {
		log.Error().Msgf("[download] : FilePath not find:" + err.Error())
	}
	defer singleFile.Close()
	_, err = io.Copy(singleFile, resp.Body)
	if err != nil {
		if err == io.EOF {
			return errors.New("保存文件失败")
		}
	}
	log.Info().Msg("download with singalDownload Success. Save as: " + filepath)
	return nil
}

func (pThis *Download) merge(storepath string) error {
	destFile, err := os.OpenFile(storepath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer destFile.Close()

	for i := 0; i < pThis.SliceNum; i++ {
		partFileName := pThis.getPartFilename(storepath, i)
		partFile, err := os.Open(partFileName)
		if err != nil {
			return err
		}
		io.Copy(destFile, partFile)
		partFile.Close()
		os.Remove(partFileName)
	}

	return nil
}

func (pThis *Download) downloadPartial(strURL, filepath string, byteStart, byteEnd, i int) {
	if byteStart >= byteEnd {
		return
	}

	req, err := http.NewRequest("GET", strURL, nil)
	if err != nil {
		log.Error().Msgf("[download] : Create Request Failed:" + err.Error())
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", byteStart, byteEnd))
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		log.Error().Msgf("[download] : Could'd not Send Requset or Server UnAvailable:" + err.Error())
	}
	defer resp.Body.Close()

	flags := os.O_CREATE | os.O_WRONLY
	partFile, err := os.OpenFile(pThis.getPartFilename(filepath, i), flags, 0666)
	if err != nil {
		log.Error().Msgf("[download] : FilePath not find:" + err.Error())
	}
	defer partFile.Close()

	buf := make([]byte, 32*1024)
	_, err = io.CopyBuffer(partFile, resp.Body, buf)
	if err != nil {
		if err == io.EOF {
			return
		}
	}
	log.Info().Msg("download with singalDownload Success. Save as: " + filepath)
}

// getPartFilename 构造部分文件的名字
func (pThis *Download) getPartFilename(filepath string, partNum int) string {
	partDir := pThis.getPartDir(filepath)
	return fmt.Sprintf("%s/%s-%d", partDir, filepath, partNum)
}

// getPartDir 部分文件存放的目录
func (pThis *Download) getPartDir(filepath string) string {
	// 找到最后一个斜杠的位置
	lastSlashIndex := strings.LastIndex(filepath, "/")

	// 如果没有找到斜杠，返回空字符串
	if lastSlashIndex == -1 {
		return ""
	}

	// 截取从开始到斜杠之前的部分作为目录路径
	dirPath := filepath[:lastSlashIndex]

	return dirPath
}
