package records

import (
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
)

func Play(ctx *gin.Context) {
	filename := ctx.Query("filename")

	var err error
	var file *os.File
	var fileinfo os.FileInfo

	if filepath.Ext(filename) != ".cast" {
		goto FAIL
	}

	if file, err = os.Open(filename); err != nil {
		goto FAIL
	}
	defer file.Close()

	if fileinfo, err = os.Lstat(filename); err != nil {
		ctx.Status(500)
		return
	}

	ctx.DataFromReader(200, fileinfo.Size(), "application/octet-stream", file, make(map[string]string))
	return

FAIL:
	ctx.Status(500)
	return
}