package common

func MarshalUpdate(u Update) NetworkMessage {
	return NetworkMessage{u.SwarmId(), u.UpdateId(), u.MarshalString(), u.Type()}
}
