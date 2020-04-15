package model

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Backend struct {
	Backendname string       `yaml: "backendname" json: "backendname"`
	Description string       `yaml: "description" json: "description"`
	Models      []Model      `yaml: "models" json: "models"`
	DataSources []DataSource `yaml: "datasources" json: "datasources"`
}

type Model struct {
	Name        string  `yaml: "name" json: "name"`
	Description string  `yaml: "description" json: "description"`
	Fields      []Field `yaml: "fields" json: "fields"`
	Indexes     []Index `yaml: "indexes" json: "indexes"`
}

const FieldTypeString = "string"
const FieldTypeInt = "int"
const FieldTypeFloat = "float"
const FieldTypeTime = "time"
const FieldTypeBool = "bool"
const FieldTypeMap = "map"
const FieldTypeFile = "file"

type Field struct {
	Name       string `yaml: "name" json: "name"`
	Type       string `yaml: "type" json: "type"`
	Mandatory  bool   `yaml: "mandatory" json: "mandatory"`
	Collection bool   `yaml: "collection" json: "collection"`
}

type Index struct {
	Name   string   `yaml: "name" json: "name"`
	Fields []string `yaml: "fields" json: "fields"`
}

type DataSource struct {
	Name        string      `yaml: "name" json: "name"`
	Type        string      `yaml: "type" json: "type"`
	Destination string      `yaml: "destination" json: "destination"`
	Config      interface{} `yaml: "config" json: "config"`
}

type DataSourceConfigMQTT struct {
	Broker  string `yaml: "broker" json: "broker"`
	Topic   string `yaml: "topic" json: "topic"`
	Payload string `yaml: "payload" json: "payload"`
}

var ErrModelNotFound = errors.New("model not found")
var BackendList = NewBackends()

//Backends backend list object
type Backends struct {
	bs map[string]Backend
}

func NewBackends() Backends {
	b := Backends{
		bs: make(map[string]Backend),
	}
	return b
}

//Contains checking if the manufacturer name is present in the list of manufacturers
func (m *Backends) Contains(name string) bool {
	for k, _ := range m.bs {
		if k == name {
			return true
		}
	}
	return false
}

//Add adding a new manufacturer to the list
func (m *Backends) Add(backend Backend) string {
	m.bs[backend.Backendname] = backend
	return backend.Backendname
}

//Remove remove a single tag
func (m *Backends) Remove(name string) {
	if m.Contains(name) {
		delete(m.bs, name)
	}
}

//Get getting a tag
func (m *Backends) Get(name string) (Backend, bool) {
	for k, be := range m.bs {
		if k == name {
			return be, true
		}
	}
	return Backend{}, false
}

//Cleear clearing the list
func (m *Backends) Clear() {
	m.bs = make(map[string]Backend)
}

func (b *Backend) IsValidDatamodel(model string, data JsonMap) bool {
	return true
}

func (b *Backend) GetReferencedFiles(modelname string, data JsonMap) ([]string, error) {
	model, ok := b.GetModel(modelname)
	if !ok {
		return nil, ErrModelNotFound
	}
	files := make([]string, 0)
	for _, field := range model.Fields {
		if field.Type == FieldTypeFile {
			dataValue := data[field.Name]
			if dataValue != nil {
				switch v := dataValue.(type) {
				case primitive.A:
					values := v
					for _, d := range values {
						files = append(files, d.(string))
					}
				case []interface{}:
					values := v
					for _, d := range values {
						files = append(files, d.(string))
					}
				case []string:
					values := v
					for _, d := range values {
						files = append(files, d)
					}
				case string:
					files = append(files, v)
				}

			}
		}
	}

	return files, nil
}

func (b *Backends) GetModel(route Route) (Model, bool) {
	backend, ok := b.Get(route.Backend)
	if !ok {
		return Model{}, false
	}
	return backend.GetModel(route.Model)
}

func (b *Backend) GetModel(modelname string) (Model, bool) {
	for _, model := range b.Models {
		if model.Name == modelname {
			return model, true
		}
	}
	return Model{}, false
}

func (m *Model) GetField(fieldname string) (Field, bool) {
	for _, field := range m.Fields {
		if field.Name == fieldname {
			return field, true
		}
	}
	return Field{}, false
}

func (m *Model) GetIndex(indexname string) (Index, bool) {
	for _, index := range m.Indexes {
		if index.Name == indexname {
			return index, true
		}
	}
	return Index{}, false
}

func (m *Model) GetFieldNames() []string {
	fields := make([]string, 0)
	for _, field := range m.Fields {
		fields = append(fields, field.Name)
	}
	return fields
}
