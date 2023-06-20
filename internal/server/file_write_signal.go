package server

import "time"

func PassSignal(cancelChan chan struct{}, fileSaveChan chan time.Time, storeInterval time.Duration) {
	if storeInterval > 0 {
		backUpTicker := time.NewTicker(storeInterval)
		go func() {
			defer backUpTicker.Stop()
			for {
				select {
				case fileSaveChan <- <-backUpTicker.C:
				case <-cancelChan:
					return
				default:
					<-backUpTicker.C
				}
			}
		}()
	} else {
		go func() {
			for {
				select {
				case fileSaveChan <- time.Now():
				case <-cancelChan:
					return
				}
			}
		}()
	}
}
