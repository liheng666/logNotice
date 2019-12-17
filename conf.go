package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
)

type Conf struct {
	Title string
	Path  []string
	Level []string
	Url string
}

// getConf 获取配置文件信息
func getConf() []Conf {
	matches, err := filepath.Glob("./conf/*.yaml")
	if err != nil {
		log.Fatal(err)
	}
	var conf []Conf
	for _, file := range matches {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		c := Conf{}
		err = yaml.Unmarshal(data, &c)
		if err != nil {
			log.Fatal(err)
		}
		conf = append(conf, c)
	}

	return conf

}
