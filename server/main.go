package main

import (
	"io"
	"github.com/gliderlabs/ssh"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
)

var (
	App = kingpin.New("Simple console ssh-server", "")

	Port  = App.Arg("Port", "Port where server will be hosting").Default("2222").String()

	UseKey = App.Flag("key", "Tell server that we are using ssh key")
	UseKeyPath = App.Arg("KeyPath", "Path to private ssh key").String()

	UsePass = App.Flag("pass", "Tell server that we are using login and password fro clients")
	UsePassPath = App.Arg("PassPath", "Path to stored passwords").String()

	privatePath = "./Keyes/private.ssh"
	knownhosts = "./Keyes/knownhosts.ssh"
)

func main(){
	server := &ssh.Server{
		Addr: ":" + *Port,
		Handler: sessionHandler,
		PublicKeyHandler: publicKeyHandler,
		PasswordHandler: passHandler,
	}

	// запускаем в отдельном потоке
	go log.Fatal(server.ListenAndServe())
}

func sessionHandler(session ssh.Session)  {
	io.WriteString(session, "Welcome to KsuSSH server, " + session.User())
}

func publicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {

	if UseKey.HasEnvarValue(){

	}

	return true
}

func passHandler(ctx ssh.Context, password string) bool {

	return true
}
