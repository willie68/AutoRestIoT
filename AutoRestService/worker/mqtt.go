package worker

import (
	"encoding/json"
	orglog "log"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/willie68/AutoRestIoT/model"
)

type MqttDatasource struct {
	Client               mqtt.Client
	Broker               string
	Backend              string
	Model                string
	Topic                string
	Payload              string
	TopicAttribute       string
	SimpleValueAttribute string
}

var mqttClients = make([]MqttDatasource, 0)

func init() {
	//	mqtt.DEBUG = orglog.New(os.Stdout, "DEBUG", 0)
	mqtt.ERROR = orglog.New(os.Stdout, "ERROR", 0)
}

func mqttStoreMessage(datasource MqttDatasource, client mqtt.Client, msg mqtt.Message) {
	//log.Infof("MODEL: %s.%s TOPIC: %s  MSG: %s", datasource.Backend, datasource.Model, msg.Topic(), msg.Payload())
	route := model.Route{
		Backend: datasource.Backend,
		Model:   datasource.Model,
	}
	var data model.JSONMap
	data = nil
	switch strings.ToLower(datasource.Payload) {
	case "application/json":
		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Alertf("%v", err)
			return
		}
	case "application/x.simple":
		added := false
		data = model.JSONMap{}
		modelDef, ok := model.BackendList.GetModel(route)
		payload := string(msg.Payload())
		if ok {
			field, ok := modelDef.GetField(datasource.SimpleValueAttribute)
			if ok {
				switch field.Type {
				case model.FieldTypeString:
					data[datasource.SimpleValueAttribute] = payload
					added = true
				case model.FieldTypeInt:
					value, err := strconv.Atoi(payload)
					if err == nil {
						data[datasource.SimpleValueAttribute] = value
						added = true
					} else {
						log.Alertf("route %s: converting error on topic %s: %v", route.String(), datasource.Topic, err)
					}
				case model.FieldTypeFloat:
					value, err := strconv.ParseFloat(payload, 64)
					if err == nil {
						data[datasource.SimpleValueAttribute] = value
						added = true
					} else {
						log.Alertf("route %s: converting error on topic %s: %v", route.String(), datasource.Topic, err)
					}
				case model.FieldTypeTime:
					value, err := time.Parse(time.RFC3339, payload)
					if err != nil {
						saveerr := err
						var vint int
						vint, err = strconv.Atoi(payload)
						if err == nil {
							value = time.Unix(0, int64(vint)*int64(time.Millisecond))
						} else {
							err = saveerr
						}
					}
					if err == nil {
						data[datasource.SimpleValueAttribute] = value
						added = true
					} else {
						log.Alertf("route %s: converting error on topic %s: %v", route.String(), datasource.Topic, err)
					}
				case model.FieldTypeBool:
					value, err := strconv.ParseBool(payload)
					if err == nil {
						data[datasource.SimpleValueAttribute] = value
						added = true
					} else {
						log.Alertf("route %s: converting error on topic %s: %v", route.String(), datasource.Topic, err)
					}
				}
			}
		}
		if !added {
			data[datasource.SimpleValueAttribute] = string(msg.Payload())
		}
	}
	if datasource.TopicAttribute != "" {
		data[datasource.TopicAttribute] = datasource.Topic
	}

	Store(route, data)
}

func mqttConnectionLost(datasource MqttDatasource, c mqtt.Client, e error) {
	connected := false
	for !connected {
		err := mqttReconnect(c)
		if err != nil {
			log.Alertf("%v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		connected = c.IsConnected()
	}

	subscribed := false
	for !subscribed {
		if !c.IsConnected() {
			mqttReconnect(c)
		}
		err := mqttSubscribe(datasource)
		if err != nil {
			log.Alertf("%v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		subscribed = true
	}
	log.Infof("registering topic %s on %s for model %s", datasource.Topic, datasource.Broker, datasource.Model)
}

func mqttReconnect(c mqtt.Client) error {
	if !c.IsConnected() {
		token := c.Connect()
		token.Wait()
		err := token.Error()
		return err
	}
	return nil
}

func mqttSubscribe(datasource MqttDatasource) error {
	token := datasource.Client.Subscribe(datasource.Topic, 0, func(c mqtt.Client, m mqtt.Message) {
		mqttStoreMessage(datasource, c, m)
	})
	token.Wait()
	err := token.Error()
	return err
}

func mqttRegisterTopic(clientID string, backendname string, destinationmodel string, config model.DataSourceConfigMQTT) error {
	opts := mqtt.NewClientOptions().AddBroker(config.Broker).SetClientID(clientID)
	opts.SetKeepAlive(2 * time.Second)
	//opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)
	opts.AutoReconnect = true
	datasource := MqttDatasource{
		Broker:               config.Broker,
		Backend:              backendname,
		Model:                destinationmodel,
		Topic:                config.Topic,
		Payload:              config.Payload,
		TopicAttribute:       config.AddTopicAsAttribute,
		SimpleValueAttribute: config.SimpleValueAttribute,
	}

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		mqttConnectionLost(datasource, c, err)
	})
	if config.Username != "" {
		opts.CredentialsProvider = func() (string, string) {
			return config.Username, config.Password
		}
	}

	c := mqtt.NewClient(opts)
	datasource.Client = c

	err := mqttReconnect(c)
	if err != nil {
		return err
	}

	mqttClients = append(mqttClients, datasource)

	err = mqttSubscribe(datasource)
	if err != nil {
		return err
	}

	log.Infof("registering topic %s on %s for model %s", config.Topic, config.Broker, destinationmodel)
	return nil
}
