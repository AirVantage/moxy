package moxy

// MqqtFilter is fitering an incoing uplink or downlink MQTT packet, by generating a new packet (or just proxyfying the received one,
// and possibly send backway a packet
type MqttFilter interface {
	Filter(inPacket []byte, uplink bool, metadata map[string]interface{}) (outPacket []byte, backward []byte)
}
