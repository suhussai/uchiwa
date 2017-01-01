package daemon

import (
	"errors"
	"fmt"

	"github.com/sensu/uchiwa/uchiwa/sensu"
	log "github.com/Sirupsen/logrus"
)

// FindDcFromInterface ...
func FindDcFromInterface(data interface{}, datacenters *[]sensu.Sensu) (*sensu.Sensu, map[string]interface{}, error) {
	m, ok := data.(map[string]interface{})
	if !ok {
		log.WithFields(log.Fields{
			"interface": data,
		}).Warn("Type assertion failed. Could not assert the given interface into a map.")
		return nil, nil, errors.New("Could not determine the datacenter.")
	}

	id := m["dc"].(string)
	if id == "" {
		log.WithFields(log.Fields{
			"interface": data,
		}).Warn("The received interface does not contain any datacenter information.")
		return nil, nil, errors.New("Could not determine the datacenter.")
	}

	for _, dc := range *datacenters {
		if dc.Name == id {
			return &dc, m, nil
		}
	}

	log.WithFields(log.Fields{
		"interface": data,
		"id": id,
	}).Warn("Could not find the datacenter.")
	return nil, nil, fmt.Errorf("Could not find the datacenter %s", id)
}

// setID sets the _id attribute on every element of the slice from the dc and name
func setID(elements []interface{}, separator string) {
	for _, e := range elements {
		element, ok := e.(map[string]interface{})
		if !ok {
			continue
		}

		dc, ok := element["dc"].(string)
		if !ok {
			continue
		}

		name, ok := element["name"].(string)
		if !ok {
			// Support silence entries
			name, ok = element["id"].(string)
			if !ok {
				// Support stashes
				name, ok = element["path"].(string)
				if !ok {
					continue
				}
			}
		}

		element["_id"] = fmt.Sprintf("%s%s%s", dc, separator, name)
	}
}

func setDc(v interface{}, dc string) {
	m, ok := v.(map[string]interface{})
	if !ok {
		log.WithFields(log.Fields{
			"interface": v,
		}).Warn("Could not assert interface.")
	} else {
		m["dc"] = dc
	}
}
