package types

import (
	"github.com/docker/docker/api/types/strslice"
	"github.com/zishang520/socket.io/socket"
)

type Subscriber struct {
	ID         string
	Connection *socket.Socket
}

type Record = map[string]any

type SubscribeParams struct {
	Id string `json:"id"`
}

type CommonRequestParams struct {
	ID string `uri:"id" binding:"required"`
}
type ContainerRequestParams struct {
	CommonRequestParams
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
	ID   string            `uri:"id" binding:"required"`
	Name strslice.StrSlice `json:"name omitempty"`
}
type ContainerCreateParams struct {
	Image          string            `binding:"required"`
	Name           string            `json:"name default=''"`
	Cmd            strslice.StrSlice `json:"cmd omitempty"`
	Tty            bool              `json:"tty"`
	Stdin          bool              `json:"stdin"`
	Stdout         bool              `json:"stdout"`
	Stderr         bool              `json:"stderr"`
	Detach         bool              `json:"detach"`
	Interactive    bool              `json:"interactive"`
	WorkingDir     string            `json:"working_dir omitempty"`
	User           string            `json:"user omitempty"`
	Entrypoint     strslice.StrSlice `json:"entrypoint omitempty"`
	Autoremove     bool              `json:"auto_remove"`
	ExposeAllPorts bool              `json:"expose_all_ports"`
	Env            strslice.StrSlice `json:"env omitempty"`
	Shell          strslice.StrSlice `json:"shell omitempty"`
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
type ContainerExecCommandParams struct {
	ID  string `uri:"id" binding:"required"`
	Cmd string `json:"cmd"`
}
type ContainerExecBody struct {
	Container   string `json:"container omitempty"`
	Cmd         string `json:"cmd omitempty"`
	Stdout      bool   `json:"stdout"`
	Stdin       bool   `json:"stdin"`
	Stderr      bool   `json:"stderr"`
	Tty         bool   `json:"tty"`
	Detach      bool   `json:"detach"`
	Interactive bool   `json:"interactive"`
	WorkingDir  string `json:"working_dir omitempty"`
	Env         string `json:"env omitempty"`
	Privileged  bool   `json:"privileged"`
	User        string `json:"user"`
}
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
