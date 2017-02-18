package serverinfo

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
	Version             string    `xml:"version"`
	Theme               string    `xml:"theme"`
	EnableAvatars       bool      `xml:"enable_avatars"`
	EnablePreviews      bool      `xml:"enable_previews"`
	MemcacheLocal       string    `xml:"memcache.local"`
	MemcacheDistributed string    `xml:"memcache.distributed"`
	MemcacheLocking     string    `xml:"memcache.locking"`
	FilelockingEnabled  bool      `xml:"filelocking.enabled"`
	Debug               bool      `xml:"debug"`
	FreeSpace           int       `xml:"freespace"`
	CPULoad             []float64 `xml:"cpuload"`
	MemoryTotal         int       `xml:"mem_total"`
	MemoryFree          int       `xml:"mem_free"`
}

// Storage contains information about the nextcloud storage system.
type Storage struct {
	Users         int `xml:"num_users"`
	Files         int `xml:"num_files"`
	Storages      int `xml:"num_storages"`
	StoragesLocal int `xml:"num_storages_local"`
	StoragesHome  int `xml:"num_storages_home"`
	StoragesOther int `xml:"num_storages_other"`
}

// Shares contains information about nextcloud shares.
type Shares struct {
	SharesTotal          int `xml:"num_shares"`
	SharesUser           int `xml:"num_shares_user"`
	SharesGroups         int `xml:"num_shares_groups"`
	SharesLink           int `xml:"num_shares_link"`
	SharesLinkNoPassword int `xml:"num_shares_link_no_password"`
	FedSent              int `xml:"num_fed_shares_sent"`
	FedReceived          int `xml:"num_fed_shares_received"`
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
	MemoryLimit       int    `xml:"memory_limit"`
	MaxExecutionTime  int    `xml:"max_execution_time"`
	UploadMaxFilesize int    `xml:"upload_max_filesize"`
}

// Database contains information about the database used by nextcloud.
type Database struct {
	Type    string `xml:"type"`
	Version string `xml:"version"`
	Size    int    `xml:"size"`
}

// ActiveUsers contains statistics about the active users.
type ActiveUsers struct {
	Last5Minutes int `xml:"last5minutes"`
	LastHour     int `xml:"last1hour"`
	LastDay      int `xml:"last24hours"`
}
