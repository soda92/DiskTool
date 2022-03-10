package main

import (
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func disk_free_raw(device string) (uint64, error) {
	_, err := os.Stat(device)
	if err != nil {
		return 0, err
	}
	s, _ := disk.Usage(device)
	return s.Free, nil
}

func CreateTestTree(sizeEmpty int, folderSizes []int) error {
	device := "E:/"
	topFolder := filepath.Join(device, "aaa")
	for i := 0; i < len(folderSizes); i++ {
		currentTime := time.Now().AddDate(0, 0, i)
		folder := currentTime.Format("2006-01-02")
		size := folderSizes[i]
		err := os.Mkdir(filepath.Join(topFolder, folder), 0755)
		if err != nil {
			return err
		}
		for j := 0; j < size; j++ {
			fileName := fmt.Sprintf("file%s.mp4", j)
			filePath := filepath.Join(topFolder, folder, fileName)
			f, err := os.Create(filePath)
			if err != nil {
				return err
			}
			err = f.Truncate(int64(math.Pow(2, 30)))
			if err != nil {
				return err
			}
			time.Sleep(1 * time.Second)
		}
	}
	sumSize := sizeEmpty
	for i := 0; i < len(folderSizes); i++ {
		sumSize += folderSizes[i]
	}
	free, _ := disk_free_raw(device)
	sizeLeft := free - uint64(sumSize)*uint64(math.Pow(2, 30))
	filePath2 := filepath.Join(topFolder, "left.file")
	f, err := os.Create(filePath2)
	if err != nil {
		return err
	}
	err = f.Truncate(int64(sizeLeft))
	if err != nil {
		return err
	}
	return nil
}

func TestDeleteDir(t *testing.T) {
	device := "E:/"
	sizes := []int{6}
	err := CreateTestTree(15, sizes)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err := DeleteOldestFiles(device)
	if err != nil {
		t.Fatalf("%v", err)
	}
	free, err := disk_free(device)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Equal(t, 19 < free && free < 21, true, "Free size should be around 20GB.")
	assert.Equal(t, DirNotExist(current()), true, "Oldest folder should be deleted.")
	assert.Equal(t, no_empty_dir(), true, "There should be no empty directories.")
}

func TestDeleteDir2(t *testing.T) {
	device := "E:/"
	sizes := []int{3, 15}
	err := CreateTestTree(15, sizes)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err := DeleteOldestFiles(device)
	if err != nil {
		t.Fatalf("%v", err)
	}
	free, err := disk_free(device)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Equal(t, 19 < free && free < 21, true, "Free size should be around 20GB.")
	assert.Equal(t, DirNotExist(current()), true, "Oldest folder should be deleted.")
	assert.Equal(t, no_empty_dir(), true, "There should be no empty directories.")
}

func TestDeleteDirPrivilaged(t *testing.T) {
	device := "E:/"
	sizes := []int{6, 15}
	sizes_err := []int{3}
	err := create_test_tree_privilaged(15, sizes, sizes_err)
	if err != nil {
		t.Fatalf(err)
	}
	err := DeleteOldestFiles(device)
	assert.Equal(t, err != nil, true, "There should be errors.")
	free, err := disk_free(device)
	assert.Equal(t, 19 < free && free < 21, true, "Free size should be around 20GB.")
	assert.Equal(t, DirSize(current()) == 3, true, "There should be 3 exceptions.")
}

func TestDeleteDirPrivilaged2(t *testing.T) {
	device := "E:/"
	sizes := []int{6, 4, 15}
	sizes_err := []int{3, 1}
	err := create_test_tree_privilaged(10, sizes, sizes_err)
	if err != nil {
		t.Fatalf(err)
	}
	err := DeleteOldestFiles(device)
	assert.Equal(t, err != nil, true, "There should be errors.")
	free, err := disk_free(device)
	assert.Equal(t, 19 < free && free < 21, true, "Free size should be around 20GB.")
	assert.Equal(t, DirSize(current()) == 3, true, "There should be 3 exceptions.")
}
