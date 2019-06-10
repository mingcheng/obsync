package util

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Debug     bool   `json:"debug"`
	Secret    string `json:"secret"`
	Key       string `json:"key"`
	EndPoint  string `json:"endpoint"`
	Bucket    string `json:"bucket"`
	MaxThread uint   `json:"thread"`
	Force     bool   `json:"force"`
	Root      string `json:"root"`
	Timeout   uint   `json:"timeout"`
}

func (c *Config) Read(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	file, _ := os.Open(path)
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(c); err != nil {
		return err
	} else {
		if c.Debug {
			log.Println("read configure file from path " + path)
		}

		return nil
	}
}

func (c *Config) Dump() (config string, err error) {
	if result, err := json.Marshal(c); err != nil {
		return "", nil
	} else {
		return string(result), nil
	}
}
