package jail

import (
	"fmt"
	"os"
	"testing"
)

func init() {
	if os.Getuid() != 0 {
		fmt.Println("Tests must be run as root :(")
		os.Exit(1)
	}
}

func TestCreateJail(t *testing.T) {
	newJail, er := NewJail("heehee", "/tmp")
	if er != nil {
		t.Fatal(er)
	}
	defer func() {
		if er := newJail.Destroy(); er != nil {
			t.Log("Unable to remove jail -- you will need to clear it yourself")
			t.Fatal(er)
		}
	}()

	if newJail.Jid() == 0 {
		t.Errorf("New jail's JID is 0")
		return
	}

	jails, er := EnumerateJails()
	if er != nil {
		t.Error(er)
		return
	}

	if len(jails) != 1 {
		t.Errorf("Jail created but not enumerable")
		return
	}

	if jails[0].Jid() != newJail.Jid() {
		t.Errorf("Jail JIDs don't match")
	}
}
