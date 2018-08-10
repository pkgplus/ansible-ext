package server

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

var jobDict = map[string]string{
	"NODE":  "node_exporter",
	"LINUX": "node_exporter",
	"REDIS": "redis_exporter",
	"MYSQL": "mysql_exporter",
}

func GetPlayBookPath(job string) string {
	var bookpath = job
	filename, BookDictfound := jobDict[job]
	if !strings.HasSuffix(job, ".yml") && strings.Index(job, "/") == -1 {
		if BookDictfound {
			bookpath = fmt.Sprintf("playbook/%s.yml", filename)
		} else {
			bookpath = fmt.Sprintf("playbook/%s.yml", job)
		}
	}
	if !PathExists(bookpath) {
		if !strings.HasSuffix(job, ".yml") && strings.Index(job, "/") == -1 {
			if BookDictfound {
				bookpath = fmt.Sprintf("playbook/playbook.d/%s.yml", filename)
			} else {
				bookpath = fmt.Sprintf("playbook/playbook.d/%s.yml", job)
			}
		}
	}

	return bookpath
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

type PlayBook struct {
	Name  string                   `yaml:"name"`
	Tasks []map[string]interface{} `yaml:"tasks"`
}

func LoadPlayBook(path string) ([]PlayBook, error) {
	pbs := make([]PlayBook, 0)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return pbs, err
	}

	err = yaml.Unmarshal(data, &pbs)
	return pbs, err
}
