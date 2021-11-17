package quark

type JsonMarshalFunc func(x interface{}) ([]byte, error)
type JsonUnmarshalFunc func(b []byte, x interface{}) error
