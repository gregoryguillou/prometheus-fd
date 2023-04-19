package main

import "testing"

func TestFindProcesses(t *testing.T) {
	a, err := PIDs("prometheus-fd")
	if err != nil {
		t.Error(err)
	}
	if len(a) != 1 {
		t.Error("return one pid, instead:", a)
	}
}

func TestNumberOfSockets(t *testing.T) {
	a, err := PIDs("prometheus-fd")
	if err != nil {
		t.Error(err)
	}
	if len(a) != 1 {
		t.Error("return one pid, instead:", a)
	}
	for _, pid := range a {
		i, err := numberOfFiles(pid)
		if err != nil {
			t.Error(err)
		}
		if i < 1 || i > 20 {
			t.Error("unexpected number of files", i)
		}
	}
}
