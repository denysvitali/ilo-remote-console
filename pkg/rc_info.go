package ilo

type RcInfo struct {
	EncKey string `json:"enc_key"`
	EncType int `json:"enc_type"`
	RcPort int `json:"rc_port"`
	VmKey string `json:"vm_key"`
	VmPort int `json:"vm_port"`
	CmdEncKey string `json:"cmd_enc_key"`
	ProtocolVersion string `json:"protocol_version"`
	OptionalFeatures string `json:"optional_features"`
	ServerName string `json:"server_name"`
	IloFQDN string `json:"ilo_fqdn"`
	Blade int `json:"blade"`
	Bay *int `json:"bay"`
	Enclosure *string `json:"enclosure"`
}
