package model

import (
	"encoding/json"
	"github.com/pflow-dev/go-metamodel/metamodel"
	"github.com/pflow-dev/pflow/codec"
)

type Schema struct {
	*metamodel.PetriNet
	Version string
}

type Schemata struct {
	Schemata map[string]*Schema
}

var Model = &Schemata{
	Schemata: make(map[string]*Schema, 0),
}

func (m *Schemata) Load(jsonSource string) {
	oid := codec.ToOid([]byte(jsonSource))
	n := new(metamodel.PetriNet)
	err := json.Unmarshal([]byte(jsonSource), n)
	if n.Schema != "" {
		m.Schemata[n.Schema] = &Schema{PetriNet: n, Version: oid.String()}
	}

	if err != nil {
		panic(err)
	}
}

func (m *Schemata) ToJson() []byte {
	return codec.Marshal(m.Schemata)
}
