package model

import (
	"encoding/json"

	"github.com/willie68/AutoRestIoT/logging"
)

const TaskOrphanedFilesReport = "orphanedFilesReport"
const TaskOrphanedFilesDelete = "orphanedFilesDelete"

type TaskStatus string

const (
	New      TaskStatus = "new"
	Running             = "running"
	Finished            = "finished"
)

type Task struct {
	ID     string     `yaml:"id" json:"id"`
	Name   string     `yaml:"name" json:"name"`
	Type   string     `yaml:"ttype" json:"ttype"`
	Status TaskStatus `yaml:"tstatus" json:"tstatus"`
	File   string     `yaml:"tfile" json:"tfile"`
	Data   JSONMap    `yaml:"tdata" json:"tdata"`
}

var log logging.ServiceLogger

//ToJSONMap converting task to json map
func (t *Task) ToJSONMap() (JSONMap, error) {
	taskModel := JSONMap{}
	jsonData, err := json.Marshal(t)
	if err != nil {
		log.Alertf("error: %v", err)
		return nil, err
	}
	err = json.Unmarshal(jsonData, &taskModel)
	if err != nil {
		log.Alertf("error: %v", err)
		return nil, err
	}
	return taskModel, nil
}
