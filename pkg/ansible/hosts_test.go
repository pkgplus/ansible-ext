package ansible

import (
	"testing"
)

func TestAddHosts(t *testing.T) {
	HostsFile = "./hosts"
	HostsFileTmp = "./.hosts.tmp"

	err := AddHosts(
		"TEST",
		[]AnsibleHost{
			AnsibleHost{"127.0.0.1", "xuebing"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}
