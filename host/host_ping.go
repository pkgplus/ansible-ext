package host

import (
	"ansible-ext/pkg/ping"
)

var (
	MSG_PING_FAILED    = "ping不通"
	REASON_PING_FAILED = "ping timeout"
)

func (hl *HostLoginInfo) PingStatus() error {
	if hl.Result.Status == STATUS_FAILED {
		return nil
	}

	pingret, err := ping.GetStatus([]string{hl.Host})
	if err != nil {
		return err
	}

	if pingret[0] != 1 {
		hl.Result.Status = STATUS_FAILED
		hl.Result.Message = MSG_PING_FAILED
		hl.Result.Reason = REASON_PING_FAILED
	}

	return nil
}

func (hlb HostLoginInfoBatch) PingCheck() error {
	ips := make([]string, len(hlb))
	for i, hl := range hlb {
		ips[i] = hl.Host
	}

	pingrets, err := ping.GetStatus(ips)
	if err != nil {
		return err
	}

	for i, hl := range hlb {
		if pingrets[i] != 1 {
			hl.Result.Status = STATUS_FAILED
			hl.Result.Message = MSG_PING_FAILED
			hl.Result.Reason = REASON_PING_FAILED
		}
	}
	return nil
}
