package xmpp

import "io"

type tee struct {
	socket io.Writer
	logger io.Writer
}

func (t tee) Write(p []byte) (n int, err error) {
	for _, w := range []io.Writer{t.socket, t.logger} {
		n, err = w.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	t.logger.Write([]byte("\n\n"))
	return len(p), nil
}
