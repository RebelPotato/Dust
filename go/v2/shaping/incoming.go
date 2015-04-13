package shaping

import (
	"io"

	"github.com/blanu/Dust/go/proc"
)

const (
	incomingBufSize = 2048
)

type incoming struct {
	decoder Decoder
	uniform io.Writer
	visible io.Reader

	inBuf   []byte
	pushBuf []byte
}

func (ic *incoming) Init(
	parent *proc.Env,
	uniform io.Writer,
	visible io.Reader,
	decoder Decoder,
) {
	*ic = incoming{
		decoder: decoder,
		uniform: uniform,
		visible: visible,

		inBuf:   make([]byte, 0, incomingBufSize),
		pushBuf: make([]byte, 0, incomingBufSize),
	}

	proc.InitHelper(parent, ic.runIncoming)
}

func (ic *incoming) runIncoming(env *proc.Env) error {
	// TODO: expand buf API and use it here
	for {
		env.CancellationPoint()
		log.Debug("[ - reading up to %d bytes", cap(ic.inBuf) - len(ic.inBuf))
		rn, err := ic.visible.Read(ic.inBuf[len(ic.inBuf):cap(ic.inBuf)])
		if err != nil {
			return err
		}
		log.Debug("[-  read %d bytes", rn)
		ic.inBuf = ic.inBuf[:len(ic.inBuf)+rn]

		env.CancellationPoint()
		dn, sn := ic.decoder.UnshapeBytes(ic.pushBuf[len(ic.pushBuf):cap(ic.pushBuf)], ic.inBuf)
		log.Debug("[-- unshaped %d from %d bytes", dn, sn)
		ic.pushBuf = ic.pushBuf[:len(ic.pushBuf)+dn]
		ic.inBuf = ic.inBuf[:copy(ic.inBuf, ic.inBuf[sn:])]

		env.CancellationPoint()
		log.Debug("[   uniform: %d bytes", len(ic.pushBuf))
		wn, err := ic.uniform.Write(ic.pushBuf)
		if err != nil && err != io.ErrShortWrite {
			return err
		}
		ic.pushBuf = ic.pushBuf[:copy(ic.pushBuf, ic.pushBuf[wn:])]
	}
}
