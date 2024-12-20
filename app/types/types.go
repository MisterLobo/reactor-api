package types

import (
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/zishang520/socket.io/socket"
)

type Subscriber struct {
	ID         string
	Connection *socket.Socket
}

type Record = map[string]any
type ContainerLsResponseBody struct {
	Folders     []ContainerFileInfo `json:"folders"`
	Files       []ContainerFileInfo `json:"files"`
	FolderCount int                 `json:"folder_count"`
	FileCount   int                 `json:"file_count"`
}

type SubscribeParams struct {
	Id string `json:"id"`
}

type CommonRequestParams struct {
	ID string `uri:"id" binding:"required"`
}
type ContainerRequestParams struct {
	CommonRequestParams
}
type ContainerGetArchiveBody struct {
	SrcPath string `json:"src_path" binding:"required"`
}
type ContainerPutArchiveBody struct {
	DstPath string `json:"dst_path" binding:"required"`
}
type ImageRequestParams struct {
	CommonRequestParams
}
type VolumeRequestParams struct {
	CommonRequestParams
}
type NetworkRequestParams struct {
	CommonRequestParams
}
type ContainerListQueryParams struct {
	All bool `form:"all"`
}
type ContainerGetParams struct {
	ID string `uri:"id"`
	// Name string `json:"name omitempty"`
}
type ContainerCreateParams struct {
	Image          string                `binding:"required"`
	Name           string                `json:"name,default=''"`
	Cmd            strslice.StrSlice     `json:"cmd,omitempty"`
	Tty            bool                  `json:"tty"`
	Stdin          bool                  `json:"stdin"`
	Stdout         bool                  `json:"stdout"`
	Stderr         bool                  `json:"stderr"`
	Detach         bool                  `json:"detach"`
	Interactive    bool                  `json:"interactive"`
	WorkingDir     string                `json:"working_dir,omitempty"`
	User           string                `json:"user,omitempty"`
	Entrypoint     strslice.StrSlice     `json:"entrypoint,omitempty"`
	Autoremove     bool                  `json:"auto_remove"`
	ExposeAllPorts bool                  `json:"expose_all_ports"`
	Env            strslice.StrSlice     `json:"env,omitempty"`
	Shell          strslice.StrSlice     `json:"shell,omitempty"`
	Volumes        map[string]struct{}   `json:"volumes,omitempty"`
	Binds          strslice.StrSlice     `json:"binds,omitempty"`
	PortBindings   nat.PortMap           `json:"port_bindings,omitempty"`
	ExposedPorts   map[nat.Port]struct{} `json:"exposed_ports,omitempty"`
	Explorable     bool                  `json:"explorable"`
	SSH            bool                  `json:"ssh"`
}
type ContainerHostInfoQueryParams struct {
	Hostname bool `form:"hostname"`
	Username bool `form:"user"`
	IP       bool `form:"ip"`
	SSH      bool `form:"ssh"`
	LS       bool `form:"ls"`
}
type ContainerHostInfoQueryResponse struct {
	Hostname string          `json:"hostname,omitempty"`
	User     string          `json:"user,omitempty"`
	IP       string          `json:"ip,omitempty"`
	SSH      nat.PortBinding `json:"ssh,omitempty"`
	LS       string          `json:"ls,omitempty"`
}
type ContainerRunParams struct {
	Image  string `binding:"required"`
	Name   string
	Cmd    string
	Tty    bool
	Stdin  bool
	Stdout bool
	Stderr bool
	Detach bool
}
type ContainerStartParams struct{}
type ContainerStopParams struct{}
type ContainerRenameParams struct {
	NewName string `json:"new_name"`
}
type ContainerTopParams struct {
	CommonRequestParams
}
type ContainerStatsParams struct {
	CommonRequestParams
}
type ContainerDiffParams struct {
	CommonRequestParams
}
type ContainerExportParams struct {
	ID string `uri:"id" binding:"required"`
}
type ContainerLogsParams struct {
	CommonRequestParams
}
type ContainerLogsQuery struct {
	Tail       string `uri:"tail omitempty"`
	Follow     bool   `uri:"follow omitempty"`
	Since      string `uri:"since omitempty"`
	Until      string `uri:"until omitempty"`
	Timestamps bool   `uri:"timestamps omitempty"`
	Details    bool   `uri:"details omitempty"`
	ShowStdout bool   `uri:"stdout omitempty"`
	ShowStderr bool   `uri:"stderr omitempty"`
}
type ContainerRemoveParams struct {
	CommonRequestParams
	Force bool `json:"force"`
}
type ContainerExecParams struct {
	ID string `uri:"id" binding:"required"`
}
type ContainerLsParams struct {
	ID string `uri:"id" binding:"required"`
}
type ContainerExecCommandParams struct {
	ID  string `uri:"id" binding:"required"`
	Cmd string `json:"cmd"`
}
type ContainerExecBody struct {
	Container   string `json:"container"`
	Cmd         string `json:"cmd"`
	Stdout      bool   `json:"stdout"`
	Stdin       bool   `json:"stdin"`
	Stderr      bool   `json:"stderr"`
	Tty         bool   `json:"tty"`
	Detach      bool   `json:"detach"`
	Interactive bool   `json:"interactive"`
	WorkingDir  string `json:"working_dir"`
	Env         string `json:"env"`
	Privileged  bool   `json:"privileged"`
	User        string `json:"user"`
}
type ContainerExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}
type ContainerLsBody struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

