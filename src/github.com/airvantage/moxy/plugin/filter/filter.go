package filter

import (
	"encoding/gob"
	"github.com/airvantage/moxy/plugin"
)

// A plugin for filtering MQTT packets
type FilterPlugin struct {
	*plugin.Plugin
}

func NewFilterPlugin(name string, id string) *FilterPlugin {
	var res FilterPlugin

	res.Plugin = plugin.NewPlugin(name, id+".sock")
	return &res
}

func (fp *FilterPlugin) Filter(inPacket []byte, uplink bool) (outPacket []byte, backward []byte) {

	var call struct {
		InPacket []byte
		Uplink   bool
	}
	call.InPacket = inPacket
	call.Uplink = uplink

	c, err := fp.Dial()

	if err != nil {
		panic(err)
	}

	defer c.Close()

	enc := gob.NewEncoder(c)
	enc.Encode(call)

	dec := gob.NewDecoder(c)

	var result struct {
		OutPacket []byte
		Backward  []byte
	}

	err = dec.Decode(&result)
	if err != nil {
		panic(err)
	}

	return result.OutPacket, result.Backward
}
