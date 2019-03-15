package keys

import (
	"../reader"
	"fmt"
	"github.com/Oshi41/ssh-keygen"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

var (
	// В этой папке будут лежать ключи
	baseFolder = "./Keyes"

	// Путь к приватной части ключа
	privateKeyFile = baseFolder + "/private.ssh"
	// Путь к публичной части ключа
	publicKeyFile = baseFolder + "/public.ssh"
	// Путь к истории всех посещаемых хостов
	knownHosts = baseFolder + "/known_hosts.ssh"

	// Флаг лдя разрешений при создании файлов/папок
	perm os.FileMode = 0700
)

// Функция инициализации пакета.
// В ней проверяяю наличие папки + наличие файла known hosts
func init() {
	if _, err := os.Open(baseFolder); os.IsNotExist(err) {
		os.Mkdir(baseFolder, perm)
	}

	if _, err := os.Open(knownHosts); os.IsNotExist(err) {
		os.Create(knownHosts)
	}
}

func GenerateNew() error {
	return ssh_keygen.GenerateNew4096(privateKeyFile, publicKeyFile)
}

// Содаю конфиг клиента для аутентификации из ssh ключа.
func GetSshConfig(name string) (*ssh.ClientConfig, error) {
	// Хитровыдуманная проверка, что пути не существует,
	// четсно стырено
	// https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	if _, err := os.Stat(publicKeyFile); os.IsNotExist(err) {
		err = GenerateNew()
		if err != nil {
			return nil, err
		}
	}

	file, err := os.Open(privateKeyFile)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            name,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: callBack,
	}

	return config, nil
}

// Содаю конфиг клиента для аутентификации из логина и пароля.
func GetPasswordConfig(login string, pass string) (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            login,
		Auth:            []ssh.AuthMethod{ssh.Password(pass)},
		HostKeyCallback: callBack,
	}

	return config, nil
}

// функция проверки наличия отпечатка хоста. Если его нет, запрашивает разрещшения пользователя,
// добавляет хост и пропускает
func callBack(hostname string, remote net.Addr, key ssh.PublicKey) error {
	result, err := knownhosts.New(knownHosts)
	if err != nil {
		log.Fatal(err)
	}

	err = result(hostname, remote, key)
	if err != nil {

		// Asking user
		fmt.Println("You want to add server fingerprint (y/n)?")
		if !strings.Contains(reader.Read(), "y") {
			return err
		}

		address := make([]string, 1)
		address = append(address, remote.String())

		file, err := os.OpenFile(knownHosts, os.O_APPEND|os.O_WRONLY, perm)
		if err != nil {
			log.Fatal(err)
		}

		defer file.Close()

		line := knownhosts.Line(address, key) + "\n"
		if _, err = file.WriteString(line); err != nil {
			log.Fatal(err)
		}

	}
	// Ещё раз перечитываю хосты
	result, err = knownhosts.New(knownHosts)
	if err != nil {
		log.Fatal(err)
	}

	return result(hostname, remote, key)
}
