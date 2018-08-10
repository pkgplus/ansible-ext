package ansible

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

var (
	mutex        sync.Mutex
	HostsFile    string = "/etc/ansible/hosts"
	HostsFileTmp string = "/etc/ansible/.hosts.tmp"
)

type AnsibleHost struct {
	Host     string
	UserName string
}

func AddHosts(label string, hosts []AnsibleHost) error {
	mutex.Lock()
	defer mutex.Unlock()

	// read file
	AnsibleHosts := make(map[string][][]string)
	content, err := ioutil.ReadFile(HostsFile)
	if err != nil {
		return err
	}

	// load file content
	var labels = make([]string, 0)
	var label_cur string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if len(line) > 0 {
			if line[0] == '[' && line[len(line)-1] == ']' {
				label_cur = string(line[1:(len(line) - 1)])
				AnsibleHosts[label_cur] = make([][]string, 0)
				labels = append(labels, label_cur)
			} else if label_cur != "" {
				cols := strings.Split(line, " ")
				if len(cols) >= 1 {
					AnsibleHosts[label_cur] = append(AnsibleHosts[label_cur], cols)
				} else {

				}
			} else {

			}
		} else {

		}
	}

	// add host
	hostlines, found := AnsibleHosts[label]
	if !found {
		AnsibleHosts[label] = make([][]string, 0, len(hosts))
		for _, host := range hosts {
			var new_hostline []string
			if host.UserName == "root" || host.UserName == "" {
				new_hostline = []string{host.Host}
			} else {
				new_hostline = []string{host.Host, "ansible_ssh_user=" + host.UserName}
			}

			AnsibleHosts[label] = append(AnsibleHosts[label], new_hostline)
		}

		labels = append(labels, label)
	} else {
		for _, host := range hosts {
			var exist bool = false
			for _, hostline := range hostlines {
				if hostline[0] == host.Host {
					exist = true
					break
				}
			}

			if !exist {
				var new_hostline []string
				if host.UserName == "root" || host.UserName == "" {
					new_hostline = []string{host.Host}
				} else {
					new_hostline = []string{host.Host, "ansible_ssh_user=" + host.UserName}
				}

				AnsibleHosts[label] = append(AnsibleHosts[label], new_hostline)
			}
		}
	}

	// write temp file
	var new_content = make([]string, 0)
	for _, label_tmp := range labels {
		new_content = append(new_content, fmt.Sprintf("\n[%s]", label_tmp))
		for _, hostline := range AnsibleHosts[label_tmp] {
			new_content = append(new_content, strings.Join(hostline, " "))
		}
	}
	err = ioutil.WriteFile(HostsFileTmp, []byte(strings.Join(new_content, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("write to %s error:%v", HostsFileTmp, err)
	}

	// rename
	return os.Rename(HostsFileTmp, HostsFile)
}
