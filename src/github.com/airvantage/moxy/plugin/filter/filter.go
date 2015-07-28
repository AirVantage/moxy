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

func (fp *FilterPlugin) Filter(inPacket []byte, uplink bool, metadata map[string]interface{}) (outPacket []byte, backward []byte) {

	var call struct {
		InPacket []byte
		Uplink   bool
		Metatada map[string]interface{}
	}
	call.InPacket = inPacket
	call.Uplink = uplink
	call.Metatada = metadata

	c, err := fp.Dial()

	if err != nil {
		panic(err)
	}

	defer c.Close()

	enc := gob.NewEncoder(c)
	err = enc.Encode("FILTER")
	if err != nil {
		panic(err)
	}
	err = enc.Encode(call)
	if err != nil {
		panic(err)
	}

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

func (fp *FilterPlugin) Connected(metadata map[string]interface{}) {
	var call struct {
		Metadata map[string]interface{}
	}

	c, err := fp.Dial()

	if err != nil {
		panic(err)
	}

	defer c.Close()

	enc := gob.NewEncoder(c)
	err = enc.Encode("CONNECT")
	if err != nil {
		panic(err)
	}
	err = enc.Encode(call)

	if err != nil {
		panic(err)
	}
}

func (fp *FilterPlugin) Disconnected(metadata map[string]interface{}) {
	var call struct {
		Metadata map[string]interface{}
	}

	c, err := fp.Dial()

	if err != nil {
		panic(err)
	}

	defer c.Close()

	enc := gob.NewEncoder(c)
	err = enc.Encode("CONNECT")
	if err != nil {
		panic(err)
	}
	err = enc.Encode(call)

	if err != nil {
		panic(err)
	}
}
