package worker

import (
	orglog "log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
	"gopkg.in/square/go-jose.v2/json"
)

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
		err := createDatasource(datasource, backend.Backendname)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	return nil
}

func createDatasource(datasource model.DataSource, backendname string) error {
	switch datasource.Type {
	case "mqtt":
		err := registerMQTTTopic(datasource.Name, backendname, datasource.Destination, datasource.Config.(model.DataSourceConfigMQTT))
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
	Backend string
	Model   string
	Topic   string
	Payload string
}

var mqttClients = make([]MqttDatasource, 0)

func f(datasource MqttDatasource, client mqtt.Client, msg mqtt.Message) {
	log.Infof("MODEL: %s.%s TOPIC: %s  MSG: %s", datasource.Backend, datasource.Model, msg.Topic(), msg.Payload())
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

func registerMQTTTopic(clientID string, backendname string, destinationmodel string, config model.DataSourceConfigMQTT) error {
	//	mqtt.DEBUG = orglog.New(os.Stdout, "DEBUG", 0)
	mqtt.ERROR = orglog.New(os.Stdout, "ERROR", 0)
	opts := mqtt.NewClientOptions().AddBroker(config.Broker).SetClientID(clientID)
	opts.SetKeepAlive(2 * time.Second)
	//opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	token := c.Connect()
	token.Wait()
	err := token.Error()
	if err != nil {
		return err
	}
	datasource := MqttDatasource{
		Client:  c,
		Backend: backendname,
		Model:   destinationmodel,
		Topic:   config.Topic,
		Payload: config.Payload,
	}
	mqttClients = append(mqttClients, datasource)

	token = c.Subscribe(config.Topic, 0, func(c mqtt.Client, m mqtt.Message) {
		f(datasource, c, m)
	})
	token.Wait()
	err = token.Error()
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
