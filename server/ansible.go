package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"encoding/json"
	"golang.org/x/net/context"

	"github.com/xuebing1110/ansible-ext/host"
	"github.com/xuebing1110/ansible-ext/pkg/ansible"
	"github.com/xuebing1110/ansible-ext/pkg/consul"
	"github.com/xuebing1110/ansible-ext/pkg/ssh"
	pb "github.com/xuebing1110/ansible-ext/proto/ansible"
)

const (
	keyPrefix = "ansible/playbooks"
)

type AnsibleServer struct{}

func Convert2SshInfos_ansible(ls []*pb.SshInfo) []*host.SshInfo {
	ss := make([]*host.SshInfo, len(ls))
	for i, l := range ls {
		ss[i] = &host.SshInfo{
			Host:     l.Host,
			Port:     l.Port,
			UserName: l.UserName,
			Passwd:   l.Passwd,
		}
	}
	return ss
}

func Convert2PbResult_ansible(r *host.Result) *pb.Result {
	return &pb.Result{
		Host:    r.Host,
		Status:  r.Status,
		Message: r.Message,
		Reason:  r.Reason,
	}
}

func (s *AnsibleServer) CheckHost(ctx context.Context, req *pb.HostConfigure) (*pb.CommonReply, error) {
	for _, s := range req.SshInfos {
		log.Printf("get host check task %s ...\n", s.Host)
	}
	// HostLoginInfos
	hlb := host.NewHostLoginInfoBatch(Convert2SshInfos_ansible(req.SshInfos))

	// init and check
	hlb.Init()
	hlb.CheckPasswd()
	hlb.SetAuthType(ssh.LOGIN_USE_PASSWD)

	// check ping
	err := hlb.PingCheck()
	if err != nil {
		hlb.Reset()
		return nil, err
	}

	// check ssh
	hlb.SSHCheck(true)

	// response
	trs := make([]*pb.Result, len(hlb))
	for i, tr := range hlb {
		trs[i] = Convert2PbResult_ansible(tr.Result)
	}
	return &pb.CommonReply{Results: trs}, nil
}

func (s *AnsibleServer) AddHost(ctx context.Context, req *pb.HostConfigureWithLabel) (*pb.CommonReply, error) {
	for _, s := range req.SshInfos {
		log.Printf("get host add task %s ...\n", s.Host)
	}

	// HostLoginInfos
	hlb := host.NewHostLoginInfoBatch(Convert2SshInfos_ansible(req.SshInfos))

	// init and check
	hlb.Init()

	// ssh truth
	hlb.HostsSSHTrust()

	// response
	ahs := make([]ansible.AnsibleHost, 0)
	trs := make([]*pb.Result, len(hlb))
	for i, tr := range hlb {
		trs[i] = Convert2PbResult_ansible(tr.Result)

		if tr.Result.Status == host.STATUS_OK {
			ahs = append(ahs, ansible.AnsibleHost{tr.Host, tr.UserName})
		}
	}

	// save to ansible hosts file
	for _, labelvalue := range req.Labels {
		err := ansible.AddHosts(labelvalue, ahs)
		if err != nil {
			log.Printf("write ansible group %s to hosts file failed:%v", labelvalue, err)
		}
	}

	return &pb.CommonReply{Results: trs}, nil
}

func (s *AnsibleServer) RunPlayBook(req *pb.PlayBook, stream pb.Ansible_RunPlayBookServer) error {
	return s.Play(req, stream)
}

