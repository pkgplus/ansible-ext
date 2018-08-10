package ansible

import (
	// "fmt"
	"github.com/prometheus/prometheus/config"
)

import (
	"testing"
)

func TestConfig(t *testing.T) {
	filename := "prometheus.yml"
	cfg, err := config.LoadFile(filename)
	if err != nil {
		t.Fatalf("load %s failed:%v", filename, err)
	} else {
		// fmt.Printf("%+v", cfg.ScrapeConfigs[1].ServiceDiscoveryConfig.StaticConfigs[0].Labels.String())
	}

	scfg, found := GetScrapeConfig(cfg.ScrapeConfigs, "linux")
	if !found {
		t.Fatalf("get job ScrapeConfig failed!")
	}

	SetStaticTargets(
		scfg,
		map[string]string{
			"cluster": "BJ",
			"project": "HBDM",
		},
		[]string{
			"10.159.63.135-linux:10051",
			"10.159.63.152-linux:10051",
			"10.159.63.156-linux:10051",
		},
	)

	AddStaticTarget(
		scfg,
		map[string]string{
			"cluster": "BJ",
			"project": "HBDM",
		},
		[]string{
			"10.159.63.135-linux:10051",
		},
	)

	// fmt.Println(cfg.String())
}
