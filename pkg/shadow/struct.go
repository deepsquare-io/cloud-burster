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
	Image            string           `json:"image"`
	UUID             string           `json:"uuid"`
	Uptime           int64            `json:"uptime"`
	VMRAM            int64            `json:"vm_ram"`
	VMCore           int64            `json:"vm_core"`
	VMGPU            int64            `json:"vm_gpu"`
	VMCost           int64            `json:"vm_cost"`
	VNC              bool             `json:"vnc"`
	Interruptible    bool             `json:"interruptible"`
	VNCPassword      *string          `json:"vnc_password"`
	Status           int64            `json:"status"`
	StatusStr        StatusStr        `json:"status_str"`
	InsertedOn       string           `json:"inserted_on"`
	AffectedOn       *string          `json:"affected_on"`
	MaxUptime        int64            `json:"max_uptime"`
	RequestTimeout   int64            `json:"request_timeout"`
	StartedOn        *string          `json:"started_on"`
	KillRequestedOn  *string          `json:"kill_requested_on"`
	LaunchBashScript string           `json:"launch_bash_script"`
	VMSku            VMSku            `json:"vm_sku"`
	VMPublicIPv4     *VMPublicIPv4    `json:"vm_public_ipv4"`
	VMPublicSSHPort  *int64           `json:"vm_public_sshport"`
	DatacenterLabel  *DatacenterLabel `json:"datacenter_label"`
	BlockDevices     []BlockDevice    `json:"block_devices"`
}

type DatacenterLabel string

type StatusStr string

type VMSku string

type VMPublicIPv4 string

// VMListResponse is the response from the /vm/list endpoint
type VMListResponse struct {
	Filters struct{} `json:"filters"`
	VMs     []VM     `json:"vms"`
}
