package uchiwa

import (
	"errors"
	"fmt"

	"github.com/sensu/uchiwa/uchiwa/sensu"
	log "github.com/Sirupsen/logrus"
)

func getAPI(datacenters *[]sensu.Sensu, name string) (*sensu.Sensu, error) {
	if len(*datacenters) == 1 {
		return &(*datacenters)[0], nil
	}

	if name == "" {
		return nil, errors.New("The datacenter name can't be empty")
	}

	for _, datacenter := range *datacenters {
		if datacenter.Name == name {
			return &datacenter, nil
		}
	}

	return nil, fmt.Errorf("Could not find the datacenter '%s'", name)
}

func findModel(id string, dc string, checks []interface{}) map[string]interface{} {
	for _, k := range checks {
		m, ok := k.(map[string]interface{})
		if !ok {
			log.WithFields(log.Fields{
				"check": k,
			}).Warn("Could not assert check interface.")
			continue
		}
		if m["name"] == id && m["dc"] == dc {
			return m
		}
	}
	return nil
}

// MergeStringSlices merges two slices of strings and remove duplicated values
func MergeStringSlices(a1, a2 []string) []string {
	if len(a1) == 0 {
		return a2
	} else if len(a2) == 0 {
		return a1
	}

	s := make([]string, len(a1), len(a1)+len(a2))
	copy(s, a1)

next:
	for _, x := range a2 {
		for _, y := range s {
			if x == y {
				continue next
			}
		}
		s = append(s, x)
	}
	return s
}

// SliceIntersection searches for values in both slices
// Returns true if there's at least one intersection
func SliceIntersection(a1, a2 []string) bool {
	if len(a1) == 0 || len(a2) == 0 {
		return false
	}

	for _, x := range a1 {
		for _, y := range a2 {
			if x == y {
				return true
			}
		}
	}

	return false
}
