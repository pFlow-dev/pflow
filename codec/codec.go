package codec

import (
	json "github.com/gibson042/canonicaljson-go"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
)

type Oid struct {
	cid.Cid
}

func (o Oid) String() string {
	return o.Encode(encoder)
}

func (o Oid) Bytes() []byte {
	return []byte(o.Encode(encoder))
}

func toCid(data []byte) (cid.Cid, error) {
	return cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   multihash.SHA2_256,
		MhLength: -1, // default length
	}.Sum(data)
}

func ToOid(data []byte) *Oid {
	id, err := toCid(data)
	if err != nil {
		panic(err)
	}
	return &Oid{id}
}

var encoder, _ = multibase.EncoderByName("base58btc")

// Cat prepends schema and encodes using base58btc
func Cat(schema string, b ...[]byte) (key []byte) {
	key = []byte(schema)
	for _, v := range b {
		key = append(key, v...)
	}
	return ToOid(key).Bytes()
}

func Marshal(i interface{}) []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return data
}

func Unmarshal(data []byte, any interface{}) error {
	return json.Unmarshal(data, any)
}
