package errors

import "fmt"

type TError struct {
	Msg string `json:"err"`
}

func (p TError) Error() string {
	return p.Msg
}

func (p TError) String() string {
	return p.Msg
}
func Bomb(format string, a ...interface{}) {
	panic(TError{Msg: fmt.Sprintf(format, a...)})
}

func Dangerous(v interface{}) {
	if v == nil {
		return
	}
	switch t := v.(type) {
	case string:
		if t != "" {
			panic(TError{Msg: t})
		}
	case error:
		panic(TError{Msg: t.Error()})
	}
}
