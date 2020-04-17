package worker

import (
	orglog "log"
	"os"
	"time"

	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
)

func init() {
	//	mqtt.DEBUG = orglog.New(os.Stdout, "DEBUG", 0)
	mqtt.ERROR = orglog.New(os.Stdout, "ERROR", 0)
}

//ValidateBackend validate if a backend definition is valid
func ValidateBackend(be model.Backend) error {
	// checking backendname format
	// checking models
	// checking indexes
	return nil
}

func RegisterBackend(backend model.Backend) error {
	// create indexes if missing
	models := backend.Models
	for _, bemodel := range models {
		err := createIndex(bemodel, backend.Backendname)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	// creating source plugins
	for _, datasource := range backend.DataSources {
		ok := false
		for !ok {
			err := createDatasource(datasource, backend.Backendname)
			if err != nil {
				log.Fatalf("%v", err)
				time.Sleep(10 * time.Second)
				continue
			}
			ok = true
		}
	}
	return nil
}

func createDatasource(datasource model.DataSource, backendname string) error {
	switch datasource.Type {
	case "mqtt":
		err := mqttRegisterTopic(datasource.Name, backendname, datasource.Destination, datasource.Config.(model.DataSourceConfigMQTT))
		if err != nil {
			return err
		}
	default:
		log.Alertf("type \"%s\" is not availble as data source type", datasource.Type)
	}
	return nil
}

type MqttDatasource struct {
	Client  mqtt.Client
	Broker  string
	Backend string
	Model   string
	Topic   string
	Payload string
}

var mqttClients = make([]MqttDatasource, 0)

func mqttStoreMessage(datasource MqttDatasource, client mqtt.Client, msg mqtt.Message) {
	//log.Infof("MODEL: %s.%s TOPIC: %s  MSG: %s", datasource.Backend, datasource.Model, msg.Topic(), msg.Payload())
	route := model.Route{
		Backend: datasource.Backend,
		Model:   datasource.Model,
	}
	if datasource.Payload == "application/json" {
		var data model.JsonMap
		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Alertf("%v", err)
		} else {
			Store(route, data)
		}
	}
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
		Broker:  config.Broker,
		Backend: backendname,
		Model:   destinationmodel,
		Topic:   config.Topic,
		Payload: config.Payload,
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

func createIndex(bemodel model.Model, backendname string) error {
	indexes := bemodel.Indexes
	// define stardard fulltext index
	_, ok := bemodel.GetIndex(dao.FulltextIndexName)
	if !ok {
		fulltextIndex := model.Index{
			Name:   dao.FulltextIndexName,
			Fields: bemodel.GetFieldNames(),
		}
		indexes = append(indexes, fulltextIndex)
	}
	// define stardard indexes
	for _, field := range bemodel.Fields {
		_, ok := bemodel.GetIndex(dao.FulltextIndexName)
		if !ok {
			index := model.Index{
				Name:   field.Name,
				Fields: []string{field.Name},
			}
			indexes = append(indexes, index)
		}
	}
	// Delete unused indexes
	route := model.Route{
		Backend: backendname,
		Model:   bemodel.Name,
	}
	names, err := dao.GetStorage().GetIndexNames(route)
	if err != nil {
		return err
	}
	for _, idxName := range names {
		found := false
		for _, index := range indexes {
			if idxName == index.Name {
				found = true
				break
			}
		}
		if !found {
			err = dao.GetStorage().DeleteIndex(route, idxName)
		}
	}
	for _, index := range indexes {
		err := dao.GetStorage().UpdateIndex(route, index)
		if err != nil {
			return err
		}
	}
	return nil
}
