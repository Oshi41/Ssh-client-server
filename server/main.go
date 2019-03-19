package main

import (
	"github.com/Oshi41/ssh-keygen"
	"github.com/gliderlabs/ssh"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	privateKeyPath = "./auth/private.ssh"
	publicKeyPath  = "./auth/public.ssh"

	App = kingpin.New("Simple console ssh-server", "")

	Port = App.Arg("Port", "Port where server will be hosting").Default("2222").String()

	UseKey = App.Flag("key", "Tell server that we are using ssh key by path - "+privateKeyPath).Default("false").Bool()

	UsePass  = App.Flag("withPass", "Tell server that we are using password for clients").Default("false").Bool()
	Password = App.Arg("pass", "Password to enter to the server").Default("").String()

	server *ssh.Server
)

func main() {
	_, err := App.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	server = &ssh.Server{
		Addr:                   ":" + *Port,
		Handler:                sessionHandler,
	}

	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		log.Println("Starting generating key...")
		err = ssh_keygen.GenerateNew4096(privateKeyPath, publicKeyPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = server.SetOption(ssh.HostKeyFile(privateKeyPath))
	if err != nil {
		log.Println(err)
	}

	if *UseKey {
		err := server.SetOption(ssh.PublicKeyAuth(keyHandler))
		if err != nil {
			log.Println(err)
		}
	}

	if *UsePass {
		err := server.SetOption(ssh.PasswordAuth(passHandler))
		if err != nil {
			log.Println(err)
		}
	}

	log.Println("Starting ssh-server on port " + string(*Port))
	log.Fatal(server.ListenAndServe())
}

func keyHandler(_ ssh.Context, key ssh.PublicKey) bool {
	file, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalln(err)
		return false
	}

	serverKey, err := ssh.ParsePublicKey(file)
	if err != nil {
		log.Fatalln(err)
		return false
	}

	return ssh.KeysEqual(serverKey, key)
}

func sessionHandler(session ssh.Session) {
	io.WriteString(session, "Welcome to KsuSSH server, "+session.User()+"\n")
}

func passHandler(_ ssh.Context, password string) bool {
	log.Println("Server pass is " + *Password + "\n Requested is " + password)
	result := password == *Password
	return result
}
