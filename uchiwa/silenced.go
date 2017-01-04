package uchiwa

import log "github.com/Sirupsen/logrus"

type silence struct {
	ID              string `json:"id"`
	Dc              string `json:"dc"`
	Subscription    string `json:"subscription,omitempty"`
	Check           string `json:"check,omitempty"`
	Reason          string `json:"reason,omitempty"`
	Creator         string `json:"creator,omitempty"`
	Expire          int32  `json:"expire,omitempty"`
	ExpireOnResolve bool   `json:"expire_on_resolve,omitempty"`
}

// ClearSilenced send a POST request to the /stashes endpoint in order to create a stash
func (u *Uchiwa) ClearSilenced(data silence) error {
	api, err := getAPI(u.Datacenters, data.Dc)
	if err != nil {
		log.Warn(err)
		return err
	}

	_, err = api.ClearSilenced(data)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

// PostSilence send a POST request to the /stashes endpoint in order to create a stash
func (u *Uchiwa) PostSilence(data silence) error {
	api, err := getAPI(u.Datacenters, data.Dc)
	if err != nil {
		log.Warn(err)
		return err
	}

	_, err = api.Silence(data)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}
