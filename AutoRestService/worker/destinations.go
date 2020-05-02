package worker

import (
	"errors"
	"fmt"

	"github.com/willie68/AutoRestIoT/model"
)

//DestinationProcessor fpr every destination there must be a processor to do the work
type DestinationProcessor interface {
	//Initialise this procssor
	Initialise(backend string, destination model.Destination) error
	//Destroy this processor
	Destroy(backend string, destination model.Destination) error
	//Store do the right storage
	Store(data model.JSONMap) (string, error)
}

//NullDestinationProcessor does nothing
type NullDestinationProcessor struct {
}

//Initialise do nothing on initialise
func (n *NullDestinationProcessor) Initialise(backend string, destination model.Destination) error {
	return nil
}

//Destroy do nothing on initialise
func (n *NullDestinationProcessor) Destroy(backend string, destination model.Destination) error {
	return nil
}

//Store do nothing on store
func (n *NullDestinationProcessor) Store(data model.JSONMap) (string, error) {
	return "noId", nil
}

//ErrDestinationNotFound the destination was not found in this system
var ErrDestinationNotFound = errors.New("Missing destination")

//DestinationList list type
type DestinationList struct {
	destinations map[string]model.Destination
	processors   map[string]DestinationProcessor
}

func GetNewDestinationProcessor(backend string, destination model.Destination) (DestinationProcessor, error) {
	switch destination.Type {
	case "mqtt":
		return CreateMQTTDestinationProcessor(backend, destination)
	case "null":
		return &NullDestinationProcessor{}, nil
	default:
		return &NullDestinationProcessor{}, nil
	}
}

//Destinations List off all registered destinations
var Destinations = DestinationList{
	destinations: make(map[string]model.Destination),
}

//Register registering a new destination under the right name
func (d *DestinationList) Register(backendName string, destination model.Destination) error {
	destinationNsName := GetDestinationNsName(backendName, destination.Name)
	d.destinations[destinationNsName] = destination
	return nil
}

//Deregister deregistering a new destination with a name
func (d *DestinationList) Deregister(backendName string, destination model.Destination) error {
	destinationNsName := GetDestinationNsName(backendName, destination.Name)
	// getting the processor for this
	processor, ok := d.processors[destinationNsName]
	if ok {
		err := processor.Destroy(backendName, destination)
		if err != nil {
			return err
		}
		delete(d.processors, destinationNsName)
	}
	// removing the destination from the list
	delete(d.destinations, destinationNsName)
	return nil
}

//Store storing a message into the desired destination
func (d *DestinationList) Store(backendName string, destinationName string, data model.JSONMap) error {
	destinationNsName := GetDestinationNsName(backendName, destinationName)
	destination, ok := d.destinations[destinationNsName]
	if !ok {
		return ErrDestinationNotFound
	}
	var processor DestinationProcessor
	processor, ok = d.processors[destinationName]
	if !ok {
		var err error
		processor, err = GetNewDestinationProcessor(backendName, destination)
		if err != nil {
			return err
		}
	}
	_, err := processor.Store(data)
	if err != nil {
		return err
	}
	//log.Infof("store object in destination %s as %s", destination, id)
	return nil
}

//GetDestinationNsName getting the unique name of a backend destination
func GetDestinationNsName(backendName, destinationName string) string {
	return fmt.Sprintf("%s.%s", backendName, destinationName)
}
