package aerror

import (
	"testing"
)

func TestA(t *testing.T) {
	err := dog()
	if err != nil {
		err = With(err, "err2")
	}
	me := err.(MultiLiner)
	println(me.MultiLine())
	if err != nil {
		b := err.Error()
		println(b)
	}
}

func dog() error {
	er := cat()
	if er != nil {
		return With(er, "err1")
	}
	return er
}

func cat() error {
	return New("err0")
}