func (s *AnsibleServer) Play(req *pb.PlayBook, stream pb.Ansible_RunPlayBookServer) error {
	log.Printf("get playbok task %s %+v ...\n", req.Name, req.Hosts)

	// download playbook
	key := keyPrefix + "/" + req.Name + "/" + req.Version
	content, err := consul.GetKey(key)
	if err != nil {
		SendFailedMessageToStream(stream, req.Name, req.Hosts, fmt.Sprintf("download playbook %s failed %s", key, err.Error()))
		return err
	}
	bookpath := fmt.Sprintf("playbook/playbook.d/%s-%s.yml", req.Name, req.Version)
	err = ioutil.WriteFile(bookpath, content, 0666)
	if err != nil {
		SendFailedMessageToStream(stream, req.Name, req.Hosts, fmt.Sprintf("download playbook %s to file failed %s", key, err.Error()))
		return err
	}
	// defer os.Remove(bookpath)

	// send streaming message
	for msg := range runPlaybook(req.Name, bookpath, req.Hosts, req.Params, req.Register) {
		if err := stream.Send(msg); err != nil {
			return err
		}
	}

	log.Printf("finish running playbook %s", bookpath)
	return nil
}

func runPlaybook(name, bookpath string, hosts []string, params map[string]string, register *pb.Register) <-chan *pb.PlayBookMessage {
	msgs := make(chan *pb.PlayBookMessage, 1)

	// load playbook yaml
	var total_step int
	pbs, err := LoadPlayBook(bookpath)
	if err != nil {
		defer close(msgs)
		SendFailedMessage(msgs, name, hosts, fmt.Sprintf("load %s failed %s", bookpath, err.Error()))
		return msgs
	} else if len(pbs) == 0 {
		defer close(msgs)
		SendFailedMessage(msgs, name, hosts, fmt.Sprintf("get %s tasks failed %s", bookpath))
		return msgs
	} else {
		log.Printf("load %s suc!", bookpath)
		total_step = len(pbs[0].Tasks)
	}

	// execute playbook
	host_str := strings.Join(hosts, ",")
	playbook, err := ansible.PlayWithParams(host_str, bookpath, params)
	if err != nil {
		defer close(msgs)
		msgs <- &pb.PlayBookMessage{
			Job:     name,
			Type:    "ERROR",
			Message: err.Error(),
		}
		return msgs
	}

	overMap := make(map[string]bool)
	for _, host := range hosts {
		overMap[host] = false
	}

	go func() {
		defer close(msgs)

		for ret := range playbook.Messages() {
			switch ret.(type) {
			case *ansible.PlayBookMessage:
				pbm := ret.(*ansible.PlayBookMessage)
				msgs <- &pb.PlayBookMessage{
					Job:  name,
					Type: pbm.MsgType,
					Name: pbm.Name,
				}
			case *ansible.PlayBookTaskHost:
				pbth := ret.(*ansible.PlayBookTaskHost)

				// progress
				var progress int32
				if total_step > 0 {
					progress = int32(pbth.Step * 100 / total_step)
					if progress >= 100 {
						progress = 99
					}
				} else {
					progress = 0
				}

				msgs <- &pb.PlayBookMessage{
					Job:      name,
					Type:     "HOST",
					Host:     pbth.Host,
					Name:     pbth.Name,
					Status:   pbth.Status,
					Message:  pbth.Message,
					Step:     int32(pbth.Step),
					Progress: progress,
				}
			case *ansible.PlayBookRecap:
				pbr := ret.(*ansible.PlayBookRecap)
				var status = "ok"
				if pbr.Unreach > 0 || pbr.Failed > 0 {
					status = "fatal"
				}
				msgs <- &pb.PlayBookMessage{
					Job:      name,
					Type:     "RECAP",
					Host:     pbr.Host,
					Ok:       int32(pbr.OK),
					Changed:  int32(pbr.Changed),
					Unreach:  int32(pbr.Unreach),
					Failed:   int32(pbr.Failed),
					Progress: 100,
					Status:   status,
				}
				overMap[pbr.Host] = true
			}
		}

		for _, host := range hosts {
			over := overMap[host]
			if !over {
				msgs <- &pb.PlayBookMessage{
					Job:      name,
					Type:     "RECAP",
					Host:     host,
					Ok:       0,
					Changed:  0,
					Unreach:  1,
					Failed:   1,
					Progress: 100,
					Status:   "fatal",
				}
			} else { // register service to consul
				if register != nil {
					err := consul.RegisterService(name, host, register.ListenPort, register.ConsulAddress, register.DataCenter, register.Labels)
					if err != nil {
						log.Printf("registe %s's %s srv to consul failed: %v\n", host, name, err)
					}
				}
			}
		}
	}()

	return msgs
}

