package worker

import (
	"fmt"
	"time"

	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
)

const SystemBackend = "_system"
const TaskModelName = "tasks"

var lastChecked time.Time
var backgroundConfig BackgroundConfig

//BackgroundConfig configuration of background tasks
type BackgroundConfig struct {
	Period              int
	DeleteOrphanedFiles bool
}

//InitBackgroundTasks initialise background tasks
func InitBackgroundTasks(config BackgroundConfig) {
	backgroundConfig = config
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

func GetTaskRoute() model.Route {
	return model.Route{
		Backend: SystemBackend,
		Model:   TaskModelName,
	}
}

//doTask internal function to process the background tasks
func doTask() {
	storage := dao.GetStorage()
	// prepare the backend models, getting all models and backends with file fields
	fileBackends := make([]map[string]string, 0)
	for _, k := range model.BackendList.Names() {
		backend, _ := model.BackendList.Get(k)
		for _, m := range backend.Models {
			for _, f := range m.Fields {
				if f.Type == model.FieldTypeFile {
					info := make(map[string]string)
					info["backend"] = k
					info["model"] = m.Name
					info["field"] = f.Name
					fileBackends = append(fileBackends, info)
				}
			}
		}
	}
	storage.ProcessFiles(func(info model.FileInfo) bool {
		if info.UploadDate.Add(1 * time.Hour).After(time.Now()) {
			return false
		}
		toDelete := true
		// log.Infof("found file: %s, id: %s, backend: %s", info.Filename, info.ID, info.Backend)
		// get the right backend
		for _, fileBackend := range fileBackends {
			if info.Backend == fileBackend["backend"] {
				route := model.Route{
					Backend: info.Backend,
					Model:   fileBackend["model"],
				}
				query := fmt.Sprintf("{ \"%s\": \"%s\"}", fileBackend["field"], info.ID)
				count, _, _ := storage.QueryModel(route, query, 0, 0)
				if count > 0 {
					toDelete = false
				}
			}
		}
		log.Infof("file has to be deleted: %s", toDelete)
		if toDelete && backgroundConfig.DeleteOrphanedFiles {
			storage.DeleteFile(info.Backend, info.ID)
		}
		return toDelete
	})
	lastChecked = time.Now()
}

func reportOrphanedFiles() {
	storage := dao.GetStorage()
	taskRoute := GetTaskRoute()

	task := model.Task{
		Type:   model.TaskOrphanedFilesReport,
		Status: model.New,
	}
	task, err := createTask(taskRoute, task)
	if err != nil {
		log.Alertf("error creating task: %v", err)
		return
	}

	files := make([]model.FileInfo, 0)
	// prepare the backend models, getting all models and backends with file fields
	fileBackends := make([]map[string]string, 0)
	for _, k := range model.BackendList.Names() {
		backend, _ := model.BackendList.Get(k)
		for _, m := range backend.Models {
			for _, f := range m.Fields {
				if f.Type == model.FieldTypeFile {
					info := make(map[string]string)
					info["backend"] = k
					info["model"] = m.Name
					info["field"] = f.Name
					fileBackends = append(fileBackends, info)
				}
			}
		}
	}
	task.Status = model.Running
	updateTask(taskRoute, task)

	err = storage.ProcessFiles(func(info model.FileInfo) bool {
		if info.UploadDate.Add(1 * time.Hour).After(time.Now()) {
			return false
		}
		toDelete := true
		// log.Infof("found file: %s, id: %s, backend: %s", info.Filename, info.ID, info.Backend)
		// get the right backend
		for _, fileBackend := range fileBackends {
			if info.Backend == fileBackend["backend"] {
				route := model.Route{
					Backend: info.Backend,
					Model:   fileBackend["model"],
				}
				query := fmt.Sprintf("{ \"%s\": \"%s\"}", fileBackend["field"], info.ID)
				count, _, _ := storage.QueryModel(route, query, 0, 0)
				if count > 0 {
					toDelete = false
				}
			}
		}
		log.Infof("file has to be deleted: %s", toDelete)
		if toDelete {
			files = append(files, info)
		}
		return toDelete
	})
	if err != nil {
		log.Alertf("error: %s\n", err.Error())
		return
	}
	task.Data = model.JSONMap{}
	task.Data["fileids"] = files
	task.Status = model.Finished
	updateTask(taskRoute, task)

}

func updateTask(taskRoute model.Route, task model.Task) {
	taskRoute.Identity = task.ID
	jsonMap, err := task.ToJSONMap()
	if err != nil {
		return
	}
	_, err = dao.GetStorage().UpdateModel(taskRoute, jsonMap)
	if err != nil {
		log.Alertf("error updating task: %v", err)
		return
	}
}

func createTask(taskRoute model.Route, task model.Task) (model.Task, error) {
	jsonMap, err := task.ToJSONMap()
	if err != nil {
		return model.Task{}, err
	}
	id, err := dao.GetStorage().CreateModel(taskRoute, jsonMap)
	if err != nil {
		log.Alertf("error creating task: %v", err)
		return model.Task{}, err
	}

	task.ID = id
	return task, nil
}
