package worker

import (
	"encoding/json"
	"fmt"
	orglog "log"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/willie68/AutoRestIoT/model"
)

type MqttDatasource struct {
	Client                   mqtt.Client
	Broker                   string
	Backend                  string
	Destinations             []string
	Topic                    string
	Payload                  string
	TopicAttribute           string
	SimpleValueAttribute     string
	SimpleValueAttributeType string
	Rule                     string
}

var mqttClients = make([]MqttDatasource, 0)

func init() {
	//	mqtt.DEBUG = orglog.New(os.Stdout, "DEBUG", 0)
	mqtt.ERROR = orglog.New(os.Stdout, "ERROR", 0)
}

func mqttStoreMessage(datasource MqttDatasource, msg mqtt.Message) {
	//log.Infof("MODEL: %s.%s TOPIC: %s  MSG: %s", datasource.Backend, datasource.Model, msg.Topic(), msg.Payload())
	data, err := prepareMessage(datasource, msg)
	if err != nil {
		log.Alertf("%v", err)
		return
	}

	if datasource.TopicAttribute != "" {
		data[datasource.TopicAttribute] = datasource.Topic
	}

	data, err = executeTransformationrule(datasource, data)
	if err != nil {
		log.Alertf("%v", err)
		return
	}

	for _, destination := range datasource.Destinations {
		if strings.HasPrefix(destination, "$model.") {
			modelname := strings.TrimPrefix(destination, "$model.")
			route := model.Route{
				Backend: datasource.Backend,
				Model:   modelname,
			}
			Store(route, data)
		} else {
			err := Destinations.Store(datasource.Backend, destination, data)
			if err != nil {
				log.Alertf("%v", err)
				return
			}
		}
	}
}

func executeTransformationrule(datasource MqttDatasource, data model.JSONMap) (model.JSONMap, error) {
	if datasource.Rule != "" {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			log.Alertf("%v", err)
			return nil, err
		}
		newJson, err := Rules.TransformJSON(datasource.Backend, datasource.Rule, jsonBytes)
		if err != nil {
			log.Alertf("%v", err)
			return nil, err
		}

		data = nil
		err = json.Unmarshal(newJson, &data)
		if err != nil {
			log.Alertf("%v", err)
			return nil, err
		}
		fmt.Printf("src: %s\ndst: %s\n", string(jsonBytes), string(newJson))
	}
	return data, nil
}

func getSimpleDataAsModel(fieldname, fieldtype string, payload string) (model.JSONMap, error) {
	data := model.JSONMap{}
	var err error
	switch fieldtype {
	case model.FieldTypeInt:
		value, err := strconv.Atoi(payload)
		if err == nil {
			data[fieldname] = value
		}
	case model.FieldTypeFloat:
		value, err := strconv.ParseFloat(payload, 64)
		if err == nil {
			data[fieldname] = value
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
			data[fieldname] = value
		}
	case model.FieldTypeBool:
		value, err := strconv.ParseBool(payload)
		if err == nil {
			data[fieldname] = value
		}
	default:
		data[fieldname] = payload
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func prepareMessage(datasource MqttDatasource, msg mqtt.Message) (model.JSONMap, error) {
	var data model.JSONMap
	data = nil
	switch strings.ToLower(datasource.Payload) {
	case "application/json":
		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Alertf("%v", err)
			return nil, err
		}
	case "application/x.simple":
		added := false
		payload := string(msg.Payload())
		var err error
		data, err = getSimpleDataAsModel(datasource.SimpleValueAttribute, datasource.SimpleValueAttributeType, payload)
		if err != nil {
			log.Alertf("converting error on topic %s: %v", datasource.Topic, err)
			return nil, err
		}
		if !added {
			data[datasource.SimpleValueAttribute] = string(msg.Payload())
		}
	}
	return data, nil
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
	log.Infof("registering topic %s on %s for model %v", datasource.Topic, datasource.Broker, datasource.Destinations)
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
		mqttStoreMessage(datasource, m)
	})
	token.Wait()
	err := token.Error()
	return err
}

func getMQTTClient(config model.DataSourceConfigMQTT) {

}

func mqttRegisterTopic(clientID string, backendname string, datasource model.DataSource) error {
	destinationmodel := datasource.Destinations
	config := datasource.Config.(model.DataSourceConfigMQTT)

	opts := mqtt.NewClientOptions().AddBroker(config.Broker).SetClientID(clientID)
	opts.SetKeepAlive(2 * time.Second)
	//opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)
	opts.AutoReconnect = true
	datasourceMqtt := MqttDatasource{
		Broker:                   config.Broker,
		Backend:                  backendname,
		Destinations:             destinationmodel,
		Topic:                    config.Topic,
		Payload:                  config.Payload,
		TopicAttribute:           config.AddTopicAsAttribute,
		SimpleValueAttribute:     config.SimpleValueAttribute,
		SimpleValueAttributeType: config.SimpleValueAttributeType,
		Rule:                     datasource.Rule,
	}

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		mqttConnectionLost(datasourceMqtt, c, err)
	})
	if config.Username != "" {
		opts.CredentialsProvider = func() (string, string) {
			return config.Username, config.Password
		}
	}

	c := mqtt.NewClient(opts)
	datasourceMqtt.Client = c

	err := mqttReconnect(c)
	if err != nil {
		return err
	}

	mqttClients = append(mqttClients, datasourceMqtt)

	err = mqttSubscribe(datasourceMqtt)
	if err != nil {
		return err
	}

	log.Infof("registering topic %s on %s for model %s", config.Topic, config.Broker, destinationmodel)
	return nil
}
