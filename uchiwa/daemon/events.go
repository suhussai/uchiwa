package daemon

import (
	"fmt"

	"github.com/sensu/uchiwa/uchiwa/helpers"
	log "github.com/Sirupsen/logrus"
)

// BuildEvents constructs events objects for frontend consumption
func (d *Daemon) buildEvents() {
	for _, e := range d.Data.Events {
		m := e.(map[string]interface{})

		// get client name
		clientMap, ok := m["client"].(map[string]interface{})
		if !ok {
			log.WithFields(log.Fields{
				"client": m["client"],
			}).Warn("Could not assert event's client interface.")
			continue
		}

		client, ok := clientMap["name"].(string)
		if !ok {
			log.WithFields(log.Fields{
				"name": clientMap["name"],
			}).Warn("Could not assert event's client name from client map.")
			continue
		}

		// get check name
		checkMap, ok := m["check"].(map[string]interface{})
		if !ok {
			log.WithFields(log.Fields{
				"check": m["check"],
			}).Warn("Could not assert event's check from client map.")
			continue
		}

		check, ok := checkMap["name"].(string)
		if !ok {
			log.WithFields(log.Fields{
				"name": checkMap["name"],
			}).Warn("Could not assert event's check name from check map.")
			continue
		}

		// get dc name
		dc, ok := m["dc"].(string)
		if !ok {
			log.WithFields(log.Fields{
				"name": m["dc"],
			}).Warn("Could not assert event's datacenter name from check map.")
			continue
		}

		// Set the event unique ID
		m["_id"] = fmt.Sprintf("%s/%s/%s", dc, client, check)

		// Determine if the client is silenced
		m["client"].(map[string]interface{})["silenced"] = helpers.IsClientSilenced(client, dc, d.Data.Silenced)

		// Determine if the check is silenced.
		// See https://github.com/sensu/uchiwa/issues/602
		m["silenced"], m["silenced_by"] = helpers.IsCheckSilenced(checkMap, client, dc, d.Data.Silenced)
	}
}
