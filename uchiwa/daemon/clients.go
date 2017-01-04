package daemon

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/sensu/uchiwa/uchiwa/helpers"
	"github.com/sensu/uchiwa/uchiwa/structs"
	log "github.com/Sirupsen/logrus"
)

// buildClients constructs clients objects for frontend consumption
func (d *Daemon) buildClients() {
	for _, c := range d.Data.Clients {
		client, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		dc, ok := client["dc"].(string)
		if !ok {
			continue
		}

		name, ok := client["name"].(string)
		if !ok {
			continue
		}

		client["_id"] = fmt.Sprintf("%s/%s", dc, name)

		client = findClientEvents(client, &d.Data.Events)

		client["silenced"] = helpers.IsClientSilenced(name, dc, d.Data.Silenced)
	}
}

// findClientEvents searches for all events related to a particular client
// and set the status and output attributes of this client based on the events found
func findClientEvents(client map[string]interface{}, events *[]interface{}) map[string]interface{} {
	if len(*events) == 0 {
		client["status"] = 0
	} else {
		var criticals, warnings int
		var results []string
		for _, e := range *events {

			eventMap, ok := e.(map[string]interface{})
			if !ok {
				log.WithFields(log.Fields{
					"event": e,
				}).Warn("Could not convert the event to a map.")
				continue
			}

			// skip this event if the check attribute does not exist
			if eventMap["check"] == nil {
				continue
			}

			// skip this event if the datacenter isn't the right one
			if eventMap["dc"] == nil || eventMap["dc"] != client["dc"] {
				continue
			}

			clientMap, ok := eventMap["client"].(map[string]interface{})
			if !ok {
				log.WithFields(log.Fields{
					"client": eventMap["client"],
				}).Warn("Could not convert the event's client to a map.")
				continue
			}

			// skip this event if the client isn't the right one
			if clientMap["name"] == nil || clientMap["name"] != client["name"] {
				continue
			}

			// convert the check to a structure for easier handling
			var check structs.GenericCheck
			err := mapstructure.Decode(eventMap["check"], &check)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Warn("Could not convert the event's check to a generic check structure.")
				continue
			}

			if check.Status == 2 {
				criticals++
			} else if check.Status == 1 {
				warnings++
			}

			results = append(results, check.Output)
		}

		if len(results) == 0 {
			client["status"] = 0
		} else if criticals > 0 {
			client["status"] = 2
		} else if warnings > 0 {
			client["status"] = 1
		} else {
			client["status"] = 3
		}

		if len(results) == 1 {
			client["output"] = results[0]
		} else if len(results) > 1 {
			output := fmt.Sprintf("%s and %d more...", results[0], (len(results) - 1))
			client["output"] = output
		}
	}

	return client
}