/*
	 type ContainerFileInfo struct {
		Name     string `json:"name"`
		MimeType string `json:"mime_type"`
		Size     string `json:"size"`
		Attrs    string `json:"attrs"`
		Owner    string `json:"owner"`
		Group    string `json:"group"`
		Created  string `json:"created"`
		Modified string `json:"modified"`
	}
*/
type ContainerAttachParams struct {
	CommonRequestParams
	Stdout bool `json:"stdout"`
	Stderr bool `json:"stderr"`
	Stdin  bool `json:"stdin"`
	Stream bool `json:"stream"`
	Logs   bool `json:"logs"`
}
type ContainerSummary struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Command string `json:"command"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Created string `json:"created"`
	Image   string `json:"image"`
}
type ImageSummary struct {
	ID      string `json:"id"`
	Repo    string `json:"repo"`
	Created string `json:"created"`
	Size    string `json:"size"`
}
type ImagePullParams struct {
	Repo string `json:"repo" binding:"required"`
	Tag  string `json:"tag"`
}
type ImagePullProgress struct {
	Status string `json:"status"`
}
type ImageBuildParams struct {
	Tag string `form:"tag omitempty"`
}
type VolumeSummary struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Created    string `json:"created"`
	MountPoint string `json:"mount_point"`
}
type NetworkSummary struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Created string   `json:"created"`
	Ports   []string `json:"ports"`
}

type ContainerActionResult struct {
	ID        string
	Container *ContainerSummary
}

type ConnectionCreateParams struct {
	Name      string `json:"id"`
	Type      string `json:"socket_type"`
	Address   string `json:"socket_address"`
	IsDefault bool   `json:"is_default"`
}
type ConnectionUpdateParams struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"socket_type"`
	Address   string `json:"socket_address"`
	IsDefault bool   `json:"is_default"`
}
type ContainerRunResponse struct {
	ID     string
	Name   string
	Status string
	Error  string
}
type ConnectionTestParams struct {
	Connection string `json:"connection"`
	Exact      bool   `json:"exact"`
}
type PortBinding struct {
	*nat.PortBinding
}

type PullEvent struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	Progress       string `json:"progress,omitempty"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

type ContainerFileInfo struct {
	Name     string `json:"name"`
	Type     string `json:"file_type"`
	Size     int    `json:"size"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
	Perm     string `json:"perm"`
	Owner    string `json:"owner"`
	Group    string `json:"group"`
	IsDir    bool   `json:"is_dir"`
	IsLink   bool   `json:"is_link"`
	RealPath string `json:"real_path"`
	ID       string `json:"id"`
}
