package model

//DataSourceConfigMQTT definition of the special configuration of a mqtt datasource
type DataSourceConfigMQTT struct {
	Broker                   string `yaml:"broker" json:"broker"`
	Topic                    string `yaml:"topic" json:"topic"`
	QoS                      int    `yaml:"qos" json:"qos"`
	Payload                  string `yaml:"payload" json:"payload"`
	Username                 string `yaml:"username" json:"username"`
	Password                 string `yaml:"password" json:"password"`
	AddTopicAsAttribute      string `yaml:"addTopicAsAttribute" json:"addTopicAsAttribute"`
	SimpleValueAttribute     string `yaml:"simpleValueAttribute" json:"simpleValueAttribute"`
	SimpleValueAttributeType string `yaml:"simpleValueAttributeType" json:"simpleValueAttributeType"`
}
