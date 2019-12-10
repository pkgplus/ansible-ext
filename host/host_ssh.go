package host

import (
	"errors"
	// "fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"ansible-ext/pkg/ssh"
	xssh "golang.org/x/crypto/ssh"
)

var (
	ERR_TRUST_ALREADY = errors.New("already trust")
	ERR_NOT_SUDOER    = errors.New("is not sudoer")

	MSG_SSH_FAILED        = "ssh登陆失败"
	MSG_SSH_NOAUTH        = "ssh用户名/密码错误"
	MSG_SSH_UNREACH       = "ssh端口不可达，请检查防火墙配置"
	MSG_SSH_TRUST_ALREADY = "已配置过ssh免密码登陆"
	MSG_SSH_NOT_SUDOER    = "ssh登陆用户非sudoer用户"

	SSH_ASYNC_MAXCOUNT = 30
	asyncLimitChan     chan bool

	REG_SUDO_PROMPT = regexp.MustCompile(`(?m)^\s*(?:\[sudo\] )?[Pp]assword(?:\s+for\s+\w+)?:\s*$`)
)

func init() {
	asyncLimitChan = make(chan bool, SSH_ASYNC_MAXCOUNT)
}

func (hl *HostLoginInfo) SSHCheck(sudo bool) error {
	client, err := hl.getSSHClient()
	if err != nil {
		return err
	}
	defer client.Close()

	if sudo {
		return checkSudo(client)
	}

	return nil
}

func (hl *HostLoginInfo) getSSHClient() (*xssh.Client, error) {
	return ssh.GetClientWithAuthType(hl.Host, hl.Port, hl.UserName, hl.Passwd, hl.authType)
}

func checkSudo(client *xssh.Client) error {
	wf, err := ssh.NewWaitforer(client)
	if err != nil {
		return err
	}
	defer wf.Close()

	wr := wf.Waitfor("sudo whoami", 5*time.Second, REG_SUDO_PROMPT, ssh.PROMPT_SHELL_REG)
	if wr.Error != nil {
		return wr.Error
	} else {
		if string(wr.Content) == "root" {
			return nil
		} else {
			return ERR_NOT_SUDOER
		}
	}
}

func (hl *HostLoginInfo) HostSSHTrust() error {
	// use public key
	client, err := ssh.GetClientWithAuthType(hl.Host, hl.Port, hl.UserName, hl.Passwd, ssh.LOGIN_USE_PUBKEY)
	if err == nil {
		defer client.Close()
		err = checkSudo(client)
		if err != nil {
			return err
		} else {
			return ERR_TRUST_ALREADY
		}
	}

	// use password
	client, err = ssh.GetClientWithAuthType(hl.Host, hl.Port, hl.UserName, hl.Passwd, ssh.LOGIN_USE_PASSWD)
	if err != nil {
		return err
	} else {
		defer client.Close()
		err = checkSudo(client)
		if err != nil {
			return err
		} else {
			return ssh.Trust(client)
		}
	}
}

func (hlb HostLoginInfoBatch) HostsSSHTrust() {
	var wg sync.WaitGroup
	for _, hl := range hlb {
		asyncLimitChan <- true
		wg.Add(1)
		go func(hl *HostLoginInfo) {
			defer func() {
				<-asyncLimitChan
				wg.Done()
			}()
			err := hl.HostSSHTrust()
			if err != nil {
				hl.parseSSHErr(err)
			}
		}(hl)
	}

	wg.Wait()
}

func (hlb HostLoginInfoBatch) SSHCheck(sudo bool) {
	// check ssh status
	var wg sync.WaitGroup
	sshRetChan := make(chan [2]interface{}, len(hlb))
	for index, hl := range hlb {
		wg.Add(1)
		go func(index int, hl *HostLoginInfo) {
			err := hl.SSHCheck(sudo)
			sshRetChan <- [2]interface{}{index, err}
		}(index, hl)
	}

	// check ssh status response
	go func() {
		for sshRet := range sshRetChan {
			i := sshRet[0].(int)
			if hlb[i].Result.Status == STATUS_OK {
				if sshRet[1] != nil {
					hlb[i].parseSSHErr(sshRet[1].(error))
				}
			}
			wg.Done()
		}
	}()
	wg.Wait()
	close(sshRetChan)
}

func (hl *HostLoginInfo) parseSSHErr(err error) {
	tr := hl.Result

	if err == nil {
		tr.Status = STATUS_OK
	} else {
		tr.Reason = err.Error()
		if err == ERR_TRUST_ALREADY {
			tr.Message = MSG_SSH_TRUST_ALREADY
			tr.Status = STATUS_OK
		} else {
			tr.Status = STATUS_FAILED
			if err == ERR_NOT_SUDOER {
				tr.Message = MSG_SSH_NOT_SUDOER
			} else if strings.Index(err.Error(), "unable to authenticate") >= 0 {
				tr.Message = MSG_SSH_NOAUTH
			} else if strings.Index(err.Error(), "getsockopt: connection refused") >= 0 ||
				strings.Index(err.Error(), "connection reset by peer") >= 0 {
				tr.Message = MSG_SSH_UNREACH
			} else {
				tr.Message = MSG_SSH_FAILED
			}
		}
	}
}
