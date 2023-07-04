package server

import (
	"database/sql"
	"log"
	"time"

	"github.com/luckyseadog/go-dev/internal/storage"
)

func PassSignal(cancelChan chan struct{}, chanStorage chan struct{}, envVariables *EnvVariables, storage *storage.MyStorage, db *sql.DB) {
	if envVariables.StoreInterval > 0 {
		backUpTicker := time.NewTicker(envVariables.StoreInterval)
		go func() {
			defer backUpTicker.Stop()
			for {
				select {
				case <-backUpTicker.C:
					if envVariables.DataSourceName != "" {
						err := storage.SaveToDB(db)
						if err != nil {
							log.Println(err)
						}
					} else {
						err := storage.SaveToFile(envVariables.StoreFile)
						if err != nil {
							log.Println(err)
						}
					}
				case <-cancelChan:
					return
				}
			}
		}()
	} else {
		go func() {
			for {
				select {
				case <-chanStorage:
					if envVariables.DataSourceName != "" {
						err := storage.SaveToDB(db)
						if err != nil {
							log.Println(err)
						}
					} else {
						err := storage.SaveToFile(envVariables.StoreFile)
						if err != nil {
							log.Println(err)
						}
					}
				case <-cancelChan:
					return
				}
			}
		}()
	}
}
