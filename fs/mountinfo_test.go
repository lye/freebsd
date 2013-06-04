package fs

import (
	"testing"
)

func TestAllMountsMounted(t *testing.T) {
	mountInfo := GetMountInfo()

	for _, info := range mountInfo {
		if !info.IsMounted() {
			t.Errorf("[%s] %s -> %s should be mounted", info.FsTypeName(), info.MntFromName(), info.MntToName())
		}
	}
}

func TestRootMounted(t *testing.T) {
	mountInfo, er := MountInfoForPath("/")
	if er != nil {
		t.Error(er)
		return
	}

	if !mountInfo.IsMounted() {
		t.Errorf("...root fs isn't mounted?")
	}
}
