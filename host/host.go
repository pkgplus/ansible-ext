package host

import (
	"github.com/xuebing1110/ansible-ext/pkg/ssh"
	//pb "github.com/xuebing1110/ansible-ext/proto/HostManager"
)

var (
	STATUS_OK     = "OK"
	STATUS_FAILED = "FAILED"
	STATUS_UN     = "UN"

	MSG_SSH_NO_USER      = "用户名不能为空"
	MSG_SSH_NO_PASSWD    = "密码不能为空"
	REASON_SSH_NO_USER   = "username and passwd is requred"
	REASON_SSH_NO_PASSWD = "passwd is requred"
)

type HostLoginInfo struct {
	*SshInfo

	authType uint8
	Result   *Result `json:"-"`
}

type SshInfo struct {
	Host     string `json:"host"`
	Port     int32  `json:"port"`
	UserName string `json:"userName,omitempty"`
	Passwd   string `json:"passwd,omitempty"`
}

type Result struct {
	Host    string `json:"host,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// type HostLoginResult struct {
// 	Host    string `json:"host"`
// 	Status  string `json:"status"`
// 	Message string `json:"message"`
// 	Reason  string `json:"reason"`
// }

func NewHostLoginInfo(sshInfo *SshInfo) *HostLoginInfo {
	hli := &HostLoginInfo{
		SshInfo: sshInfo,
	}
	hli.Init()
	return hli
}

func (hl *HostLoginInfo) Init() {
	hl.Result = &Result{
		Host:   hl.Host,
		Status: STATUS_OK,
	}

	if hl.authType == 0 {
		hl.authType = ssh.LOGIN_USE_ANY
	}

	if hl.UserName == "" {
		hl.Result.Status = STATUS_FAILED
		hl.Result.Message = MSG_SSH_NO_USER
		hl.Result.Reason = REASON_SSH_NO_USER
		return
	}

	if hl.Port == 0 {
		hl.Port = 22
	}
}

func (hl *HostLoginInfo) SetAuthType(at uint8) {
	hl.authType = at
}

func (hl *HostLoginInfo) SetAuthUserPwd() {
	hl.authType = ssh.LOGIN_USE_PASSWD
}

func (hl *HostLoginInfo) CheckPasswd() {
	if hl.Passwd == "" {
		hl.Result.Status = STATUS_FAILED
		hl.Result.Message = MSG_SSH_NO_PASSWD
		hl.Result.Reason = REASON_SSH_NO_PASSWD
	}
}

func (hl *HostLoginInfo) Reset() {
	hl.Result.Status = STATUS_UN
}

type HostLoginInfoBatch []*HostLoginInfo

func NewHostLoginInfoBatch(lis []*SshInfo) HostLoginInfoBatch {
	hl_slice := make([]*HostLoginInfo, 0, len(lis))
	for _, li := range lis {
		hl_slice = append(hl_slice, NewHostLoginInfo(li))
	}
	return HostLoginInfoBatch(hl_slice)
}

func (hlb HostLoginInfoBatch) Init() {
	for _, hl := range hlb {
		hl.Init()
	}
}
func (hlb HostLoginInfoBatch) CheckPasswd() {
	for _, hl := range hlb {
		hl.CheckPasswd()
	}
}
func (hlb HostLoginInfoBatch) Reset() {
	for _, hl := range hlb {
		hl.Reset()
	}
}

func (hlb HostLoginInfoBatch) SetAuthType(at uint8) {
	for _, hl := range hlb {
		hl.SetAuthType(at)
	}
}
func (hlb HostLoginInfoBatch) SetAuthUserPwd() {
	for _, hl := range hlb {
		hl.SetAuthUserPwd()
	}
}
