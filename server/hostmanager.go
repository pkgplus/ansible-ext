package server

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/xuebing1110/ansible-ext/host"
	"github.com/xuebing1110/ansible-ext/pkg/ansible"
	"github.com/xuebing1110/ansible-ext/pkg/ssh"
	pb "github.com/xuebing1110/ansible-ext/proto/hostmanager"
)

type HostManagerServer struct{}

func convertHm2SshInfo(ls []*pb.LoginInfo) []*host.SshInfo {
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

func convertHm2PbResult(r *host.Result) *pb.TaskResult {
	return &pb.TaskResult{
		Host:    r.Host,
		Status:  r.Status,
		Message: r.Message,
		Reason:  r.Reason,
	}
}

func (s *HostManagerServer) Precheck(ctx context.Context, req *pb.PrecheckRequest) (*pb.InitOrPrecheckReply, error) {
	// HostLoginInfos
	hlb := host.NewHostLoginInfoBatch(convertHm2SshInfo(req.LoginInfos))

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
	trs := make([]*pb.TaskResult, len(hlb))
	for i, tr := range hlb {
		trs[i] = convertHm2PbResult(tr.Result)
	}
	return &pb.InitOrPrecheckReply{Results: trs}, nil
}

func (s *HostManagerServer) InitHosts(ctx context.Context, req *pb.InitRequest) (*pb.InitOrPrecheckReply, error) {

	// HostLoginInfos
	hlb := host.NewHostLoginInfoBatch(convertHm2SshInfo(req.LoginInfos))

	// init and check
	hlb.Init()

	// ssh truth
	hlb.HostsSSHTrust()

	// response
	ahs := make([]ansible.AnsibleHost, 0)
	trs := make([]*pb.TaskResult, len(hlb))
	for i, tr := range hlb {
		trs[i] = convertHm2PbResult(tr.Result)

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

	return &pb.InitOrPrecheckReply{Results: trs}, nil
}

func (s *HostManagerServer) Install(req *pb.InstallRequest, stream pb.HostManager_InstallServer) error {
	// for i := 0; i < 5; i++ {
	// 	msg := &pb.InstallMessage{
	// 		Job:     "node",
	// 		Type:    "machine",
	// 		Host:    "127.0.0.1",
	// 		Step:    fmt.Sprintf("%d", i+1),
	// 		Name:    "create exporter group",
	// 		Status:  "OK",
	// 		Message: "",
	// 	}
	// 	if err := stream.Send(msg); err != nil {
	// 		return err
	// 	}
	// 	time.Sleep(time.Second)
	// }

	//job => hosts
	jobMap := make(map[string][]string)
	for host, jobs := range req.Jobs {
		for _, job := range jobs.AnsibleJobs {
			if _, found := jobMap[job]; !found {
				jobMap[job] = make([]string, 0)
			}
			jobMap[job] = append(jobMap[job], host)
		}
	}

	//ervery job
	msgs := make(chan *pb.InstallMessage, 1)
	var wg sync.WaitGroup
	for job, hosts := range jobMap {
		wg.Add(1)
		go func(job string, hosts []string) {
			defer wg.Done()

			//bookpath := job
			//bookinfo, BookDictfound := playBookConvertDict[job]
			//if !strings.HasSuffix(job, ".yml") && strings.Index(job, "/") == -1 {
			//	if BookDictfound {
			//		bookpath = fmt.Sprintf("playbook/%s.yml", bookinfo.Name)
			//	} else {
			//		bookpath = fmt.Sprintf("playbook/%s.yml", job)
			//	}
			//}

			// return if playbook not exist
			bookpath := GetPlayBookPath(job)
			if !PathExists(bookpath) {
				SendHmFailedMessage(msgs, job, hosts, fmt.Sprintf("%s was not found", bookpath))
				return
			}

			// load playbook yaml
			var total_step int
			pbs, err := LoadPlayBook(bookpath)
			if err != nil {
				SendHmFailedMessage(msgs, job, hosts, fmt.Sprintf("load %s failed %s", bookpath, err.Error()))
				return
			} else if len(pbs) == 0 {
				SendHmFailedMessage(msgs, job, hosts, fmt.Sprintf("get %s tasks failed %s", bookpath))
				return
			} else {
				total_step = len(pbs[0].Tasks)
			}

			// execute playbook
			host_str := strings.Join(hosts, ",")
			playbook, err := ansible.Play(host_str, bookpath)
			if err != nil {
				msgs <- &pb.InstallMessage{
					Job:     job,
					Type:    "ERROR",
					Message: err.Error(),
				}
				return
			}

			overMap := make(map[string]bool)
			for _, host := range hosts {
				overMap[host] = false
			}

			for ret := range playbook.Messages() {
				switch ret.(type) {
				case *ansible.PlayBookMessage:
					pbm := ret.(*ansible.PlayBookMessage)
					msgs <- &pb.InstallMessage{
						Job:  job,
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

					msgs <- &pb.InstallMessage{
						Job:      job,
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
					msgs <- &pb.InstallMessage{
						Job:      job,
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
					msgs <- &pb.InstallMessage{
						Job:      job,
						Type:     "RECAP",
						Host:     host,
						Ok:       0,
						Changed:  0,
						Unreach:  1,
						Failed:   1,
						Progress: 100,
						Status:   "fatal",
					}
				} else { // registe service to consul
					//err := RegisterSrv(job, host, req.Labels)
					//if err != nil {
					//	log.Printf("registe %s's %s srv to consul failed:%v", host, job, err)
					//}
				}
			}
			// playbook.Wait()
		}(job, hosts)
	}

	// wait job exec completed
	go func() {
		log.Println("wait job completed...")
		wg.Wait()
		close(msgs)
	}()

	// send streaming message
	for msg := range msgs {
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
	log.Println("write completed...")

	return nil
}

func (s *HostManagerServer) Install2(req *pb.InstallRequest2, stream pb.HostManager_Install2Server) error {
	//task => hosts
	jobMap := make(map[string][]string)
	for _, hostJobs := range req.Jobs {
		host := hostJobs.Host
		for _, job_name := range hostJobs.Names {
			// TODO: support another task type
			if _, found := jobMap[job_name]; !found {
				jobMap[job_name] = make([]string, 0)
			}
			jobMap[job_name] = append(jobMap[job_name], host)
		}
	}

	//ervery job
	msgs := make(chan *pb.InstallMessage, 1)
	var wg sync.WaitGroup
	for job, hosts := range jobMap {
		wg.Add(1)
		go func(job string, hosts []string) {
			defer wg.Done()

			// return if playbook not exist
			bookpath := GetPlayBookPath(job)
			if !PathExists(bookpath) {
				SendHmFailedMessage(msgs, job, hosts, fmt.Sprintf("%s was not found", bookpath))
				return
			}

			// load playbook yaml
			var total_step int
			pbs, err := LoadPlayBook(bookpath)
			if err != nil {
				SendHmFailedMessage(msgs, job, hosts, fmt.Sprintf("load %s failed %s", bookpath, err.Error()))
				return
			} else if len(pbs) == 0 {
				SendHmFailedMessage(msgs, job, hosts, fmt.Sprintf("get %s tasks failed %s", bookpath))
				return
			} else {
				total_step = len(pbs[0].Tasks)
			}

			// execute playbook
			host_str := strings.Join(hosts, ",")
			playbook, err := ansible.PlayWithParams(host_str, bookpath, req.Params)
			if err != nil {
				msgs <- &pb.InstallMessage{
					Job:     job,
					Type:    "ERROR",
					Message: err.Error(),
				}
				return
			}

			overMap := make(map[string]bool)
			for _, host := range hosts {
				overMap[host] = false
			}

			for ret := range playbook.Messages() {
				switch ret.(type) {
				case *ansible.PlayBookMessage:
					pbm := ret.(*ansible.PlayBookMessage)
					msgs <- &pb.InstallMessage{
						Job:  job,
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

					msgs <- &pb.InstallMessage{
						Job:      job,
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
					msgs <- &pb.InstallMessage{
						Job:      job,
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
					msgs <- &pb.InstallMessage{
						Job:      job,
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
					// err := RegisteSrv(job, host, req.Labels)
					// if err != nil {
					// 	log.Printf("registe %s's %s srv to consul failed:%v", host, job, err)
					// }
				}
			}
			// playbook.Wait()
		}(job, hosts)
	}

	// wait job exec completed
	go func() {
		log.Println("wait job completed...")
		wg.Wait()
		close(msgs)
	}()

	// send streaming message
	for msg := range msgs {
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
	log.Println("write completed...")

	return nil
}

func SendHmFailedMessage(msg_chan chan *pb.InstallMessage, job string, hosts []string, message string) {
	for _, host := range hosts {
		msg_chan <- &pb.InstallMessage{
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
