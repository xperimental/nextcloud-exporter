package serverinfo

import (
	"encoding/json"
	"io"
)

// ParseJSON reads ServerInfo from a Reader in JSON format.
func ParseJSON(r io.Reader) (ServerInfo, error) {
	result := struct {
		ServerInfo ServerInfo `json:"ocs"`
	}{}
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return ServerInfo{}, err
	}

	return result.ServerInfo, nil
}
