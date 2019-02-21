package serverinfo

import (
	"encoding/xml"
	"io"
)

// Parse reads ServerInfo from a Reader.
func Parse(r io.Reader) (ServerInfo, error) {
	result := ServerInfo{}
	if err := xml.NewDecoder(r).Decode(&result); err != nil {
		return ServerInfo{}, err
	}

	return result, nil
}
