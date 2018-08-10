package host

import (
	"os"
	"testing"
)

func TestHostSSHTrust(t *testing.T) {
	host := os.Getenv("SSH_HOST")
	username := os.Getenv("SSH_USERNAME")
	pwd := os.Getenv("SSH_PASSWD")
	if host != "" && username != "" && pwd != "" {
		hlb := HostLoginInfoBatch{
			&HostLoginInfo{
				SshInfo: &SshInfo{
					Host:     host,
					UserName: username,
					Passwd:   pwd,
					Port:     22,
				},
			},
		}
		hlb.Init()
		hlb.HostsSSHTrust()

		if hlb[0].Result.Status != STATUS_OK {
			t.Fatalf("%s:%s", hlb[0].Result.Message, hlb[0].Result.Reason)
		}
	}
}
