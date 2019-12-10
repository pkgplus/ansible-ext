package ansible

import (
	"fmt"
	"testing"
)

func TestExec(t *testing.T) {
	cmd, err := Exec("10.138.16.192,10.138.40.224", "-m", "shell", "-a", `date;hostname`)
	if err != nil {
		t.Fatal(err)
	}

	for ret := range cmd.Messages() {
		har, ok := ret.(*HostAnsibleResult)
		if !ok {
			t.Fatalf("expect *HostAnsibleResult, but get %+v", ret)
		}

		fmt.Printf("%+v\n", har)
		if har.Status != STATUS_SUCC {
			t.Fatalf("exec failed,get status: %s\n", har.Status)
		}
	}

	cmd.Wait()
}
