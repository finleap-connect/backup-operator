package testutil

import (
	consulApi "github.com/hashicorp/consul/api"
	"github.com/ory/dockertest/v3"
)

func WaitForConsul(pool *dockertest.Pool, endpoint string) error {
	return pool.Retry(func() error {
		consulConf := consulApi.DefaultConfig()
		consulConf.Address = endpoint
		client, err := consulApi.NewClient(consulConf)
		if err != nil {
			return err
		}
		_, err = client.Status().Leader()
		return err
	})
}
