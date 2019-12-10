package ping

import (
	"testing"
)

func TestPing(t *testing.T) {
	hosts := []string{"10.138.16.192", "3.3.3.3"}
	ret, err := GetStatus(hosts)
	if err != nil {
		t.Fatal(err)
	}

	if ret[0] != 1 {
		t.Fatalf("%s expect 1,but get %d", hosts[0], ret[0])
	}
	if ret[1] != 0 {
		t.Fatalf("%s expect 0,but get %d", hosts[1], ret[1])
	}
}
