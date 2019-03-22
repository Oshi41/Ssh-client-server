package commands

import (
	"../keys"
	"golang.org/x/crypto/ssh"
	"log"
	"strings"
	"sync"
)

func GetClient(host, login, password string, isPass bool, sync *sync.WaitGroup) (*ssh.Client, error) {
	sync.Add(1)
	defer sync.Done()

	config, err := keys.GetSshConfig(isPass, login, password)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// дефолтный порт
	if !strings.Contains(host, ":") {
		host += ":22"
	}

	// DEBUG ONLY
	// time.Sleep(2 * time.Second)

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println("Successfully connected - ", host)
	return client, nil
}
