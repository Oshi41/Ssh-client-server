package commands

import (
	"golang.org/x/crypto/ssh"
	"strings"
)

// Создаем ssh-соединение
func AddConnection(host string, config *ssh.ClientConfig) (*ssh.Client, error) {

	// default port - 22
	if !strings.Contains(host, ";") {
		host += ":22"
	}

	// Используем всегда tcp протокол
	return ssh.Dial("tcp", host, config)
}
