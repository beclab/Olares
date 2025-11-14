package lvm

import (
	"encoding/json"
)

const (
	LVS = "lvs"
	VGS = "vgs"
	PVS = "pvs"
)

/*
	{
	    "report": [
	        {
	            "lv": [
	                {"lv_name":"data", "vg_name":"olares-vg", "lv_attr":"-wi-ao----", "lv_size":"1.76t", "pool_lv":"", "origin":"", "data_percent":"", "metadata_percent":"", "move_pv":"", "mirror_log":"", "copy_percent":"", "convert_lv":""},
	                {"lv_name":"root", "vg_name":"olares-vg", "lv_attr":"-wi-ao----", "lv_size":"100.00g", "pool_lv":"", "origin":"", "data_percent":"", "metadata_percent":"", "move_pv":"", "mirror_log":"", "copy_percent":"", "convert_lv":""},
	                {"lv_name":"swap", "vg_name":"olares-vg", "lv_attr":"-wi-ao----", "lv_size":"1.00g", "pool_lv":"", "origin":"", "data_percent":"", "metadata_percent":"", "move_pv":"", "mirror_log":"", "copy_percent":"", "convert_lv":""}
	            ]
	        }
	    ]
	}
*/
type LvItem struct {
	LvName          string `json:"lv_name"`
	VgName          string `json:"vg_name"`
	LvAttr          string `json:"lv_attr"`
	LvSize          string `json:"lv_size"`
	PoolLv          string `json:"pool_lv"`
	Origin          string `json:"origin"`
	DataPercent     string `json:"data_percent"`
	MetadataPercent string `json:"metadata_percent"`
	MovePv          string `json:"move_pv"`
	MirrorLog       string `json:"mirror_log"`
	CopyPercent     string `json:"copy_percent"`
	ConvertLv       string `json:"convert_lv"`
}

type LvsResult struct {
	Report []struct {
		Lv []LvItem `json:"lv"`
	} `json:"report"`
}

func CommandLVS() *command[LvsResult] {
	cmd := findCmd(LVS)

	return &command[LvsResult]{
		cmd:         cmd,
		defaultArgs: []string{"--reportformat", "json"},

		format: func(data []byte) (LvsResult, error) {
			var res LvsResult
			err := json.Unmarshal(data, &res)
			return res, err
		},
	}
}

/*
	{
	    "report": [
	        {
	            "vg": [
	                {"vg_name":"olares-vg", "pv_count":"1", "lv_count":"3", "snap_count":"0", "vg_attr":"wz--n-", "vg_size":"1.86t", "vg_free":"0 "}
	            ]
	        }
	    ]
	}
*/
type VgItem struct {
	VgName    string `json:"vg_name"`
	PvCount   string `json:"pv_count"`
	LvCount   string `json:"lv_count"`
	SnapCount string `json:"snap_count"`
	VgAttr    string `json:"vg_attr"`
	VgSize    string `json:"vg_size"`
	VgFree    string `json:"vg_free"`
}

type VgsResult struct {
	Report []struct {
		Vg []VgItem `json:"vg"`
	} `json:"report"`
}

func CommandVGS() *command[VgsResult] {
	cmd := findCmd(VGS)

	return &command[VgsResult]{
		cmd:         cmd,
		defaultArgs: []string{"--reportformat", "json"},

		format: func(data []byte) (VgsResult, error) {
			var res VgsResult
			err := json.Unmarshal(data, &res)
			return res, err
		},
	}
}

/*
	{
	    "report": [
	        {
	            "pv": [
	                {"pv_name":"/dev/nvme0n1p2", "vg_name":"olares-vg", "pv_fmt":"lvm2", "pv_attr":"a--", "pv_size":"1.86t", "pv_free":"0 "}
	            ]
	        }
	    ]
	}
*/
type PvItem struct {
	PvName string `json:"pv_name"`
	VgName string `json:"vg_name"`
	PvFmt  string `json:"pv_fmt"`
	PvAttr string `json:"pv_attr"`
	PvSize string `json:"pv_size"`
	PvFree string `json:"pv_free"`
}

type PvsResult struct {
	Report []struct {
		Pv []PvItem `json:"pv"`
	} `json:"report"`
}

func CommandPVS() *command[PvsResult] {
	cmd := findCmd(PVS)

	return &command[PvsResult]{
		cmd:         cmd,
		defaultArgs: []string{"--reportformat", "json"},

		format: func(data []byte) (PvsResult, error) {
			var res PvsResult
			err := json.Unmarshal(data, &res)
			return res, err
		},
	}
}
