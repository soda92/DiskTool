package main

import (
	"fmt"
	"os"

	"github.com/shirou/gopsutil/disk"
	"gopkg.in/ini.v1"
)

func read_config() (string, string, error) {
	cfg, err := ini.Load("D:/LT_WXCL_Dll/LT_WXCLCFG.ini")
	if err != nil {
		return "", "", err
	}
	number := cfg.Section("LT_WXCLCFG").Key("TrainNum").String()
	device := cfg.Section("LT_WXCLCFG").Key("HDD").String()
	return device, number, nil
}

func disk_usage(device string) (float64, error) {
	_, err := os.Stat(device)
	if err != nil {
		return 0, err
	}
	s, _ := disk.Usage(device)
	return s.UsedPercent, nil
}

func main() {
	_, device, err := read_config()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	percent, err := disk_usage(device)

	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	if percent < 0.01 {
		fmt.Printf("Disk is full")
	} else {
		fmt.Printf("Disk not full")
	}

}
