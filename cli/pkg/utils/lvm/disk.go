package lvm

import (
	"bytes"
	"encoding/json"
)

/*
lsblk -J

	{
	   "blockdevices": [
	      {
	         "name": "nvme0n1",
	         "maj:min": "259:0",
	         "rm": false,
	         "size": "1.9T",
	         "ro": false,
	         "type": "disk",
	         "mountpoints": [
	             null
	         ],
	         "children": [
	            {
	               "name": "nvme0n1p1",
	               "maj:min": "259:1",
	               "rm": false,
	               "size": "512M",
	               "ro": false,
	               "type": "part",
	               "mountpoints": [
	                   "/boot/efi"
	               ]
	            },{
	               "name": "nvme0n1p2",
	               "maj:min": "259:2",
	               "rm": false,
	               "size": "1.9T",
	               "ro": false,
	               "type": "part",
	               "mountpoints": [
	                   null
	               ],
	               "children": [
	                  {
	                     "name": "olares--vg-swap",
	                     "maj:min": "252:0",
	                     "rm": false,
	                     "size": "1G",
	                     "ro": false,
	                     "type": "lvm",
	                     "mountpoints": [
	                         "[SWAP]"
	                     ]
	                  },{
	                     "name": "olares--vg-root",
	                     "maj:min": "252:1",
	                     "rm": false,
	                     "size": "100G",
	                     "ro": false,
	                     "type": "lvm",
	                     "mountpoints": [
	                         "/"
	                     ]
	                  },{
	                     "name": "olares--vg-data",
	                     "maj:min": "252:2",
	                     "rm": false,
	                     "size": "1.8T",
	                     "ro": false,
	                     "type": "lvm",
	                     "mountpoints": [
	                         "/olares", "/var"
	                     ]
	                  }
	               ]
	            }
	         ]
	      }
	   ]
	}
*/
const LBLK = "lsblk"

type BlkPart struct {
	Name        string           `json:"name"`
	MajMin      string           `json:"maj:min"`
	Rm          bool             `json:"rm"`
	Size        string           `json:"size"`
	Ro          bool             `json:"ro"`
	Type        string           `json:"type"`
	Mountpoints BlkList[string]  `json:"mountpoints"`
	Children    BlkList[BlkPart] `json:"children,omitempty"`
}

type BlkList[T any] []T

type BlkResult struct {
	Blockdevices BlkList[BlkPart] `json:"blockdevices"`
}

func CommandLBLK() *command[BlkResult] {
	cmd := findCmd(LBLK)
	return &command[BlkResult]{
		cmd:         cmd,
		defaultArgs: []string{"-J"},

		format: func(data []byte) (BlkResult, error) {
			var res BlkResult
			err := json.Unmarshal(data, &res)
			return res, err
		},
	}
}

func (s *BlkList[T]) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if bytes.Equal(b, []byte("null")) {
		*s = nil
		return nil
	}
	var raws []json.RawMessage
	if err := json.Unmarshal(b, &raws); err != nil {
		return err
	}
	var out []T
	for _, r := range raws {
		if bytes.Equal(bytes.TrimSpace(r), []byte("null")) {
			continue
		}
		var v T
		if err := json.Unmarshal(r, &v); err != nil {
			return err
		}
		out = append(out, v)
	}
	*s = out
	return nil
}

/*
findmnt -n -J --target /olares

	{
	   "filesystems": [
	      {
	         "target": "/olares",
	         "source": "/dev/mapper/olares--vg-data[/olares]",
	         "fstype": "ext4",
	         "options": "rw,relatime"
	      }
	   ]
	}
*/
type Filesystem struct {
	Target  string `json:"target"`
	Source  string `json:"source"`
	Fstype  string `json:"fstype"`
	Options string `json:"options"`
}

type FindMntResult struct {
	Filesystems []Filesystem `json:"filesystems"`
}

const FINDMNT = "findmnt"

func CommandFindMnt() *command[FindMntResult] {
	cmd := findCmd(FINDMNT)
	return &command[FindMntResult]{
		cmd:         cmd,
		defaultArgs: []string{"-J"},
		format: func(data []byte) (FindMntResult, error) {
			var res FindMntResult
			err := json.Unmarshal(data, &res)
			return res, err
		},
	}
}
