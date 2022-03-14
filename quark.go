package quark

import "encoding/json"

func NewQuark() *Quark {
	return &Quark{
		Marshal:   json.Marshal,
		Unmarshal: json.Unmarshal,
		option:    &Option{},
	}
}
