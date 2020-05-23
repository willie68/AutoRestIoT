package model

import "time"

//JSONMap structure for json objects
type JSONMap map[string]interface{}

//FileInfo type
type FileInfo struct {
	Filename   string
	ID         string
	Backend    string
	UploadDate time.Time
}
