package sensu

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

// These are the methods directly used by the public methods of the sensu
// package in order to handle the failover and load balancing between the APIs of a datacenter

func (s *Sensu) delete(endpoint string) error {
	apis := shuffle(s.APIs)

	var err error
	for i := 0; i < len(apis); i++ {
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
		}).Debug("DELETE operation.")
		err = apis[i].delete(endpoint)
		if err == nil {
			return err
		}
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
			"error": err,
		}).Warn("DELETE operation returned error.")
	}

	return err
}

func (s *Sensu) getBytes(endpoint string) ([]byte, *http.Response, error) {
	apis := shuffle(s.APIs)

	for i := 0; i < len(apis); i++ {
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
		}).Debug("GET operation.")
		bytes, res, err := apis[i].getBytes(endpoint)
		if err == nil {
			return bytes, res, err
		}
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
			"error": err,
		}).Warn("GET operation returned error.")
	}

	return nil, nil, errors.New("")
}

func (s *Sensu) getSlice(endpoint string, limit int) ([]interface{}, error) {
	apis := shuffle(s.APIs)

	for i := 0; i < len(apis); i++ {
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
		}).Debug("GET operation.")
		slice, err := apis[i].getSlice(endpoint, limit)
		if err == nil {
			return slice, err
		}
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
			"error": err,
		}).Warn("GET operation returned error.")
	}

	return nil, errors.New("")
}

func (s *Sensu) getMap(endpoint string) (map[string]interface{}, error) {
	apis := shuffle(s.APIs)

	for i := 0; i < len(apis); i++ {
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
		}).Debug("GET operation.")
		m, err := apis[i].getMap(endpoint)
		if err == nil {
			return m, err
		}
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
			"error": err,
		}).Warn("GET operation returned error.")
	}

	return nil, errors.New("")
}

func (s *Sensu) postPayload(endpoint string, payload string) (map[string]interface{}, error) {
	apis := shuffle(s.APIs)

	for i := 0; i < len(apis); i++ {
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
		}).Debug("POST operation.")
		m, err := apis[i].postPayload(endpoint, payload)
		if err == nil {
			return m, err
		}
		log.WithFields(log.Fields{
			"url": s.APIs[i].URL,
			"endpoint": endpoint,
			"error": err,
		}).Warn("POST operation returned error.")
	}

	return nil, errors.New("")
}

// shuffle the provided []API
func shuffle(apis []API) []API {
	rand.Seed(time.Now().UnixNano())
	for i := range apis {
		j := rand.Intn(i + 1)
		apis[i], apis[j] = apis[j], apis[i]
	}
	return apis
}
