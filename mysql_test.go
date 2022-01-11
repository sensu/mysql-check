package main

import (
	"testing"
)

func TestGatherIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skiping integration tests")
	}

	servers := []string{"root@tcp(localhost)/mysql"}

	actualFamilies, err := GatherMetrics(servers)
	if err != nil {
		t.Errorf("unexpected error gathering metrics: %v", err)
	}
	for _, family := range actualFamilies {
		if len(family.Metric) != 1 {
			t.Errorf("expected one metric point per family. Got %d: %v", len(family.Metric), family)
		}
		var serverLabelVal string
		for _, l := range family.Metric[0].Label {
			if *l.Name == "server" {
				serverLabelVal = *l.Value
			}
		}
		if serverLabelVal != "localhost:3306" {
			t.Errorf("expected label 'server=localhost:3306': got %s", serverLabelVal)
		}
	}

}
