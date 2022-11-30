package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const DefaultConfPath string = "./configs/config.json"

var Conf Configuration = read()

func read() Configuration {
	file, _ := os.Open(DefaultConfPath)
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := Configuration{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf(conf.String())
	return conf
}

func (conf *Configuration) String() string {
	b, err := json.Marshal(*conf)
	if err != nil {
		return fmt.Sprintf("%+v", *conf)
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	if err != nil {
		return fmt.Sprintf("%+v", *conf)
	}
	return out.String()
}
