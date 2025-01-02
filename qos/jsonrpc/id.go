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

// String returns the string form of ID.
// strID field, if set, takes precedence as the returned value.
func (id ID) String() string {
	if id.strID != "" {
		return id.strID
	}

	return fmt.Sprintf("%d", id.intID)
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
