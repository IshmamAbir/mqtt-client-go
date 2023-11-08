package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_Util "main/util"

	"main/model"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var db *gorm.DB
var clientOption *mqtt.ClientOptions
var brokerConf map[string]string
var err error

// steps to mqtt connection setup
// step 1: config the db. i used gorm and postgres
// step 2: set a broker and connect to the client
// step 3: handle the incoming msg and modify the data according your need in the 'setDefaultPublishHandler' method that would be published to the client. save data to db.
// step 4: connect the client to broker
// step 5: subscribe to your desired topic to get the message
// step 6: unsubscribe and disconnect if needed

func main() {
	fmt.Println("Running")
	fmt.Println("------------------------")

	dbConfig()

	brokerConfig()

	client := mqtt.NewClient(clientOption)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Panic(token.Error())
	}

	// for i := 0; i < 10; i++ {
	// 	if token := client.Subscribe(brokerConf["topic"], 0, nil); token.Wait() && token.Error() == nil {
	// 		log.Println("subscribed")
	// 		break
	// 	} else {
	// 		log.Println("Retrying")
	// 	}
	// 	time.Sleep(time.Second)
	// }

	for i := 0; i < 10; i++ {
		if token := client.Subscribe(brokerConf["topic"], 0, func(client mqtt.Client, msg mqtt.Message) {
			sensorData := model.SensorData{}
			sensorData.SensorId = i
			sensorData.Time = time.Now()
			sensorData.SensorValue = msg.Payload()
			db.Save(&sensorData)
			fmt.Printf("Received message on topic: %s\nMessage: %s\n", msg.Topic(), string(msg.Payload()))
		}); token.Wait() && token.Error() != nil {
			fmt.Fprintf(os.Stderr, "Error subscribing to topic: %s\n", token.Error())
			os.Exit(1)
		}
		time.Sleep(time.Second)
	}

	defer func() {
		if token := client.Unsubscribe(brokerConf["topic"]); token.Wait() && token.Error() != nil {
			log.Panic(token.Error())
		}
		client.Disconnect(500)

	}()

}

// setup default message handler
var f mqtt.MessageHandler = func(c mqtt.Client, m mqtt.Message) {
	go func(c mqtt.Client, m mqtt.Message) {
		sensorData := model.SensorData{}
		packet := model.Packet{}
		json.Unmarshal(m.Payload(), &packet)
		if !packet.Enc {
			fmt.Println(packet.Buffer)
			b, _ := json.Marshal(packet.Buffer)
			json.Unmarshal(b, &sensorData)
			sensorData.Time = time.Now()
			publishPacket := model.Packet{}
			publishPacket.Enc = false
			publishPacket.Buffer = sensorData
			packetBytrArr, _ := json.Marshal(publishPacket)
			jsonPayload := string(packetBytrArr)

			token := c.Publish("plot-data", 0, false, jsonPayload)
			token.Wait()
		} else {

		}

	}(c, m)
}

// set mqtt broker
func brokerConfig() {
	brokerConf, err = _Util.ReadConfig("./configs/broker.config")
	if err != nil {
		log.Panic(err)
	}
	clientOption = mqtt.NewClientOptions().AddBroker("tcp://" + brokerConf["host"] + ":" + brokerConf["port"])
	clientOption.SetUsername(brokerConf["username"])
	clientOption.SetPassword(brokerConf["password"])
	clientOption.SetClientID(brokerConf["client"])

	// it handles the incoming message, modify it according to users need, and save data to db
	clientOption.SetDefaultPublishHandler(f) //publishing or distributing content or data to its intended audience

	clientOption.AutoReconnect = true
	clientOption.SetAutoReconnect(true)

}

// This is to checking the connection is ok with the database and then connect with the database
func dbConfig() {
	// Getting db config values from db.config files by processing the file through readConfig method in config_util.go
	dbConfig, err := _Util.ReadConfig("./configs/db.config")
	if err != nil {
		log.Panic(err)
	}

	dbString := "host= " + dbConfig["host"] + " user= " + dbConfig["username"] + " password= " + dbConfig["password"] + " dbname= " + dbConfig["database"] + " port= " + dbConfig["port"] + " sslmode = disable"

	db, err = gorm.Open(postgres.Open(dbString), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("Connection Successful")
	db.AutoMigrate(&model.SensorData{})
}
