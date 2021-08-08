package utils

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type BackupJson struct {
	Encoded  string `json:"encoded"`
	Encoding string `json:"encoding"`
	Address  string `json:"address"`
	Meta     string `json:"meta"`
}
type Node struct {
	Chain    string `yaml:"chain"`
	Sworker  string `yaml:"sworker"`
	Smanager string `yaml:"smanager"`
	Ipfs     string `yaml:"ipfs"`
}

type Identity struct {
	Backup string `yaml:"backup"`
}
type Chain struct {
	Name string `yaml:"name"`
}
type Conf struct {
	Node     *Node     `yaml:"node"`
	Identity *Identity `yaml:"identity"`
	Chain    *Chain    `yaml:"chain"`
}

func Json2Struct(jsonStr string) BackupJson {
	var backupJson BackupJson
	json.Unmarshal([]byte(jsonStr), &backupJson)
	return backupJson
}

func (c *Conf) GetConf(fileName string) *Conf {
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

//func main() {
//	var c Conf
//	c.getConf("/opt/crust/crust-node/config.yaml")
//	fmt.Println(c.Identity.Backup)
//}
