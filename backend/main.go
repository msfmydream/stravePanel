package main

import (
	"stravePanel/backend/utils"
)

func main() {

	strURL := "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz"
	storepath := "/env/dev/code/go/stravePanel/download"
	filename := "steamcmd_linux.tar.gz"
	sliceNum := 3

	utils.NewDownloader(sliceNum).Download(strURL, storepath, filename)

}
