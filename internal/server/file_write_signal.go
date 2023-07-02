package server

import (
	"log"
	"time"

	"github.com/luckyseadog/go-dev/internal/storage"
)

func PassSignal(cancelChan chan struct{}, chanStorage chan struct{}, envVariables *EnvVariables, storage storage.Storage) {
	if envVariables.StoreInterval > 0 {
		backUpTicker := time.NewTicker(envVariables.StoreInterval)
		go func() {
			defer backUpTicker.Stop()
			for {
				select {
				case <-backUpTicker.C:
					err := storage.SaveToFile(envVariables.StoreFile)
					if err != nil {
						log.Println(err)
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
					err := storage.SaveToFile(envVariables.StoreFile)
					if err != nil {
						log.Println(err)
					}
				case <-cancelChan:
					return
				}
			}
		}()
	}
}
