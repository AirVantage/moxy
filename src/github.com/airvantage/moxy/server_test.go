package moxy

import (
	"strings"
	"testing"
)

func TestWalkFilters(t *testing.T) {
	m := map[string]interface{}{"A": "B"}

	input := []byte("YeaH")

	// upper case filter
	f1 :=
		filterize(func(inPacket []byte, uplink bool, metadata map[string]interface{}) (outPacket []byte, backward []byte) {
			if !uplink {
				t.Error("not uplink")
			}

			if metadata == nil {
				t.Error("missing metadata")
			}

			return append(inPacket, []byte(strings.ToUpper(string(inPacket)))...), []byte{}
		})

	// lowercase filter
	f2 :=
		filterize(func(inPacket []byte, uplink bool, metadata map[string]interface{}) (outPacket []byte, backward []byte) {
			if !uplink {
				t.Error("not uplink")
			}

			if metadata == nil {
				t.Error("missing metadata")
			}

			return append(inPacket, []byte(strings.ToLower(string(inPacket)))...), []byte{}
		})

	filters := []MqttFilter{f1, f2}
	res := walkFilters(input, filters, true, m)
	if string(res) != "YeaHYEAHyeahyeah" {
		t.Error("expected YeaHYEAHyeahyeah, received ", string(res))
	}
}

type filterize func(inPacket []byte, uplink bool, metadata map[string]interface{}) (outPacket []byte, backward []byte)

func (f filterize) Filter(inPacket []byte, uplink bool, metadata map[string]interface{}) (outPacket []byte, backward []byte) {
	return f(inPacket, uplink, metadata)
}
