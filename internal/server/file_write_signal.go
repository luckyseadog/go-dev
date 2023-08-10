package server

import (
	"time"

	"github.com/luckyseadog/go-dev/internal/storage"
)

func PassSignal(cancelChan chan struct{}, chanStorage chan struct{}, envVariables *EnvVariables, s storage.Storage) {
	if envVariables.StoreInterval > 0 {
		backUpTicker := time.NewTicker(envVariables.StoreInterval)
		go func() {
			defer backUpTicker.Stop()
			for {
				select {
				case <-backUpTicker.C:
					if envVariables.DataSourceName == "" {
						if ms, ok := s.(*storage.MyStorage); ok {
							err := ms.SaveToFile(envVariables.StoreFile)
							if err != nil {
								MyLog.Println(err)
							}
						} else {
							MyLog.Println(storage.ErrNotMyStorage)
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
					if envVariables.DataSourceName == "" {
						if ms, ok := s.(*storage.MyStorage); ok {
							err := ms.SaveToFile(envVariables.StoreFile)
							if err != nil {
								MyLog.Println(err)
							}
						} else {
							MyLog.Println(storage.ErrNotMyStorage)
						}
					}
				case <-cancelChan:
					return
				}
			}
		}()
	}
}
