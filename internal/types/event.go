package types

type AuthEvent interface {
	iAmAnEvent()
}

type FailedEvent struct{ Error error }
type SessionAuthorizedEvent struct{ Session Session }
type LinkEvent struct{ Link string }
type CodeReceivedEvent struct{}

func (SessionAuthorizedEvent) iAmAnEvent() {}
func (LinkEvent) iAmAnEvent()              {}
func (FailedEvent) iAmAnEvent()            {}
func (CodeReceivedEvent) iAmAnEvent()      {}
