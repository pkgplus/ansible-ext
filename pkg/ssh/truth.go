package ssh

import (
	"fmt"
	"regexp"

	xssh "golang.org/x/crypto/ssh"
)

var (
	REGEXP_SUDO_PWD *regexp.Regexp
)

func init() {
	REGEXP_SUDO_PWD = regexp.MustCompile(`password\s*(?:for [\w\-]+)?:\s*`)
}

func Trust(client *xssh.Client) error {

	// mkdir
	cmd := "mkdir ./.ssh"
	output, err := Cmd(client, cmd)
	if err != nil {
		switch err.(type) {
		case *xssh.ExitError:
		default:
			return fmt.Errorf("ExitError:: %s", string(output))
		}
	} else {
		cmd = "chmod 700 ./.ssh"
		Cmd(client, cmd)
	}

	// append
	cmd = fmt.Sprintf(`echo "%s" >> ./.ssh/authorized_keys`, string(pubkey))
	output, err = Cmd(client, cmd)
	if err != nil {
		return fmt.Errorf("add publickey failed(%v):: %s", err, string(output))
	}

	//uniq the authorized_keys
	cmd = "sort -u ./.ssh/authorized_keys > ./.ssh/authorized_keys.bak"
	output, err = Cmd(client, cmd)
	if err == nil {
		cmd = "cp -f ./.ssh/authorized_keys.bak ./.ssh/authorized_keys"
		Cmd(client, cmd)
	}
	cmd = "chmod 600 ./.ssh/authorized_keys"
	Cmd(client, cmd)

	// WaitFor(client, "sudo pwd", REGEXP_SUDO_PWD)

	return nil
}
