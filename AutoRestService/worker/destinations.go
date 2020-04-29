package worker

import (
	"errors"
	"fmt"

	"github.com/willie68/AutoRestIoT/model"
)

var ErrDestinationNotFound = errors.New("Missing destination")

//DestinationList list type
type DestinationList struct {
	destinations map[string]model.Destination
}

//Destinations List off all registered destinations
var Destinations = DestinationList{
	destinations: make(map[string]model.Destination),
}

//AddDestination registering a new destination under the right name
func (d *DestinationList) Add(backendName string, destination model.Destination) error {
	destinationNsName := GetDestinationNsName(backendName, destination.Name)
	d.destinations[destinationNsName] = destination
	return nil
}

//Store storing a message into the desired destination
func (d *DestinationList) Store(backendName string, destinationName string, data model.JSONMap) error {
	destinationNsName := GetDestinationNsName(backendName, destinationName)
	destination, ok := d.destinations[destinationNsName]
	if !ok {
		return ErrDestinationNotFound
	}
	log.Infof("store object in destination: %s", destination)
	return nil
}

//GetDestinationNsName getting the unique name of a backend destination
func GetDestinationNsName(backendName, destinationName string) string {
	return fmt.Sprintf("%s.%s", backendName, destinationName)
}
