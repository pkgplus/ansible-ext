package ansible

import (
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
)

var (
	LABELNAME_ADDRESS = model.LabelName("__address__")
)

func SetStaticTargets(scfg *config.ScrapeConfig, labeles map[string]string, targets []string) {
	lset := getLabelSet(labeles)
	targetLabel := getTargets(targets)

	var found bool
	staticCfgs := scfg.ServiceDiscoveryConfig.StaticConfigs
	for i, staticCfg := range staticCfgs {
		if lset.Equal(staticCfg.Labels) {
			staticCfgs[i].Targets = targetLabel
			found = true
		}
	}

	if !found {
		scfg.ServiceDiscoveryConfig.StaticConfigs = append(
			scfg.ServiceDiscoveryConfig.StaticConfigs,
			&config.TargetGroup{
				Targets: targetLabel,
				Labels:  lset,
				Source:  lset.String(),
			},
		)
	}
}
func AddStaticTarget(scfg *config.ScrapeConfig, labeles map[string]string, targets []string) {
	lset := getLabelSet(labeles)
	targetLabel := getTargets(targets)

	var found bool
	staticCfgs := scfg.ServiceDiscoveryConfig.StaticConfigs
	for i, staticCfg := range staticCfgs {
		if lset.Equal(staticCfg.Labels) {
			for _, new_target := range targetLabel {
				var repeat bool
				for _, old_target := range staticCfgs[i].Targets {
					if old_target[LABELNAME_ADDRESS] == new_target[LABELNAME_ADDRESS] {
						repeat = true
						break
					}
				}
				if !repeat {
					staticCfgs[i].Targets = append(staticCfgs[i].Targets, new_target)
				}
			}

			found = true
		}
	}

	if !found {
		scfg.ServiceDiscoveryConfig.StaticConfigs = append(
			scfg.ServiceDiscoveryConfig.StaticConfigs,
			&config.TargetGroup{
				Targets: targetLabel,
				Labels:  lset,
				Source:  lset.String(),
			},
		)
	}
}

func GetScrapeConfig(scfgs []*config.ScrapeConfig, jobname string) (*config.ScrapeConfig, bool) {
	for _, cfg := range scfgs {
		if cfg.JobName == jobname {
			return cfg, true
		}
	}

	return nil, false
}

func getLabelSet(labeles map[string]string) model.LabelSet {
	lset := make(map[model.LabelName]model.LabelValue)
	for lname, lvalue := range labeles {
		lset[model.LabelName(lname)] = model.LabelValue(lvalue)
	}

	return lset
}

func getTargets(targets []string) []model.LabelSet {
	targetLabel := make([]model.LabelSet, len(targets))
	for i, target := range targets {
		targetLabel[i] = model.LabelSet(
			map[model.LabelName]model.LabelValue{
				LABELNAME_ADDRESS: model.LabelValue(target),
			})
	}

	return targetLabel
}
