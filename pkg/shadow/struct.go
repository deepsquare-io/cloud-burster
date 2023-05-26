package shadow

type VM struct {
	AffectedOn       string        `json:"affected_on"`
	BlockDevices     []interface{} `json:"block_devices"`
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
	VMPublicIPv4     interface{}   `json:"vm_public_ipv4"`
	VMPublicSSHPort  interface{}   `json:"vm_public_sshport"`
	VMRAM            int           `json:"vm_ram"`
	VMSKU            string        `json:"vm_sku"`
	VNC              bool          `json:"vnc"`
}

type Data struct {
	Filters struct{} `json:"filters"`
	VMs     []VM     `json:"vms"`
}

type ListResponse struct {
	Data Data
}
