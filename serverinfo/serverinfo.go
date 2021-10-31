package serverinfo

import (
	"encoding/json"
	"encoding/xml"
)

const (
	// InfoPath contains the path to the serverinfo endpoint.
	InfoPath = "/ocs/v2.php/apps/serverinfo/api/v1/info?format=json"
)

// ServerInfo contains the complete data received from the server.
type ServerInfo struct {
	Meta Meta `xml:"meta" json:"meta"`
	Data Data `xml:"data" json:"data"`
}

// Meta contains meta information about the result.
type Meta struct {
	Status     string `xml:"status" json:"status"`
	StatusCode int    `xml:"statuscode" json:"statuscode"`
	Message    string `xml:"message" json:"message"`
}

// Data contains the status information about the instance.
type Data struct {
	Nextcloud   Nextcloud   `xml:"nextcloud" json:"nextcloud"`
	Server      Server      `xml:"server" json:"server"`
	ActiveUsers ActiveUsers `xml:"activeUsers" json:"activeUsers"`
}

// Nextcloud contains information about the nextcloud installation.
type Nextcloud struct {
	System  System  `xml:"system" json:"system"`
	Storage Storage `xml:"storage" json:"storage"`
	Shares  Shares  `xml:"shares" json:"shares"`
}

// System contains nextcloud configuration and system information.
type System struct {
	Version             string `xml:"version" json:"version"`
	Theme               string `xml:"theme" json:"theme"`
	EnableAvatars       bool   `xml:"enable_avatars" json:"enable_avatars"`
	EnablePreviews      bool   `xml:"enable_previews" json:"enable_previews"`
	MemcacheLocal       string `xml:"memcache.local" json:"memcache.local"`
	MemcacheDistributed string `xml:"memcache.distributed" json:"memcache.distributed"`
	MemcacheLocking     string `xml:"memcache.locking" json:"memcache.locking"`
	FilelockingEnabled  bool   `xml:"filelocking.enabled" json:"filelocking.enabled"`
	Debug               bool   `xml:"debug" json:"debug"`
	FreeSpace           int64  `xml:"freespace" json:"freespace"`
	Apps                Apps   `xml:"apps" json:"apps"`
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
	Installed        uint `xml:"num_installed" json:"num_installed"`
	AvailableUpdates uint `xml:"num_updates_available" json:"num_updates_available"`
}

// Storage contains information about the nextcloud storage system.
type Storage struct {
	Users         uint `xml:"num_users" json:"num_users"`
	Files         uint `xml:"num_files" json:"num_files"`
	Storages      uint `xml:"num_storages" json:"num_storages"`
	StoragesLocal uint `xml:"num_storages_local" json:"num_storages_local"`
	StoragesHome  uint `xml:"num_storages_home" json:"num_storages_home"`
	StoragesOther uint `xml:"num_storages_other" json:"num_storages_other"`
}

// Shares contains information about nextcloud shares.
type Shares struct {
	SharesTotal          uint `xml:"num_shares" json:"num_shares"`
	SharesUser           uint `xml:"num_shares_user" json:"num_shares_user"`
	SharesGroups         uint `xml:"num_shares_groups" json:"num_shares_groups"`
	SharesLink           uint `xml:"num_shares_link" json:"num_shares_link"`
	SharesLinkNoPassword uint `xml:"num_shares_link_no_password" json:"num_shares_link_no_pasword"`
	FedSent              uint `xml:"num_fed_shares_sent" json:"num_fed_shares_sent"`
	FedReceived          uint `xml:"num_fed_shares_received" json:"num_fed_shares_received"`
	// <permissions_0_1>2</permissions_0_1>
	// <permissions_3_1>4</permissions_3_1>
	// <permissions_0_15>2</permissions_0_15>
	// <permissions_3_15>2</permissions_3_15>
	// <permissions_0_31>3</permissions_0_31>
	// <permissions_1_31>1</permissions_1_31>
}

// Server contains information about the servers running nextcloud.
type Server struct {
	Webserver string   `xml:"webserver" json:"webserver"`
	PHP       PHP      `xml:"php" json:"php"`
	Database  Database `xml:"database" json:"database"`
}

// PHP contains information about the PHP installation.
type PHP struct {
	Version           string `xml:"version" json:"version"`
	MemoryLimit       int64  `xml:"memory_limit" json:"memory_limit"`
	MaxExecutionTime  uint   `xml:"max_execution_time" json:"max_execution_time"`
	UploadMaxFilesize int64  `xml:"upload_max_filesize" json:"upload_max_filesize"`
}

// Database contains information about the database used by nextcloud.
type Database struct {
	Type    string `xml:"type" json:"type"`
	Version string `xml:"version" json:"version"`
	Size    uint64 `xml:"size" json:"size"`
}

// ActiveUsers contains statistics about the active users.
type ActiveUsers struct {
	Last5Minutes uint `xml:"last5minutes" json:"last5minutes"`
	LastHour     uint `xml:"last1hour" json:"last1hour"`
	LastDay      uint `xml:"last24hours" json:"last24hours"`
}
