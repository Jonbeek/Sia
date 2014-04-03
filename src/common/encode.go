package common

func MarshalUpdate(u Update) Message {
	return Message{u.SwarmId(), u.UpdateId(), u.MarshalString(), u.Type()}
}
