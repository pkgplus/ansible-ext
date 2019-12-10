package consul

import (
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"time"
)

func ListKeys(key_prefix string) (keys []string, err error) {
	if consulClient == nil {
		err = errors.New("init consul client failed")
		return
	}

	kvs, _, err := consulClient.KV().List(key_prefix, &api.QueryOptions{Datacenter: dataCenter})
	if err != nil {
		return keys, err
	}
	keys = make([]string, len(kvs))
	for i, key := range kvs {
		keys[i] = key.Key
	}
	return
}

func GetKey(key string) (data []byte, err error) {
	if consulClient == nil {
		err = errors.New("init consul client failed")
		return
	}

	kv, _, err := consulClient.KV().Get(
		key,
		&api.QueryOptions{Datacenter: dataCenter},
	)
	if err != nil {
		return data, err
	}
	if kv == nil {
		return data, fmt.Errorf("can't found key %s", key)
	}

	return kv.Value, nil
}

func PutKey(key string, data []byte) (spendtime time.Duration, err error) {
	if consulClient == nil {
		err = errors.New("init consul client failed")
		return
	}

	ret, err := consulClient.KV().Put(
		&api.KVPair{Key: key, Value: data},
		&api.WriteOptions{Datacenter: dataCenter},
	)
	if err != nil {
		return spendtime, err
	}

	return ret.RequestTime, nil
}

func DelKey(key string) (spendtime time.Duration, err error) {
	if consulClient == nil {
		err = errors.New("init consul client failed")
		return
	}

	ret, err := consulClient.KV().Delete(
		key,
		&api.WriteOptions{Datacenter: dataCenter},
	)
	if err != nil {
		return spendtime, err
	}

	return ret.RequestTime, nil
}
