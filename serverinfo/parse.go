package serverinfo

import (
	"encoding/json"
	"io"
)

// ParseJSON reads ServerInfo from a Reader in JSON format.
func ParseJSON(r io.Reader) (ServerInfo, error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return ServerInfo{}, err
	}

	result := struct {
		ServerInfo ServerInfo `json:"ocs"`
	}{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return ServerInfo{}, err
	}

	return result.ServerInfo, nil
}
