package ansible

import (
	"fmt"
	"testing"
)

func TestPlay(t *testing.T) {
	cmd, err := Play("10.138.16.192,10.138.40.224", "hostname.yml")
	if err != nil {
		t.Fatal(err)
	}

	for ret := range cmd.Messages() {
		fmt.Printf("%T: %+v\n", ret, ret)

		switch ret.(type) {
		case *PlayBookMessage:
		case *PlayBookTaskHost:
			pbth := ret.(*PlayBookTaskHost)
			if !pbth.IsOK() {
				t.Fatalf("the %s task in %s failed:", pbth.Name, pbth.Host)
			}
		}
	}

	cmd.Wait()
}
