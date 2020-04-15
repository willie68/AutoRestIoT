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
	"math/rand"
	"os"
	"time"

	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	flag "github.com/spf13/pflag"
)

var mqttbroker string
var clientid string
var topic string

func init() {
	// variables for parameter override
	flag.StringVarP(&mqttbroker, "broker", "b", "", "this is the address of the mqtt broker")
	flag.StringVarP(&clientid, "clientid", "c", "", "this is the client id to use")
	flag.StringVarP(&topic, "topic", "t", "", "this is the topic to use")
}

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func main() {
	flag.Parse()
	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(mqttbroker).SetClientID(clientid)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	/*
		if token := c.Subscribe("go-mqtt/sample", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	*/

	rand.Seed(time.Now().UnixNano())
	min := 150
	max := 300

	for i := 0; i < 10000; i++ {
		payload := make(map[string]interface{})
		n := min + rand.Intn(max-min+1)

		payload["temperatur"] = float32(n) / 10.0
		payload["number"] = i
		payload["source"] = topic
		payloadbytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		fmt.Printf("sending %d payload: %s\n", i, string(payloadbytes))
		token := c.Publish(topic, 0, false, payloadbytes)
		token.Wait()
		time.Sleep(500 * time.Millisecond)
	}

	/*
		if token := c.Unsubscribe("go-mqtt/sample"); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	*/
	c.Disconnect(250)

	time.Sleep(1 * time.Second)
}
