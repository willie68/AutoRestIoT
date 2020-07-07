package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const broker = "tcp://192.168.178.14:1883"
const clientID = "tasmotaFakeClient"

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s  MSG: %s\n", msg.Topic(), msg.Payload())
}

func registerMQTTTopic(clientID string, topic string) error {
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)
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
	token = c.Subscribe(topic, 0, f)
	token.Wait()
	err = token.Error()
	if err != nil {
		return err
	}
	log.Printf("registering client %s  topic %s on %s", clientID, topic, broker)
	return nil
}

func sendMessage(topic string, message string) error {
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID).SetUsername("temp").SetPassword("temp")
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

	token = c.Publish(topic, 0, false, message)
	token.Wait()
	err = token.Error()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	data, err := ioutil.ReadFile("long_template.json")
	if err != nil {
		fmt.Printf("err: %v", err)
		panic(1)
	}
	err = sendMessage("tele/fake/SENSOR", string(data))
	if err != nil {
		fmt.Printf("err: %v", err)
		panic(1)
	}
}
