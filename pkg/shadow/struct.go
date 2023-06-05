package shadow

import "time"

// BlockDeviceList is the Block Device field in the response from the /block_device/list endpoint
type BlockDeviceList struct {
	AllocatedOn     *time.Time `json:"allocated_on"`
	Cost            int        `json:"cost"`
	DatacenterLabel string     `json:"datacenter_label"`
	ID              int        `json:"id"`
	InsertedOn      time.Time  `json:"inserted_on"`
	Mount           bool       `json:"mount"`
	ReleasedOn      *time.Time `json:"released_on"`
	SizeGiB         int        `json:"size_gib"`
	Status          int        `json:"status"`
	StatusStr       string     `json:"status_str"`
	UUID            string     `json:"uuid"`
}

// BlockDeviceListResponse is the response from the /block_device/list endpoint
type BlockDeviceListResponse struct {
	BlockDevices []BlockDeviceList `json:"block_devices"`
	Filters      struct{}          `json:"filters"`
}

// BlockDevice is the Block Device field in the response from the /vm/list endpoint
type BlockDevice struct {
	UUID string `json:"uuid"`
}

// VM is the VM field from the /vm/list endpoint
type VM struct {
	AffectedOn       string        `json:"affected_on"`
	BlockDevices     []BlockDevice `json:"block_devices"`
	DatacenterLabel  interface{}   `json:"datacenter_label"`
	ID               int           `json:"id"`
	Image            string        `json:"image"`
	InsertedOn       string        `json:"inserted_on"`
	KillRequestedOn  interface{}   `json:"kill_requested_on"`
	LaunchBashScript string        `json:"launch_bash_script"`
	MaxUptime        int           `json:"max_uptime"`
	RequestTimeout   int           `json:"request_timeout"`
	StartedOn        interface{}   `json:"started_on"`
	Status           int           `json:"status"`
	StatusStr        string        `json:"status_str"`
	Uptime           int           `json:"uptime"`
	UUID             string        `json:"uuid"`
	VMCore           int           `json:"vm_core"`
	VMCost           int           `json:"vm_cost"`
	VMGPU            int           `json:"vm_gpu"`
	VMPublicIPv4     string        `json:"vm_public_ipv4"`
	VMPublicSSHPort  int           `json:"vm_public_sshport"`
	VMRAM            int           `json:"vm_ram"`
	VMSKU            string        `json:"vm_sku"`
	VNC              bool          `json:"vnc"`
}

// VMListResponse is the response from the /vm/list endpoint
type VMListResponse struct {
	Filters struct{} `json:"filters"`
	VMs     []VM     `json:"vms"`
}
