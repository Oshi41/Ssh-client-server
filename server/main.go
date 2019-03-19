package main

import (
	"github.com/Oshi41/ssh-keygen"
	"github.com/gliderlabs/ssh"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
	"net"
	"io/ioutil"
	"os/exec"
	"fmt"
	"github.com/kr/pty"
	"io"
	"syscall"
	"unsafe"
	"runtime"
	"os/user"
)

var (
	privateKeyPath = "./auth/private.ssh"
	publicKeyPath  = "./auth/public.ssh"

	App = kingpin.New("Simple console ssh-server", "")

	Port = App.Arg("Port", "Port where server will be hosting").Default("2222").String()

	UseKey = App.Flag("key", "Tell server that we are using ssh key by path - "+privateKeyPath).Default("false").Bool()

	UsePass = App.Flag("withPass", "Tell server that we are using password for clients").Default("true").Bool()
	SshPassword = App.Arg("pass", "Password for ssh-session authentification").Default("1qazXSW@").String()

	server *ssh.Server
)

func main() {
	_, err := App.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	server = &ssh.Server{
		Addr:         ":" + *Port,
		ConnCallback: greetingHandler,
	}

	if runtime.GOOS == "linux" {
		server.Handler = linuxSessionHandler
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

	log.Println("Starting ssh-server on port ", *Port)
	log.Fatal(server.ListenAndServe())
}

func greetingHandler(conn net.Conn) net.Conn {

	conn.Write([] byte("Welcome to KsuSSH server!\n"))

	return conn
}

func linuxSessionHandler(s ssh.Session) {
	cmd := exec.Command("bash")
	ptyReq, winCh, isPty := s.Pty()
	if isPty {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
		f, err := pty.Start(cmd)
		if err != nil {
			panic(err)
		}

		go func() {
			// Выще хз что это, работает - не трож!!!
			for win := range winCh {
				syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
					uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(win.Height), uint16(win.Width), 0, 0})))
			}
		}()

		go func() {
			io.Copy(f, s) // stdin
		}()

		io.Copy(s, f) // stdout
	}
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

func passHandler(context ssh.Context, password string) bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Println(err)
		return false
	}

	if context.User() != currentUser.Username {
		log.Println("Requested - ",context.User() , " but should - ", currentUser.Username)
		return false
	}

	if *SshPassword != password{
		log.Println("Wrong password")
		return false
	}

	return true
}
