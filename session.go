package hole

type Session struct {
    Id []byte
    r *ReadStream
    w *WriteStream
}

func NewSession(sessionId []byte, conn Conn) Session {
    return Session{
        Id: sessionId,
        r: &ReadStream{},
        w: &WriteStream{
            conn: conn,
        },
    }
}
