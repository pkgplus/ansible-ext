package ansible

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"

	hexec "ansible-ext/pkg/exec"
)

var (
	REGEXP_TASK_RET  = regexp.MustCompile(`(?m)^(\w+):\s*\[([^\]]+)\](?:: (\S+) => (\{.*\}))?\s*$`)
	REGEXP_TASK_HEAD = regexp.MustCompile(`(?m)^(\w+)\s+(\w+|\[[^\]]+\])\s+\*+$`)
	//REGEXP_TASK_HEAD = regexp.MustCompile(`(?m)^(\w+)\s+(\w+|\[[\]]+\])\s+\*+$`)
	REGEXP_PLAY_RET = regexp.MustCompile(`(?m)^(\S+)\s*:\s*ok=(\d+)\s+changed=(\d+)\s+unreachable=(\d+)\s+failed=(\d+)\s*$`)
)

type PlayBookMessage struct {
	Name    string `json:"name"`
	MsgType string `json:"type"`
	Step    int    `json:"step,omitempty"`
}

type PlayBookTaskHost struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Host    string `json:"host"`
	Message string `json:"mssage"`
	Step    int    `json:"step,omitempty"`
}
type PlayBookRecap struct {
	Host    string `json:"host"`
	OK      int    `json:"ok"`
	Changed int    `json:"changed"`
	Unreach int    `json:"unreachable"`
	Failed  int    `json:"failed"`
}

func (p *PlayBookTaskHost) IsOK() bool {
	switch p.Status {
	case "ok", "changed":
		return true
	default:
		return false
	}
}

func Play(host string, book string) (*hexec.Command, error) {
	var a_args = []string{
		book,
		"-e",
		fmt.Sprintf("hosts=%s", host),
	}

	fmt.Printf("ARGS: %+v\n", a_args)
	hc := hexec.NewCommand("ansible-playbook", a_args...)
	err := hc.RunStreaming(&AnsiblePlaybookParser{})
	if err != nil {
		return hc, err
	}

	return hc, nil
}

func PlayWithParams(host string, book string, params map[string]string) (*hexec.Command, error) {
	var a_args = make([]string, 0, len(params)*2+3)
	a_args = append(a_args, book,
		"-e",
		fmt.Sprintf("hosts=%s", host))

	for name, value := range params {
		a_args = append(a_args, "-e", fmt.Sprintf("%s='%s'", name, value))
	}

	log.Printf("ARGS: %+v\n", a_args)
	hc := hexec.NewCommand("ansible-playbook", a_args...)
	err := hc.RunStreaming(&AnsiblePlaybookParser{})
	if err != nil {
		return hc, err
	}

	return hc, nil
}

type AnsiblePlaybookParser struct {
	playBookName string
	taskName     string
	recap        bool
	step         int
}

func (app *AnsiblePlaybookParser) Execute(line []byte, outChan chan interface{}, eof bool) {
	// fmt.Printf("----->\n%s\n<------\n", string(line))

	task_head_args := REGEXP_TASK_HEAD.FindAllSubmatch(line, -1)
	if len(task_head_args) > 0 {
		for _, task_head_arg := range task_head_args {
			headtype := string(task_head_arg[1])
			headname := string(TrimBracket(task_head_arg[2]))

			pbm := &PlayBookMessage{
				Name:    headname,
				MsgType: headtype,
			}

			switch headtype {
			case "PLAY":
				app.playBookName = headname
				if headname == "RECAP" {
					app.recap = true
				}
			case "TASK":
				app.taskName = headname
				app.step++
				pbm.Step = app.step
			}

			outChan <- pbm
		}
	} else if app.recap {
		recap_rets := REGEXP_PLAY_RET.FindAllSubmatch(line, -1)
		for _, recap_ret := range recap_rets {
			pbr := &PlayBookRecap{
				Host: string(recap_ret[1]),
			}
			pbr.OK, _ = strconv.Atoi(string(recap_ret[2]))
			pbr.Changed, _ = strconv.Atoi(string(recap_ret[3]))
			pbr.Unreach, _ = strconv.Atoi(string(recap_ret[4]))
			pbr.Failed, _ = strconv.Atoi(string(recap_ret[5]))
			outChan <- pbr
		}
	} else if app.taskName != "" {
		task_rets := REGEXP_TASK_RET.FindAllSubmatch(line, -1)
		for _, task_ret := range task_rets {
			pbth := &PlayBookTaskHost{
				Name:   app.taskName,
				Status: string(task_ret[1]),
				Host:   string(task_ret[2]),
				Step:   app.step,
			}

			if len(task_ret) >= 5 {
				msgMap := make(map[string]interface{})
				err := json.Unmarshal(task_ret[4], &msgMap)
				if err != nil {
					// fmt.Printf("%v\n", err)
					pbth.Message = string(task_ret[4])
				} else {
					if msgi, ok := msgMap["msg"]; ok {
						pbth.Message = msgi.(string)
					} else if msgi, ok := msgMap["stderr"]; ok {
						pbth.Message = msgi.(string)
					} else if msgi, ok := msgMap["stdout"]; ok {
						pbth.Message = msgi.(string)
					} else {
						pbth.Message = string(task_ret[4])
					}
				}
			}

			outChan <- pbth
		}
	}

	// if eof {
	// 	close(outChan)
	// }
}

func TrimBracket(c []byte) []byte {
	return bytes.TrimSuffix(bytes.TrimPrefix(c, []byte{'['}), []byte{']'})
}
