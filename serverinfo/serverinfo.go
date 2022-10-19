package serverinfo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
)

const (
	// InfoPath contains the path to the serverinfo endpoint.
	InfoPath = "/ocs/v2.php/apps/serverinfo/api/v1/info?format=json"
)

// ServerInfo contains the complete data received from the server.
type ServerInfo struct {
	Meta Meta `json:"meta"`
	Data Data `json:"data"`
}

// Meta contains meta information about the result.
type Meta struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statuscode"`
	Message    string `json:"message"`
}

// Data contains the status information about the instance.
type Data struct {
	Nextcloud   Nextcloud   `json:"nextcloud"`
	Server      Server      `json:"server"`
	ActiveUsers ActiveUsers `json:"activeUsers"`
}

// Nextcloud contains information about the nextcloud installation.
type Nextcloud struct {
	System  System  `json:"system"`
	Storage Storage `json:"storage"`
	Shares  Shares  `json:"shares"`
}

// System contains nextcloud configuration and system information.
type System struct {
	Version             string `json:"version"`
	Theme               string `json:"theme"`
	EnableAvatars       bool   `json:"enable_avatars"`
	EnablePreviews      bool   `json:"enable_previews"`
	MemcacheLocal       string `json:"memcache.local"`
	MemcacheDistributed string `json:"memcache.distributed"`
	MemcacheLocking     string `json:"memcache.locking"`
	FilelockingEnabled  bool   `json:"filelocking.enabled"`
	Debug               bool   `json:"debug"`
	FreeSpace           int64  `json:"freespace"`
	Apps                Apps   `json:"apps"`
}

const boolYes = "yes"

func (s *System) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw struct {
		Version             string `xml:"version"`
		Theme               string `xml:"theme"`
		EnableAvatars       string `xml:"enable_avatars"`
		EnablePreviews      string `xml:"enable_previews"`
		MemcacheLocal       string `xml:"memcache.local"`
		MemcacheDistributed string `xml:"memcache.distributed"`
		MemcacheLocking     string `xml:"memcache.locking"`
		FilelockingEnabled  string `xml:"filelocking.enabled"`
		Debug               string `xml:"debug"`
		FreeSpace           int64  `xml:"freespace"`
		Apps                Apps   `xml:"apps"`
	}
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}
	s.Version = raw.Version
	s.Theme = raw.Theme
	s.EnableAvatars = raw.EnableAvatars == boolYes
	s.EnablePreviews = raw.EnablePreviews == boolYes
	s.MemcacheLocal = raw.MemcacheLocal
	s.MemcacheDistributed = raw.MemcacheDistributed
	s.MemcacheLocking = raw.MemcacheLocking
	s.FilelockingEnabled = raw.FilelockingEnabled == boolYes
	s.Debug = raw.Debug == boolYes
	s.FreeSpace = raw.FreeSpace
	s.Apps = raw.Apps
	return nil
}

func (s *System) UnmarshalJSON(data []byte) error {
	var raw struct {
		Version             string `json:"version"`
		Theme               string `json:"theme"`
		EnableAvatars       string `json:"enable_avatars"`
		EnablePreviews      string `json:"enable_previews"`
		MemcacheLocal       string `json:"memcache.local"`
		MemcacheDistributed string `json:"memcache.distributed"`
		MemcacheLocking     string `json:"memcache.locking"`
		FilelockingEnabled  string `json:"filelocking.enabled"`
		Debug               string `json:"debug"`
		FreeSpace           int64  `json:"freespace"`
		Apps                Apps   `json:"apps"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Version = raw.Version
	s.Theme = raw.Theme
	s.EnableAvatars = raw.EnableAvatars == boolYes
	s.EnablePreviews = raw.EnablePreviews == boolYes
	s.MemcacheLocal = raw.MemcacheLocal
	s.MemcacheDistributed = raw.MemcacheDistributed
	s.MemcacheLocking = raw.MemcacheLocking
	s.FilelockingEnabled = raw.FilelockingEnabled == boolYes
	s.Debug = raw.Debug == boolYes
	s.FreeSpace = raw.FreeSpace
	s.Apps = raw.Apps
	return nil
}

// Apps contains information about installed apps and updates.
type Apps struct {
	Installed        uint `json:"num_installed"`
	AvailableUpdates uint `json:"num_updates_available"`
}

// Storage contains information about the nextcloud storage system.
type Storage struct {
	Users         uint `json:"num_users"`
	Files         uint `json:"num_files"`
	Storages      uint `json:"num_storages"`
	StoragesLocal uint `json:"num_storages_local"`
	StoragesHome  uint `json:"num_storages_home"`
	StoragesOther uint `json:"num_storages_other"`
}

// Shares contains information about nextcloud shares.
type Shares struct {
	SharesTotal          uint `json:"num_shares"`
	SharesUser           uint `json:"num_shares_user"`
	SharesGroups         uint `json:"num_shares_groups"`
	SharesLink           uint `json:"num_shares_link"`
	SharesLinkNoPassword uint `json:"num_shares_link_no_password"`
	SharesMail           uint `json:"num_shares_mail"`
	SharesRoom           uint `json:"num_shares_room"`
	FedSent              uint `json:"num_fed_shares_sent"`
	FedReceived          uint `json:"num_fed_shares_received"`
	// <permissions_0_1>2</permissions_0_1>
	// <permissions_3_1>4</permissions_3_1>
	// <permissions_0_15>2</permissions_0_15>
	// <permissions_3_15>2</permissions_3_15>
	// <permissions_0_31>3</permissions_0_31>
	// <permissions_1_31>1</permissions_1_31>
}

// Server contains information about the servers running nextcloud.
type Server struct {
	Webserver string   `json:"webserver"`
	PHP       PHP      `json:"php"`
	Database  Database `json:"database"`
}

// PHP contains information about the PHP installation.
type PHP struct {
	Version           string `json:"version"`
	MemoryLimit       int64  `json:"memory_limit"`
	MaxExecutionTime  uint   `json:"max_execution_time"`
	UploadMaxFilesize int64  `json:"upload_max_filesize"`
}

// Database contains information about the database used by nextcloud.
type Database struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Size    uint64 `json:"size"`
}

func (d *Database) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type    string      `json:"type"`
		Version string      `json:"version"`
		Size    interface{} `json:"size"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	d.Type = raw.Type
	d.Version = raw.Version

	switch rawSize := raw.Size.(type) {
	case float64:
		if rawSize < 0 {
			return fmt.Errorf("negative value for database.size: %f", rawSize)
		}
		d.Size = uint64(rawSize)
	case string:
		parsedSize, err := strconv.ParseUint(rawSize, 10, 64)
		if err != nil {
			return fmt.Errorf("can not parse database.size %q: %w", rawSize, err)
		}
		d.Size = parsedSize
	default:
		return fmt.Errorf("unexpected type for database.size: %t", rawSize)
	}

	return nil
}

// ActiveUsers contains statistics about the active users.
type ActiveUsers struct {
	Last5Minutes uint `json:"last5minutes"`
	LastHour     uint `json:"last1hour"`
	LastDay      uint `json:"last24hours"`
}
