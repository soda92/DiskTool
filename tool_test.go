package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
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
	size_per_file := math.Pow(2, 20)
	topFolder := filepath.Join(device, "aaa")
	for i := 0; i < len(folderSizes); i++ {
		currentTime := time.Now().AddDate(0, 0, i)
		folder := currentTime.Format("2006-01-02")
		size := folderSizes[i]
		err := os.MkdirAll(filepath.Join(topFolder, folder), 0755)
		if err != nil {
			return err
		}
		for j := 0; j < size; j++ {
			fileName := fmt.Sprintf("file%d.mp4", j)
			filePath := filepath.Join(topFolder, folder, fileName)
			f, err := os.Create(filePath)
			if err != nil {
				return err
			}
			err = f.Truncate(int64(size_per_file))
			if err != nil {
				return err
			}
			f.Close()
			time.Sleep(10 * time.Millisecond)
		}
	}
	free, _ := disk_free_raw(device)
	sizeLeft := free - uint64(sizeEmpty)*uint64(size_per_file)
	filePath2 := filepath.Join(topFolder, "left.file")
	f, err := os.Create(filePath2)
	if err != nil {
		return err
	}
	err = f.Truncate(int64(sizeLeft))
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func create_test_tree_privilaged(sizeEmpty int, folderSizes []int, errorSizes []int) error {
	device := "E:/"
	topFolder := filepath.Join(device, "aaa")
	size_per_file := math.Pow(2, 20)
	for i := 0; i < len(folderSizes); i++ {
		currentTime := time.Now().AddDate(0, 0, i)
		folder := currentTime.Format("2006-01-02")
		err := os.MkdirAll(filepath.Join(topFolder, folder), 0755)
		if err != nil {
			return err
		}
		size := folderSizes[i]
		if i < len(errorSizes) {
			size2 := errorSizes[i]
			size -= size2
			for j := 0; j < size2; j++ {
				fileName := fmt.Sprintf("file_priv_%d.mp4", j)
				filePath := filepath.Join(topFolder, folder, fileName)
				f, err := os.Create(filePath)
				if err != nil {
					return err
				}
				err = f.Truncate(int64(size_per_file))
				if err != nil {
					return err
				}
				f.Close()
				time.Sleep(10 * time.Millisecond)
			}
		}
		for j := 0; j < size; j++ {
			fileName := fmt.Sprintf("file%d.mp4", j)
			filePath := filepath.Join(topFolder, folder, fileName)
			f, err := os.Create(filePath)
			if err != nil {
				return err
			}
			err = f.Truncate(int64(size_per_file))
			if err != nil {
				return err
			}
			f.Close()
			time.Sleep(10 * time.Millisecond)
		}
	}
	free, _ := disk_free_raw(device)
	sizeLeft := free - uint64(sizeEmpty)*uint64(size_per_file)
	filePath2 := filepath.Join(topFolder, "left.file")
	f, err := os.Create(filePath2)
	if err != nil {
		return err
	}
	err = f.Truncate(int64(sizeLeft))
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func current() string {
	device := "E:/"
	topFolder := filepath.Join(device, "aaa")
	currentTime := time.Now()
	folder := currentTime.Format("2006-01-02")
	return filepath.Join(topFolder, folder)
}

func DirNotExist(path string) bool {
	_, err := os.Stat(path)
	return err != nil
}

func no_empty_dir() bool {
	device := "E:/"
	topFolder := filepath.Join(device, "aaa")
	files, err := ioutil.ReadDir(topFolder)
	if err != nil {
		return false
	}
	for _, file := range files {
		if file.IsDir() {
			files, err := ioutil.ReadDir(filepath.Join(topFolder, file.Name()))
			if err != nil {
				return false
			}
			if len(files) == 0 {
				return false
			}
		}
	}
	return true
}

func DirSize(dir string) int {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var size int64 = 0
	for _, file := range files {
		if !file.IsDir() {
			stat, err := os.Stat(filepath.Join(dir, file.Name()))
			if err != nil {
				log.Fatal(err)
			}
			size += stat.Size()
		}
	}
	return int(size / int64(math.Pow(2, 20)))
}

func clean(device string) error {
	files, err := ioutil.ReadDir(device)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, topFolder := range files {
		if topFolder.Name() == "$RECYCLE.BIN" || topFolder.Name() == "System Volume Information" {
			continue
		}
		topFolderPath := filepath.Join(device, topFolder.Name())
		err = os.RemoveAll(topFolderPath)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func TestDeleteDir(t *testing.T) {
	device := "E:/"
	clean(device)
	required_free := 20
	sizes := []int{6}
	err := CreateTestTree(15, sizes)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var db *sql.DB = nil
	_, err = DeleteOldestFiles(device, required_free, db)
	if err != nil {
		t.Fatalf("%v", err)
	}
	free, err := disk_free(device)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Equal(t, uint64(required_free)-1 <= free && free <= uint64(required_free)+1,
		true, "Free size should be around required.")
	assert.Equal(t, DirNotExist(current()), false, "Oldest folder should not be deleted.")
	assert.Equal(t, no_empty_dir(), true, "There should be no empty directories.")
}

func TestDeleteDir2(t *testing.T) {
	device := "E:/"
	clean(device)
	required_free := 20
	sizes := []int{3, 15}
	err := CreateTestTree(15, sizes)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var db *sql.DB = nil
	_, err = DeleteOldestFiles(device, required_free, db)
	if err != nil {
		t.Fatalf("%v", err)
	}
	free, err := disk_free(device)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Equal(t, uint64(required_free)-1 <= free && free <= uint64(required_free)+1,
		true, "Free size should be around required.")
	assert.Equal(t, DirNotExist(current()), true, "Oldest folder should be deleted.")
	assert.Equal(t, no_empty_dir(), true, "There should be no empty directories.")
}

func TestDeleteDirPrivilaged(t *testing.T) {
	device := "E:/"
	clean(device)
	required_free := 20
	sizes := []int{6, 15}
	sizes_err := []int{3}
	err := create_test_tree_privilaged(15, sizes, sizes_err)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var db *sql.DB = nil
	_, err = DeleteOldestFiles(device, required_free, db)
	assert.Equal(t, err == nil, true, "There should be no errors.")
	free, err := disk_free(device)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Equal(t, uint64(required_free)-1 <= free && free <= uint64(required_free)+1,
		true, "Free size should be around required.")
	assert.Equal(t, DirSize(current()) == 3, true, "There should be 3 exceptions.")
}

func TestDeleteDirPrivilaged2(t *testing.T) {
	device := "E:/"
	clean(device)
	required_free := 20
	sizes := []int{6, 4, 15}
	sizes_err := []int{3, 1}
	err := create_test_tree_privilaged(10, sizes, sizes_err)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var db *sql.DB = nil
	_, err = DeleteOldestFiles(device, required_free, db)
	assert.Equal(t, err == nil, true, "There should be no errors.")
	free, err := disk_free(device)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Equal(t, uint64(required_free)-1 <= free && free <= uint64(required_free)+1,
		true, "Free size should be around required.")
	assert.Equal(t, DirSize(current()) == 3, true, "There should be 3 exceptions.")
}

func testLog() {
	log.Println("aaa")
}
func TestLogWrite(t *testing.T) {
	file, err := os.OpenFile("文件日志.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(err)
	} else {
		log.SetOutput(file)
		testLog()
	}
}
