package surf

import (
	"testing"
)

func TestIsZero(t *testing.T) {
	if bl := isZero(nil); !bl {
		t.Fail()
	}
	if bl := isZero(make(map[string]string)); bl {
		t.Fail()
	}
	var mapVal map[string]string
	if bl := isZero(mapVal); !bl {
		t.Fail()
	}
	if bl := isZero(""); !bl {
		t.Fail()
	}
	if bl := isZero(0); !bl {
		t.Fail()
	}
}