func SendFailedMessage(msg_chan chan *pb.PlayBookMessage, job string, hosts []string, message string) {
	for _, host := range hosts {
		msg_chan <- &pb.PlayBookMessage{
			Job:      job,
			Type:     "RECAP",
			Host:     host,
			Ok:       0,
			Changed:  0,
			Unreach:  1,
			Failed:   1,
			Progress: 100,
			Status:   "invalid",
			Message:  message,
		}
	}
}

func SendFailedMessageToStream(stream pb.Ansible_RunPlayBookServer, job string, hosts []string, message string) {
	for _, host := range hosts {
		stream.Send(&pb.PlayBookMessage{
			Job:      job,
			Type:     "RECAP",
			Host:     host,
			Ok:       0,
			Changed:  0,
			Unreach:  1,
			Failed:   1,
			Progress: 100,
			Status:   "invalid",
			Message:  message,
		})
	}
}

var REG_PLAYBOOK_PATH *regexp.Regexp

func init() {
	// PATH:  /api/v1/playbooks/:name/:version
	REG_PLAYBOOK_PATH = regexp.MustCompile(`.*\/playbooks/([^\/]+)(?:\/([^\/]+))?$`)
}

func PlayBookServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var name, version string
		matched := REG_PLAYBOOK_PATH.FindStringSubmatch(r.URL.Path)
		switch len(matched) {
		case 2:
			name = matched[1]
		case 3:
			name = matched[1]
			version = matched[2]
		default:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, getBadResponse(fmt.Errorf("unknown path %s", r.URL.Path)))
			return
		}

		// playbook handler
		if version != "" {
			key := keyPrefix + "/" + name + "/" + version
			log.Printf("start to handle playbooks %s", key)
			playBookHandler(key, w, r)
			return
		} else {
			// list playbook
			if r.Method == http.MethodGet {
				log.Printf("start to list playbooks using key \"%s\"", keyPrefix+name)
				keys, err := consul.ListKeys(keyPrefix + "/" + name)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, getBadResponse(err))
					return
				}

				resp := getKeysResponse(keys, keyPrefix)
				data, _ := json.Marshal(resp)
				fmt.Fprint(w, string(data))
				return
			} else {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, getBadResponse(errors.New("method not allowed")))
				return
			}
		}
	}
}

type PalyBookResp struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Link    string `json:"link"`
}

func getKeysResponse(keys []string, prefix string) []PalyBookResp {
	resp := make([]PalyBookResp, len(keys))
	for i, key := range keys {
		name_version := strings.Split(strings.TrimPrefix(key, prefix+"/"), "/")
		resp[i] = PalyBookResp{
			name_version[0],
			name_version[1],
			fmt.Sprintf("/api/v1/ansible/playbooks/%s/%s", name_version[0], name_version[1]),
		}
	}

	return resp
}

func playBookHandler(key string, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, getBadResponse(err))
		return
	}
	defer r.Body.Close()

	switch r.Method {
	case http.MethodPost, http.MethodPut:
		time, err := consul.PutKey(key, body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, getBadResponse(err))
			return
		}
		fmt.Fprintf(w, `{"took": "%s"}`, time.String())
	case http.MethodGet:
		content, err := consul.GetKey(key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, getBadResponse(err))
			return
		}
		fmt.Fprint(w, string(content))
	case http.MethodDelete:
		time, err := consul.DelKey(key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, getBadResponse(err))
			return
		}
		fmt.Fprintf(w, `{"took": "%s"}`, time.String())
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, getBadResponse(errors.New("method not allowed")))
		return
	}
}

func getBadResponse(err error) string {
	return fmt.Sprintf(`{"error":"%s"}`, err.Error())
}
