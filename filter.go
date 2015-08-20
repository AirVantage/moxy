package moxy

// MqttFilter is fitering an incoming uplink or downlink MQTT packet, by generating a new packet (or just proxyfying the received one,
// and possibly send backway a packet
type MqttFilter interface {
	Connected(metadata map[string]interface{})
	Filter(inPacket []byte, uplink bool, metadata map[string]interface{}) (outPacket []byte, backward []byte)
	Disconnected(metadata map[string]interface{})
}
