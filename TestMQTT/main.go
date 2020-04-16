/*
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const broker = "tcp://127.0.0.1:1883"

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

func main() {
	registerMQTTTopic("wohnzimmer_reciver", "data/temperatur/wohnzimmer")
	registerMQTTTopic("kueche_reciver", "data/temperatur/kueche")
	time.Sleep(60 * time.Second)
}
