package jsonrpc

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// TODO_TECHDEBT(@commoddity): handle all possible ID values based on JSONRPC spec.
// Specifically, the ability to handle the "null" value as defined in the spec.
// See the following link for more details:
// https://www.jsonrpc.org/specification
//
// JSON-RPC ID requirements:
// - Must be a String, Number, or NULL if included
// - Should not be NULL in normal operation
// - Numbers should not contain fractional parts
// - Server must reply with the same value in the Response object
// - Used to correlate context between request/response
type ID struct {
	intID int
	strID string
}

// String returns ID as a string.
// strID takes precedence if both fields are set.
func (id ID) String() string {
	if id.strID != "" {
		return id.strID
	}

	return fmt.Sprintf("%d", id.intID)
}

// Int returns ID as an int.
// intID takes precedence over strId if both fields are set.
func (id ID) Int() int {
	if id.intID != 0 {
		return id.intID
	}
	parsed, err := strconv.Atoi(id.strID)
	if err != nil {
		return 0
	}
	return parsed
}

func (id ID) IsEmpty() bool {
	return id.intID == 0 && id.strID == ""
}

func (id ID) MarshalJSON() ([]byte, error) {
	if id.intID != 0 {
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
