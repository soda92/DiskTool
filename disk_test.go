package main

import (
	"testing"
)

func TestConfigRead(t *testing.T) {
	device, number, err := read_config()
	if err != nil || number == "" || device == "" {
		t.Fatalf("%v, %v, %v", number, device, err)
	}
}

func TestDiskExists(t *testing.T) {
	device, _, _ := read_config()
	msg, err := disk_usage(device)
	if msg < 0.01 || err != nil {
		t.Fatalf(`%v, %v`, msg, err)
	}
}
