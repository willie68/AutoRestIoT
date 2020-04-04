package models

type Application struct {
	ApplicationName string    `yaml: "applicationname" json: "applicationname"`
	Models          []BEModel `yaml: "models" json: "models"`
}

type BEModel struct {
	Name    string  `yaml: "name" json: "name"`
	Fields  []Field `yaml: "fields" json: "fields"`
	Indexes []Index `yaml: "indexes" json: "indexes"`
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
