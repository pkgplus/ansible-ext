package ansible

import (
	"bytes"
	"fmt"
	"strconv"

	hexec "github.com/xuebing1110/ansible-ext/pkg/exec"
)

var (
	STATUS_SUCC    = "SUCCESS"
	STATUS_UNREACH = "UNREACHABLE!"
	STATUS_FAILED  = "FAILED"
)

type HostAnsibleResult struct {
	Host       string `json:"host"`
	Status     string `json:"status"`
	ReturnCode int    `json:"rc"`
	Content    string `json:"content"`
}

func Exec(host string, args ...string) (*hexec.Command, error) {
	a_args := make([]string, 0, len(args)+1)
	a_args = append(a_args, host)
	a_args = append(a_args, args...)

	fmt.Printf("ARGS: %+v\n", a_args)
	hc := hexec.NewCommand("ansible", a_args...)

	err := hc.RunStreaming(&AnsibleParser{})
	if err != nil {
		return hc, err
	}

	return hc, nil
}

type AnsibleParser struct {
	har *HostAnsibleResult
}

func (ap *AnsibleParser) Execute(line []byte, outChan chan interface{}, eof bool) {
	// fmt.Printf("----->\n%s\n%v\n<------\n", string(line), line)

	delim := []byte{'>', '>', '\n'}
	if bytes.HasSuffix(line, delim) {

		last_i := len(line) - len(delim)
		head_content := line[:last_i]

		if ap.har != nil {
			outChan <- ap.har
		}

		ap.har = new(HostAnsibleResult)
		ip_status_rc := bytes.Split(head_content, []byte{' ', '|', ' '})
		if len(ip_status_rc) == 3 {
			ap.har.ReturnCode, _ = strconv.Atoi(string(ip_status_rc[2]))
		}

		ap.har.Host = string(ip_status_rc[0])
		ap.har.Status = string(ip_status_rc[1])
	} else if eof {
		if ap.har != nil {
			outChan <- ap.har
		}
		ap.har = nil
	} else if ap.har != nil && ap.har.Host != "" {
		ap.har.Content = ap.har.Content + string(line)
	} else {
	}
}

func ExecToArray(host string, r []*HostAnsibleResult) error {
	return nil
}
