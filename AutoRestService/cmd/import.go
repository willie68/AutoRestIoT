package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
	"github.com/willie68/AutoRestIoT/worker"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func importData(importPath string) {
	count := 0
	dir := importPath
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info != nil {
			if info.IsDir() {
				filepath := path + "/schematic.json"
				_, err := os.Stat(filepath)
				if !os.IsNotExist(err) {
					count++
					schematic := getSchematic(filepath)
					if schematic.Owner == "" {
						schematic.Owner = "w.klaas@gmx.de"
					}
					fileids := make([]string, 0)
					for filename, _ := range schematic.Files {
						file := path + "/" + filename
						f, err := os.Open(file)
						if err != nil {
							fmt.Printf("error: %s\n", err.Error())
						}
						defer f.Close()
						reader := bufio.NewReader(f)

						fileid, err := dao.GetStorage().AddFile("schematics", filename, reader)
						if err != nil {
							fmt.Printf("%v\n", err)
						} else {
							fmt.Printf("fileid: %s\n", fileid)
							fileids = append(fileids, fileid)
						}
					}
					bemodel := model.JsonMap{}

					bemodel["Foreignid"] = schematic.ID.Hex()
					bemodel["Description"] = schematic.Description
					bemodel["Manufacturer"] = schematic.Manufacturer
					bemodel["Model"] = schematic.Model
					bemodel["Owner"] = schematic.Owner
					bemodel["PrivateFile"] = schematic.PrivateFile
					bemodel["Subtitle"] = schematic.SubTitle
					bemodel["Tags"] = schematic.Tags
					bemodel["Files"] = fileids

					route := model.Route{
						Backend:  "schematics",
						Model:    "schematic",
						Username: "w.klaas@gmx.de",
						SystemID: "autorest-srv",
					}

					id, err := worker.Store(route, bemodel)
					if err != nil {
						fmt.Printf("%v\n", err)
					}
					fmt.Printf("%d: found %s: man: %s, model: %s\n", count, id["_id"], schematic.Manufacturer, schematic.Model)
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func getSchematic(file string) Schematic {
	jsonFile, err := os.Open(file)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	//	fmt.Println(string(byteValue))

	var schematic Schematic
	err = json.Unmarshal(byteValue, &schematic)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return schematic
}

func uploadFile(filename string, file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	fileid, err := dao.GetStorage().AddFile("schematics", filename, reader)
	if err != nil {
		return "", err
	}
	return fileid, nil
}

type Schematic struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ForeignID      string             `json:"foreignId" bson:"foreignId,omitempty"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt,omitempty"`
	LastModifiedAt time.Time          `json:"lastModifiedAt" bson:"lastModifiedAt,omitempty"`
	Manufacturer   string             `json:"manufacturer" bson:"manufacturer,omitempty"`
	Model          string             `json:"model" bson:"model,omitempty"`
	SubTitle       string             `json:"subtitle" bson:"subtitle,omitempty"`
	Tags           []string           `json:"tags" bson:"tags,omitempty"`
	Description    string             `json:"description" bson:"description,omitempty"`
	PrivateFile    bool               `json:"privateFile" bson:"privateFile,omitempty"`
	Owner          string             `json:"owner" bson:"owner,omitempty"`
	Files          map[string]string  `json:"files" bson:"files,omitempty"`
	BuildIn        time.Time          `json:"buildIn" bson:"buildIn,omitempty"`
	BuildTO        time.Time          `json:"buildTO" bson:"buildTO,omitempty"`
}

//UnmarshalJSON unmarshall a json to a schematic with some user properties
func (s *Schematic) UnmarshalJSON(data []byte) error {
	var dat map[string]interface{}
	if err := json.Unmarshal(data, &dat); err != nil {
		return err
	}
	if dat["id"] != nil {
		id, _ := primitive.ObjectIDFromHex(dat["id"].(string))
		s.ID = id
	}
	if dat["foreignId"] != nil {
		s.ForeignID = dat["foreignId"].(string)
	}
	if dat["createdAt"] != nil {
		switch v := dat["createdAt"].(type) {
		case string:
			layout := "2006-01-02T15:04:05.000Z"
			s.CreatedAt, _ = time.Parse(layout, v)
		case float64:
			s.CreatedAt = time.Unix(0, int64(v)*int64(time.Millisecond))
		}
	}
	if dat["lastModifiedAt"] != nil {
		switch v := dat["lastModifiedAt"].(type) {
		case string:
			layout := "2006-01-02T15:04:05.000Z"
			s.LastModifiedAt, _ = time.Parse(layout, v)
		case float64:
			s.LastModifiedAt = time.Unix(0, int64(v)*int64(time.Millisecond))
		}
	}
	if dat["manufacturer"] != nil {
		s.Manufacturer = dat["manufacturer"].(string)
	}
	if dat["model"] != nil {
		s.Model = dat["model"].(string)
	}
	if dat["subtitle"] != nil {
		s.SubTitle = dat["subtitle"].(string)
	}
	if dat["tags"] != nil {
		values := dat["tags"].([]interface{})
		s.Tags = make([]string, len(values))
		for i, d := range values {
			s.Tags[i] = d.(string)
		}
	}
	if dat["description"] != nil {
		s.Description = dat["description"].(string)
	}
	if dat["privateFile"] != nil {
		switch v := dat["privateFile"].(type) {
		case float64:
			s.PrivateFile = v > 0
		case int:
			s.PrivateFile = v > 0
		case bool:
			s.PrivateFile = v
		}
	}
	if dat["owner"] != nil {
		s.Owner = dat["owner"].(string)
	}
	if dat["files"] != nil {
		switch v := dat["files"].(type) {
		case []interface{}:
			values := v
			s.Files = make(map[string]string)
			for _, d := range values {
				s.Files[d.(string)] = ""
			}
		case map[string]interface{}:
			values := v
			s.Files = make(map[string]string)
			for k, d := range values {
				s.Files[k] = d.(string)
			}
		}
	}
	//	if dat["buildIn"] != nil {
	//		s.BuildIn = dat["buildIn"].(string)
	//	}
	//	if dat["buildTO"] != nil {
	//		s.BuildTO = dat["buildTO"].(string)
	//	}
	return nil
}
