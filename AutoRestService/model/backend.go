package model

type Application struct {
	ApplicationName string  `yaml: "applicationname" json: "applicationname"`
	Description     string  `yaml: "description" json: "description"`
	Models          []Model `yaml: "models" json: "models"`
}

type Model struct {
	Name        string  `yaml: "name" json: "name"`
	Description string  `yaml: "description" json: "description"`
	Fields      []Field `yaml: "fields" json: "fields"`
	Indexes     []Index `yaml: "indexes" json: "indexes"`
}

type Field struct {
	Name       string `yaml: "name" json: "name"`
	Type       string `yaml: "type" json: "type"`
	Mandantory bool   `yaml: "mandantory" json: "mandantory"`
	Collection bool   `yaml: "collection" json: "collection"`
}

type Index struct {
	Name   string   `yaml: "name" json: "name"`
	Fields []string `yaml: "fields" json: "fields"`
}
