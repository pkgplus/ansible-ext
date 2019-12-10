package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	xssh "golang.org/x/crypto/ssh"
)

var (
	LOGIN_USE_PASSWD = uint8(1)
	LOGIN_USE_PUBKEY = uint8(2)
	LOGIN_USE_ANY    = uint8(3)

	pubkey []byte
	signer xssh.Signer

	Timeout = 5 * time.Second
)

func init() {
	// ssh private key
	key, err := ioutil.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa")
	if err != nil {
		panic("unable to read private key: " + err.Error())
	}
	signer, err = xssh.ParsePrivateKey(key)
	if err != nil {
		panic("unable to parse private key: " + err.Error())
	}

	// ssh public key
	pubkey, err = ioutil.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa.pub")
	if err != nil {
		panic("unable to parse public key: " + err.Error())
	}
	pubkey = bytes.TrimSuffix(pubkey, []byte{'\n'})
}

func GetClient(host string, port int32, username, passwd string) (*xssh.Client, error) {
	return getClientWithAuthType(host, port, username, passwd, LOGIN_USE_PASSWD)
}

func GetClientWithAuthType(host string, port int32, username, passwd string, at uint8) (*xssh.Client, error) {
	return getClientWithAuthType(host, port, username, passwd, at)
}

func getClientWithAuthType(host string, port int32, username, passwd string, at uint8) (*xssh.Client, error) {
	config := &xssh.ClientConfig{
		User:    username,
		Timeout: Timeout,
		HostKeyCallback: func(hostname string, remote net.Addr, key xssh.PublicKey) error {
			return nil
		},
	}

	if at&LOGIN_USE_PASSWD > 0 {
		config.Auth = append(config.Auth, xssh.Password(passwd))
	}
	if at&LOGIN_USE_PUBKEY > 0 {
		config.Auth = append(config.Auth, xssh.PublicKeys(signer))
	}

	target := fmt.Sprintf("%s:%d", host, port)
	return xssh.Dial("tcp", target, config)
}

func Cmd(client *xssh.Client, cmd string) ([]byte, error) {
	sess, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()
	return sess.CombinedOutput(cmd)
}
