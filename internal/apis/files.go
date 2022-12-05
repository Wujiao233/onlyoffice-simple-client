package apis

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-uuid"
	"wwwj.dev/wujiao/onlyoffice-simple-client/internal/utils/resp"
)

const DataPath = "./data"

var LocalFileMap map[string][]LocalFileHistory

func init() {
	RefreshLocalFileList()
}

type LocalFileMeta struct {
	Filename string             `json:"filename"`
	Version  []LocalFileHistory `json:"version"`
}

type LocalFileHistory struct {
	OriginName string
	UpdateTime string
}

func RefreshLocalFileList() error {
	var err error
	LocalFileMap, err = fetchLocalFilelist()
	if err != nil {
		return err
	}
	return nil
}

func fetchLocalFilelist() (map[string][]LocalFileHistory, error) {
	fileMap := make(map[string][]LocalFileHistory)

	rd, err := ioutil.ReadDir(DataPath)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return nil, err
	}

	for _, fi := range rd {
		if !fi.IsDir() {
			fileMeta, err := ParseFilename(fi.Name())
			if err != nil {
				continue
			}
			originFileName := fmt.Sprintf("%s.%s", fileMeta.FileName, fileMeta.Suffix)
			if value, ok := fileMap[originFileName]; ok {
				fileMap[originFileName] = append(value, LocalFileHistory{
					OriginName: fileMeta.OriginName,
					UpdateTime: time.Unix(fileMeta.UpdatedTime, 0).Format("2006-01-02 15:04:05"),
				})
			} else {
				fileMap[originFileName] = []LocalFileHistory{
					LocalFileHistory{
						OriginName: fileMeta.OriginName,
						UpdateTime: time.Unix(fileMeta.UpdatedTime, 0).Format("2006-01-02 15:04:05"),
					},
				}
			}
		}
	}
	return fileMap, nil
}

func generateFileName(originFileName string) (string, error) {
	ext := strings.Split(originFileName, ".")
	if len(ext) < 1 {
		return "", errors.New("wrong file name")
	}
	suffix := ext[len(ext)-1]
	fileBaseName := strings.Join(ext[:len(ext)-1], ".")

	id, err := uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	timeStamp := time.Now().Unix()

	newFileName := fmt.Sprintf("%s,%s,%s.%s", id, fileBaseName, strconv.FormatInt(timeStamp, 10), suffix)
	return newFileName, nil
}

func Del(c *gin.Context) {
	var postReq map[string]interface{}

	if err := c.ShouldBind(&postReq); err != nil {
		resp.Error(c, errors.New("err"))
		return
	}

	fileName, ok := postReq["file"]
	if !ok {
		resp.Error(c, errors.New("empty filename"))
		return
	}
	fileNameStr := strings.Split(fileName.(string), "/")[0]

	if err := os.Remove("./data/" + fileNameStr); err == nil {
		RefreshLocalFileList()
		resp.OK(c, "ok")
		return
	} else {
		resp.Error(c, err)
	}
}

func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		resp.Error(c, err)
		return
	}

	newFileName, err := generateFileName(file.Filename)
	if err != nil {
		resp.Error(c, err)
		return
	}

	filePath := fmt.Sprintf("%s/", DataPath)
	_, err = os.Stat(filePath)
	if err != nil {
		if !os.IsExist(err) { // 目录不存在，创建目录
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	filePath = filePath + newFileName
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		resp.Error(c, err)
		return
	}
	go RefreshLocalFileList()
	resp.OK(c, newFileName)
}

func FileList(c *gin.Context) {
	resp.OK(c, LocalFileMap)
}
