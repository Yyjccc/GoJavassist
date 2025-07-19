package reflect

import "testing"

func TestCut(t *testing.T) {
	err := CutJar("../lib/rt.jar", "../lib/cut.jar")
	if err != nil {
		t.Fatal(err)
	}
}
