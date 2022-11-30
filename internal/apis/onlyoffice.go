package apis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-uuid"
	"wwwj.dev/wujiao/onlyoffice-simple-client/internal/utils/config"
	"wwwj.dev/wujiao/onlyoffice-simple-client/internal/utils/req"
	"wwwj.dev/wujiao/onlyoffice-simple-client/internal/utils/resp"
)

type OnlyofficeConfig struct {
	Document     map[string]interface{} `json:"document"`
	EditorConfig map[string]interface{} `json:"editorConfig"`
	jwt.StandardClaims
}

type FileMeta struct {
	OriginName  string
	FileName    string
	Id          string
	UpdatedTime int64
	Suffix      string
	FileType    string
}

func suffixToType(suffix string) string {
	switch suffix {
	case "doc", "docm", "docx", "docxf", "dot", "dotm", "dotx", "epub", "fodt", "fb2", "htm", "html", "mht", "odt", "oform", "ott", "oxps", "pdf", "rtf", "txt", "djvu", "xml", "xps":
		return "word"
	case "csv", "fods", "ods", "ots", "xls", "xlsb", "xlsm", "xlsx", "xlt", "xltm", "xltx":
		return "cell"
	case "fodp", "odp", "otp", "pot", "potm", "potx", "pps", "ppsm", "ppsx", "ppt", "pptm", "pptx":
		return "slide"
	default:
		return ""
	}
}

func ParseFilename(originFileName string) (*FileMeta, error) {
	// ID,urlencode(<originName>),<UpdatedTime>.<suffix>
	t := strings.Split(originFileName, ",")
	if len(t) != 3 {
		return nil, errors.New("wrong file names")
	}
	id := t[0]
	updatedTimeStr := strings.Split(t[2], ".")
	if len(updatedTimeStr) != 2 {
		return nil, errors.New("wrong updatetime")
	}
	updatedTime, err := strconv.Atoi(updatedTimeStr[0])
	if err != nil {
		return nil, err
	}
	suffix := updatedTimeStr[1]
	filename, err := url.QueryUnescape(t[1])
	if err != nil {
		return nil, err
	}
	fileType := suffixToType(suffix)
	if fileType == "" {
		return nil, errors.New("file type not vaild")
	}
	result := &FileMeta{
		OriginName:  originFileName,
		FileName:    filename,
		Id:          id,
		UpdatedTime: int64(updatedTime),
		Suffix:      suffix,
		FileType:    fileType,
	}

	return result, nil
}

func buildConfig(originFileName string) (string, string, *FileMeta, error) {
	fileMeta, err := ParseFilename(originFileName)
	if err != nil {
		return "", "", nil, err
	}

	t := OnlyofficeConfig{
		Document: map[string]interface{}{
			"fileType": fileMeta.Suffix,
			"key":      fileMeta.Id, // not use cache TODO: Use Cache
			"title":    fileMeta.FileName,
			"url":      fmt.Sprintf("%s/files/%s", config.Conf.Server.Url, fileMeta.OriginName),
			"permissions": map[string]interface{}{
				"comment":                 false,
				"copy":                    true,
				"deleteCommentAuthorOnly": false,
				"download":                true,
				"edit":                    true,
				"editCommentAuthorOnly":   false,
				"fillForms":               true,
				"modifyContentControl":    true,
				"modifyFilter":            true,
				"print":                   true,
				"review":                  false,
			},
		},
		EditorConfig: map[string]interface{}{
			"callbackUrl": fmt.Sprintf("%s/callback", config.Conf.Server.Url),
			"mode":        "edit",
			"user": map[string]string{
				"group": "AdminGroup",
				"id":    "1",
				"name":  "Admin",
			},
			"lang":     "zh",
			"location": "zh",
			"customization": map[string]interface{}{
				"goback": map[string]interface{}{
					"blank":        false,
					"requestClose": false,
					"text":         "Finish & Close",
					"url":          config.Conf.Server.Url,
				},
			},
			"region":        "zh-CN",
			"hideRightMenu": true,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, t)
	tokenResult, err := token.SignedString([]byte(config.Conf.Onlyoffice.Secret))
	config, _ := json.Marshal(t)
	log.Println(string(config))

	return string(config), tokenResult, fileMeta, nil
}

func GenConfig(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		resp.Error(c, errors.New("filename blank"))
		return
	}

	fileConfig, token, fileMeta, err := buildConfig(filename)
	if err != nil {
		resp.Error(c, err)
		return
	}

	resp.OK(c, map[string]interface{}{
		"documentserver": config.Conf.Onlyoffice.Host,
		"secret":         token,
		"config":         fileConfig,
		"documentType":   fileMeta.FileType,
	})
}

