package main

import (
	"stravePanel/backend/configs"
	"stravePanel/backend/utils"
)

func main() {

	strURL := "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz"
	storepath := "/env/dev/code/go/stravePanel/download"
	filename := "steamcmd_linux.tar.gz"
	sliceNum := 3
	resume := true
	configs.LogInit()

	utils.NewDownloader(sliceNum, resume).Download(strURL, storepath, filename)

}
