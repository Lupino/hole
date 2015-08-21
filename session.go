package hole

// Session define session.
type Session struct {
	ID []byte
	r  *ReadStream
	w  *WriteStream
}

// NewSession create a new session with sessionID and conn.
func NewSession(sessionID []byte, conn Conn) Session {
	return Session{
		ID: sessionID,
		r:  NewReadStream(),
		w: &WriteStream{
			conn:      conn,
			sessionID: sessionID,
		},
	}
}
