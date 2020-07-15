package parsers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

/* BETTER CI CONFIG STRUCTURE

RUN:
	-
BUILD-IMAGE:
	- IMAGE_NAME1 : absfdsf
	  FILE : dockerfile
	  CONTEXT : .
	  PUSH : true
	- IMAGE_NAME2 :
		FILE : dockerfile
		CONTEXT : .
		PUSH : true
DOCKER : ci-docker-compose.yml
K8S : ci-k8s-manifest.yml
*/

type Image struct {
	NAME    string `json:"name"`
	FILE    string `json:"file"`
	CONTEXT string `json:"context"`
	PUSH    bool   `json:"push"`
}
type CIConfig struct {
	CMD    []string `json:"cmd"`
	BUILD  []Image  `json:"build"`
	DOCKER string   `json:"docker-compose"`
	K8S    string   `json:"k8s-manifest"`
}

func ConfigParser(dir string) (CIConfig, error) {
	configFile, err := os.Open(dir + "/.betterci/config.json")
	if err != nil {
		return CIConfig{}, err
	}
	defer configFile.Close()
	byteValue, _ := ioutil.ReadAll(configFile)
	var configData CIConfig
	err = json.Unmarshal([]byte(byteValue), &configData)
	if err != nil {
		return CIConfig{}, err
	}
	if configData.DOCKER != "" && configData.K8S != "" {
		return CIConfig{}, errors.New("ERROR : CANNOT HAVE BOTH DOCKER & K8S CONFIG")
	}
	//fmt.Println(configData)
	return configData, nil
}