func JsonMarshalNoSetEscapeHTML(data interface{}) ([]byte, error) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	if err := jsonEncoder.Encode(data); err != nil {
		return nil, err
	}

	return bf.Bytes(), nil
}

func Callback(c *gin.Context) {
	var result map[string]interface{}
	var err error
	defer func() {
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, map[string]int{
				"error": 1,
			})
		}
		return
	}()
	if err = c.ShouldBind(&result); err != nil {
		// 处理错误请求
		return
	}
	if tokenString, ok := result["token"]; ok {
		result, err := tokenVerify(tokenString.(string))
		if err != nil {
			return
		}
		if !result {
			err = errors.New("auth failed")
		}
	} else {
		err = errors.New("auth required")
	}
	status := int(result["status"].(float64))
	switch status {
	case 2, 3:
		url, ok := result["url"]
		if !ok {
			err = errors.New("no url found")
			return
		}
		key, ok := result["key"]
		if !ok {
			err = errors.New("no key found")
			return
		}
		log.Println(url, key)
		go func() {
			err := handleFileChange(key.(string), url.(string))
			if err != nil {
				log.Println(err)
			}
		}()
		// 保存并退出 需要下载修改后的文件
		// 根据返回的ID 将原本这个ID的文件重命名，然后新文件存为原来的ID
	default:
		log.Println(result)
	}

	c.JSON(http.StatusOK, map[string]int{
		"error": 0,
	})
}

func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			resp.Error(c, errors.New("auth failed"))
			c.Abort()
			return
		}
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		result, err := tokenVerify(tokenString)
		if err != nil {
			resp.Error(c, errors.New("auth failed"))
			c.Abort()
			return
		}

		if result {
			log.Println("token vaild")
			c.Next()
			return
		} else {
			resp.Error(c, errors.New("auth failed"))
			c.Abort()
			return
		}

	}
}

func tokenVerify(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Conf.Onlyoffice.Secret), nil
	})
	if err != nil {
		return false, err
	}
	// 检查token是否合法
	return token.Valid, nil
}

func handleFileChange(oldKey string, newFileUrl string) error {
	// find old file
	rd, err := ioutil.ReadDir(DataPath)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return err
	}

	for _, fi := range rd {
		if !fi.IsDir() && strings.HasPrefix(fi.Name(), oldKey) {
			newKey, err := uuid.GenerateUUID()
			if err != nil {
				return err
			}
			// move un-edit file
			oldFileMeta, err := ParseFilename(fi.Name())
			if err != nil {
				return err
			}

			newFilename := fmt.Sprintf("%s,%s,%s.%s", newKey, oldFileMeta.FileName, strconv.FormatInt(oldFileMeta.UpdatedTime, 10), oldFileMeta.Suffix)
			log.Println(fi.Name(), newFilename)
			err = os.Rename(fmt.Sprintf("%s/%s", DataPath, fi.Name()), fmt.Sprintf("%s/%s", DataPath, newFilename))

			newKey, err = uuid.GenerateUUID()
			if err != nil {
				return err
			}
			downloadFileName := fmt.Sprintf("%s,%s,%s.%s", newKey, oldFileMeta.FileName, strconv.FormatInt(time.Now().Unix(), 10), oldFileMeta.Suffix)
			err = req.Download(newFileUrl, fmt.Sprintf("%s/%s", DataPath, downloadFileName))
			log.Println("download finished")
			RefreshLocalFileList()
			return err
		}
	}
	return errors.New("file not found")
}
