package database

import (
	"encoding/json"
	"io/ioutil"
)

type genesis struct {
	Balances map[Account]uint `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return genesis{}, err
	}
	var loadGenesis genesis
	err = json.Unmarshal(content, &loadGenesis)
	if err != nil {
		return genesis{}, err
	}
	return loadGenesis, nil
}
