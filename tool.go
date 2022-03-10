package main

import (
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"
)

func disk_free(device string) (uint64, error) {
	_, err := os.Stat(device)
	if err != nil {
		return 0, err
	}
	s, _ := disk.Usage(device)
	return s.Free / uint64(math.Pow(2, 30)), nil
}

func DeleteOldestFiles(device string) error {
	files, err := ioutil.ReadDir(device)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
	}
	return nil
}

func main() {
	for {
		file, err := os.OpenFile("E:/文件日志.txt",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Println(err)
		} else {
			log.SetOutput(file)
			device := "E:/"
			free, err := disk_free(device)
			if err != nil {
				log.Println(err)
			} else {
				if free < 20 {
					DeleteOldestFiles(device)
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}
