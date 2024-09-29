package jsonrpc

import (
	"encoding/json"
)

// TODO_TECHDEBT: handle all possible ID values based on JSONRPC spec.
// See the following link for more details:
// https://www.jsonrpc.org/specification
type ID struct {
	intID int
	strID string
}

func (id ID) MarshalJSON() ([]byte, error) {
	if id.intID > 0 {
		idWithInt := idWithInt{ID: id.intID}
		return json.Marshal(idWithInt)
	}

	idWithStr := idWithStr{ID: id.strID}
	return json.Marshal(idWithStr)
}

func (id *ID) UnmarshalJSON(data []byte) error {
	var idWithInt idWithInt
	if err := json.Unmarshal(data, &idWithInt); err == nil {
		id.intID = idWithInt.ID
		return nil
	}

	var idWithStr idWithStr
	err := json.Unmarshal(data, &idWithStr)
	if err != nil {
		return err
	}

	id.strID = idWithStr.ID
	return nil
}

func IDFromInt(id int) ID {
	return ID{intID: id}
}

func IDFromStr(id string) ID {
	return ID{strID: id}
}

type idWithInt struct {
	ID int `json:"id"`
}

type idWithStr struct {
	ID string `json:"id"`
}
