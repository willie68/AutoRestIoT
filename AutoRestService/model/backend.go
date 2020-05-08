package model

import (
	"errors"
	"sort"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Backend struct for definition of a backend
type Backend struct {
	Backendname  string        `yaml:"backendname" json:"backendname"`
	Description  string        `yaml:"description" json:"description"`
	Models       []Model       `yaml:"models" json:"models"`
	DataSources  []DataSource  `yaml:"datasources" json:"datasources"`
	Rules        []Rule        `yaml:"rules" json:"rules"`
	Destinations []Destination `yaml:"destinations" json:"destinations"`
}

//Model definition of a single model
type Model struct {
	Name        string  `yaml:"name" json:"name"`
	Description string  `yaml:"description" json:"description"`
	Fields      []Field `yaml:"fields" json:"fields"`
	Indexes     []Index `yaml:"indexes" json:"indexes"`
}

//FieldTypeString field type string
const FieldTypeString = "string"

//FieldTypeInt field type integer
const FieldTypeInt = "int"

//FieldTypeFloat field type float
const FieldTypeFloat = "float"

//FieldTypeTime field type time
const FieldTypeTime = "time"

//FieldTypeBool field type bool
const FieldTypeBool = "bool"

//FieldTypeMap field type map
const FieldTypeMap = "map"

//FieldTypeFile field type file
const FieldTypeFile = "file"

//Field definition of a field
type Field struct {
	Name       string `yaml:"name" json:"name"`
	Type       string `yaml:"type" json:"type"`
	Mandatory  bool   `yaml:"mandatory" json:"mandatory"`
	Collection bool   `yaml:"collection" json:"collection"`
}

//Index definition of an index
type Index struct {
	Name   string   `yaml:"name" json:"name"`
	Unique bool     `yaml:"unique" json:"unique"`
	Fields []string `yaml:"fields" json:"fields"`
}

//DataSource Definition of a datasource
type DataSource struct {
	Name         string      `yaml:"name" json:"name"`
	Type         string      `yaml:"type" json:"type"`
	Destinations []string    `yaml:"destinations" json:"destinations"`
	Rule         string      `yaml:"rule" json:"rule"`
	Config       interface{} `yaml:"config" json:"config"`
}

type Destination struct {
	Name   string      `yaml:"name" json:"name"`
	Type   string      `yaml:"type" json:"type"`
	Config interface{} `yaml:"config" json:"config"`
}

type Rule struct {
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description" json:"description"`
	Transform   interface{} `yaml:"transform" json:"transform"`
}

//ErrModelDefinitionNotFound model definition not found
var ErrModelDefinitionNotFound = errors.New("model defintion not found")

//BackendList list of all definied backends
var BackendList = NewBackends()

//Backends backend list object
type Backends struct {
	bs map[string]Backend
}

//NewBackends creating a new Backend list
func NewBackends() Backends {
	b := Backends{
		bs: make(map[string]Backend),
	}
	return b
}

//Names getting all backendnames
func (m *Backends) Names() []string {
	names := make([]string, 0)
	for name := range m.bs {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	return names
}

//Contains checking if the manufacturer name is present in the list of manufacturers
func (m *Backends) Contains(name string) bool {
	for k := range m.bs {
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

//Clear clearing the list
func (m *Backends) Clear() {
	m.bs = make(map[string]Backend)
}

//IsValidDatamodel checking if a data model is valid
func (b *Backend) IsValidDatamodel(model string, data JSONMap) bool {
	return true
}

//GetReferencedFiles getting a list of ids of referenced files
func (b *Backend) GetReferencedFiles(modelname string, data JSONMap) ([]string, error) {
	model, ok := b.GetModel(modelname)
	if !ok {
		return nil, ErrModelDefinitionNotFound
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

//GetModel getting a model definition from the backend definition
func (m *Backends) GetModel(route Route) (Model, bool) {
	backend, ok := m.Get(route.Backend)
	if !ok {
		return Model{}, false
	}
	return backend.GetModel(route.Model)
}

//GetModel getting a model definition from the backend definition
func (b *Backend) GetModel(modelname string) (Model, bool) {
	for _, model := range b.Models {
		if model.Name == modelname {
			return model, true
		}
	}
	return Model{}, false
}

//GetRule getting a rule definition from the backend definition
func (b *Backend) GetRule(rulename string) (Rule, bool) {
	for _, rule := range b.Rules {
		if rule.Name == rulename {
			return rule, true
		}
	}
	return Rule{}, false
}

//GetField getting a field definition from the model definition
func (m *Model) GetField(fieldname string) (Field, bool) {
	for _, field := range m.Fields {
		if field.Name == fieldname {
			return field, true
		}
	}
	return Field{}, false
}

//GetIndex getting a index definition from the model definition
func (m *Model) GetIndex(indexname string) (Index, bool) {
	for _, index := range m.Indexes {
		if index.Name == indexname {
			return index, true
		}
	}
	return Index{}, false
}

//GetFieldNames getting a list of fieldnames from the model definition
func (m *Model) GetFieldNames() []string {
	fields := make([]string, 0)
	for _, field := range m.Fields {
		fields = append(fields, field.Name)
	}
	return fields
}
