package uchiwa

import log "github.com/Sirupsen/logrus"

// ResolveEvent sends a DELETE request in order to
// resolve an event for a given check on a given client
func (u *Uchiwa) ResolveEvent(check, client, dc string) error {
	api, err := getAPI(u.Datacenters, dc)
	if err != nil {
		log.Warn(err)
		return err
	}

	err = api.DeleteEvent(check, client)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}
