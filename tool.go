package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shirou/gopsutil/disk"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func disk_free(device string) (uint64, error) {
	_, err := os.Stat(device)
	if err != nil {
		return 0, err
	}
	s, _ := disk.Usage(device)
	return s.Free / uint64(math.Pow(2, 20)), nil
}

func fileExists(filePath string, db *sql.DB) (bool, error) {
	stmt, err := db.Prepare("SELECT * FROM filesinfo WHERE path=?")
	if err != nil {
		log.Println(err)
		return false, err
	}
	rows, err := stmt.Query(filePath)
	if err != nil {
		log.Println(err)
		return false, err
	}
	if rows.Next() {
		return true, nil
	}
	rows.Close()
	return false, nil
}

func DeleteOldestFiles(device string, reqFree int, db *sql.DB) (*sql.DB, error) {
	files, err := ioutil.ReadDir(device)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if db == nil {
		db, err = sql.Open("sqlite3", "file::memory:")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		_, err = db.Exec(strings.ReplaceAll(
			`CREATE TABLE IF NOT EXISTS +filesinfo+ (
				+path+ VARCHAR(500) PRIMARY KEY NOT NULL,
				+created+ DATE NOT NULL,
				+folder+ varchar(500) NOT NULL,
				+error+ BOOLEAN NOT NULL
			);`, "+", "`"))
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	for _, topFolder := range files {
		if topFolder.Name() == "$RECYCLE.BIN" || topFolder.Name() == "System Volume Information" {
			continue
		}
		if topFolder.IsDir() {
			topFolderPath := filepath.Join(device, topFolder.Name())
			files2, err := ioutil.ReadDir(topFolderPath)
			if err != nil {
				log.Println(err)
				return db, err
			}
			for _, dateFolder := range files2 {
				if dateFolder.IsDir() {
					dateFolderPath := filepath.Join(topFolderPath, dateFolder.Name())
					files3, err := ioutil.ReadDir(dateFolderPath)
					if err != nil {
						log.Println(err)
						return db, err
					}
					for _, files := range files3 {
						if !files.IsDir() {
							file := files
							filePath := filepath.Join(dateFolderPath, file.Name())
							exists, err := fileExists(filePath, db)
							if err != nil {
								log.Println(err)
								return db, err
							}
							if exists {
								continue
							}
							stmt, err := db.Prepare("INSERT INTO filesinfo(path,created,folder,error) values(?,?,?,?)")
							if err != nil {
								log.Println(err)
								return db, err
							}
							modifiedDate, err := os.Stat(filePath)
							if err != nil {
								log.Println(err)
								return db, err
							}
							_, err = stmt.Exec(filePath, modifiedDate.ModTime(), dateFolderPath, 0)
							if err != nil {
								log.Println(err)
								return db, err
							}
							stmt.Close()
						}
					}
				}
			}
		}
	}
	possibleEmptyFolders := make(map[string]bool)
	for {
		free, err := disk_free(device)
		if err != nil {
			log.Println(err)
			return db, err
		}
		if int(free) >= reqFree {
			break
		} else {
			for {
				rows, err := db.Query("SELECT * FROM filesinfo WHERE error=false ORDER BY created LIMIT 1")
				if err != nil {
					log.Println(err)
					return db, err
				}
				var path string
				var created time.Time
				var folder string
				var _err bool
				if rows.Next() {
					err = rows.Scan(&path, &created, &folder, &_err)
					possibleEmptyFolders[folder] = true
					if err != nil {
						log.Println(err)
						return db, err
					}
					rows.Close()
					if strings.Contains(path, "priv") {
						stmt, err := db.Prepare("UPDATE filesinfo SET error=? WHERE path=?")
						if err != nil {
							log.Println(err)
							return db, err
						}
						_, err = stmt.Exec(true, path)
						if err != nil {
							log.Println(err)
							return db, err
						}
						stmt.Close()
						continue
					}

					err = os.Remove(path)
					if err == nil {
						stmt, err := db.Prepare("DELETE FROM filesinfo WHERE path=?")
						if err != nil {
							log.Println(err)
							return db, err
						}
						_, err = stmt.Exec(path)
						if err != nil {
							log.Println(err)
							return db, err
						}
						stmt.Close()
						break
					} else {
						log.Println(err)
						stmt, err := db.Prepare("UPDATE filesinfo SET error=1 WHERE path=?")
						if err != nil {
							log.Println(err)
							return db, err
						}
						_, err = stmt.Exec(path)
						if err != nil {
							log.Println(err)
							return db, err
						}
						stmt.Close()
						continue
					}
				}
			}
		}
	}
	for folder := range possibleEmptyFolders {
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			log.Println(err)
			return db, err
		}
		empty := true
		for range files {
			empty = false
		}
		if empty {
			os.Remove(folder)
		}
	}
	return db, nil
}

func main() {
	var db *sql.DB = nil
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
				if free < 20*1024 {
					db, err = DeleteOldestFiles(device, 20*1024, db)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}
