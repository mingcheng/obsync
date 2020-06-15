package util

import (
	"encoding/json"
	"log"
	"os"

	"github.com/mingcheng/obsync"
)

type Config struct {
	Debug      bool                  `json:"debug"`
	Force      bool                  `json:"force"`
	Root       string                `json:"root"`
	Standalone bool                  `json:"standalone"`
	Interval   uint                  `json:"interval"`
	Buckets    []obsync.BucketConfig `json:"buckets"`
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
