package serverinfo

import "encoding/xml"

// ServerInfo contains the complete data received from the server.
type ServerInfo struct {
	Meta Meta `xml:"meta"`
	Data Data `xml:"data"`
}

// Meta contains meta information about the result.
type Meta struct {
	Status     string `xml:"status"`
	StatusCode int    `xml:"statuscode"`
	Message    string `xml:"message"`
}

// Data contains the status information about the instance.
type Data struct {
	Nextcloud   Nextcloud   `xml:"nextcloud"`
	Server      Server      `xml:"server"`
	ActiveUsers ActiveUsers `xml:"activeUsers"`
}

// Nextcloud contains information about the nextcloud installation.
type Nextcloud struct {
	System  System  `xml:"system"`
	Storage Storage `xml:"storage"`
	Shares  Shares  `xml:"shares"`
}

// System contains nextcloud configuration and system information.
type System struct {
	Version             string `xml:"version"`
	Theme               string `xml:"theme"`
	EnableAvatars       bool   `xml:"enable_avatars"`
	EnablePreviews      bool   `xml:"enable_previews"`
	MemcacheLocal       string `xml:"memcache.local"`
	MemcacheDistributed string `xml:"memcache.distributed"`
	MemcacheLocking     string `xml:"memcache.locking"`
	FilelockingEnabled  bool   `xml:"filelocking.enabled"`
	Debug               bool   `xml:"debug"`
	FreeSpace           int64  `xml:"freespace"`
	// <cpuload>
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
	return nil
}

// Storage contains information about the nextcloud storage system.
type Storage struct {
	Users         uint `xml:"num_users"`
	Files         uint `xml:"num_files"`
	Storages      uint `xml:"num_storages"`
	StoragesLocal uint `xml:"num_storages_local"`
	StoragesHome  uint `xml:"num_storages_home"`
	StoragesOther uint `xml:"num_storages_other"`
}

// Shares contains information about nextcloud shares.
type Shares struct {
	SharesTotal          uint `xml:"num_shares"`
	SharesUser           uint `xml:"num_shares_user"`
	SharesGroups         uint `xml:"num_shares_groups"`
	SharesLink           uint `xml:"num_shares_link"`
	SharesLinkNoPassword uint `xml:"num_shares_link_no_password"`
	FedSent              uint `xml:"num_fed_shares_sent"`
	FedReceived          uint `xml:"num_fed_shares_received"`
	// <permissions_0_1>2</permissions_0_1>
	// <permissions_3_1>4</permissions_3_1>
	// <permissions_0_15>2</permissions_0_15>
	// <permissions_3_15>2</permissions_3_15>
	// <permissions_0_31>3</permissions_0_31>
	// <permissions_1_31>1</permissions_1_31>
}

// Server contains information about the servers running nextcloud.
type Server struct {
	Webserver string   `xml:"webserver"`
	PHP       PHP      `xml:"php"`
	Database  Database `xml:"database"`
}

// PHP contains information about the PHP installation.
type PHP struct {
	Version           string `xml:"version"`
	MemoryLimit       int64  `xml:"memory_limit"`
	MaxExecutionTime  uint   `xml:"max_execution_time"`
	UploadMaxFilesize int64  `xml:"upload_max_filesize"`
}

// Database contains information about the database used by nextcloud.
type Database struct {
	Type    string `xml:"type"`
	Version string `xml:"version"`
	Size    uint64 `xml:"size"`
}

// ActiveUsers contains statistics about the active users.
type ActiveUsers struct {
	Last5Minutes uint `xml:"last5minutes"`
	LastHour     uint `xml:"last1hour"`
	LastDay      uint `xml:"last24hours"`
}
