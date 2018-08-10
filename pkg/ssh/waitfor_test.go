package ssh

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

var (
	regexp_sudo_pwd = regexp.MustCompile(`(?m)^\s*(?:\[sudo\] )?[Pp]assword(?:\s+for\s+\w+)?:\s*$`)
)

func TestWaitfor(t *testing.T) {
	client, err := GetClientWithAuthType("10.138.16.192", 22, "xuebing", "xuebing", LOGIN_USE_PASSWD)
	if err != nil {
		t.Fatalf("get ssh client failed:%v", err)
	}

	wf, err := NewWaitforer(client)
	if err != nil {
		t.Fatalf("new waitforer failed:%v", err)
	}

	wr := wf.Waitfor("sudo whoami", 5*time.Second, regexp_sudo_pwd, PROMPT_SHELL_REG)
	if wr.Error != nil {
		t.Fatalf("waitfor failed:%v", wr.Error)
	} else {
		fmt.Printf("CONTENT: '%s'\n", wr.Content)
		fmt.Printf("MATCHED: '%s'\n", wr.Matched)
	}

}
