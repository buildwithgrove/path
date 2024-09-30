package jsonrpc

import (
	"encoding/json"
	"fmt"
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
		return []byte(fmt.Sprintf("%d", id.intID)), nil
	}

	return []byte(fmt.Sprintf("%q", id.strID)), nil
}

func (id *ID) UnmarshalJSON(data []byte) error {
	var intID int
	if err := json.Unmarshal(data, &intID); err == nil {
		id.intID = intID
		return nil
	}

	var strID string
	err := json.Unmarshal(data, &strID)
	if err != nil {
		return err
	}

	id.strID = strID
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
