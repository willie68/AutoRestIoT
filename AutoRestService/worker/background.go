package worker

import (
	"time"

	"github.com/willie68/AutoRestIoT/dao"
)

var lastChecked time.Time
var period int

//BackgroundConfig configuration of background tasks
type BackgroundConfig struct {
	Period int
}

//InitBackgroundTasks initialise background tasks
func InitBackgroundTasks(config BackgroundConfig) {
	period := config.Period
	log.Infof("healthcheck starting with period: %d seconds", period)
	if period > 0 {
		go func() {
			background := time.NewTicker(time.Second * time.Duration(period))
			for range background.C {
				doTask()
			}
		}()
	}
}

//doTask internal function to process the background tasks
func doTask() {
	storage := dao.GetStorage()
	storage.ProcessFiles(func(filename, id, backend string) bool {
		log.Infof("found file: %s, id: %s, backend: %s", filename, id, backend)
		return true
	})
	lastChecked = time.Now()
}
