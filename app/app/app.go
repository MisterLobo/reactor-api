package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reactor/models"
	"reactor/types"
	"reactor/utils"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/zishang520/socket.io/socket"
)

var app *App

type App struct {
	connectionManager  *models.ConnectionManager
	configManager      *utils.ConfigurationManager
	client             *client.Client
	SocketServer       *socket.Server
	Logger             *log.Logger
	Subscribers        map[string]*types.Subscriber
	AttachedContainers map[string]*dockertypes.HijackedResponse
	AttachedExecs      map[string]*dockertypes.HijackedResponse
}

func DefaultApp() *App {
	if app == nil {
		app = &App{}
	}
	return app
}

func (app *App) GetSub(id string) (bool, *types.Subscriber) {
	if app.Subscribers == nil {
		return false, nil
	}
	sub := app.Subscribers[id]
	if sub == nil {
		return false, nil
	}
	c := sub.Connection
	if c == nil {
		return false, nil
	}
	return true, sub
}

func (app *App) Setup() {
	app.beforeInitHooks()

	app.setupDockerClient()

	app.afterInitHooks()
}
func (app *App) setupDockerClient() {
	app.AttachedContainers = make(map[string]*dockertypes.HijackedResponse)
	app.AttachedExecs = make(map[string]*dockertypes.HijackedResponse)

	cm := app.connectionManager
	_, ds := cm.GetDefaultConnection()
	app.Logger.Println("DOCKER HOST: ", ds)

	apiClient, err := client.NewClientWithOpts(client.WithHost(ds), client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println("Could connect to the Docker daemon:", err)
		panic(err)
	}
	ping, err := apiClient.Ping(context.Background())
	if err != nil {
		fmt.Println("[ERROR]:", err.Error())
		ok, sub := app.GetSub("sub")
		fmt.Println("[sub]:", sub != nil)
		if ok {
			fmt.Println("sending error message")
			cc := sub.Connection
			cc.Emit("apierror", err.Error(), fmt.Sprintf("%v", http.StatusInternalServerError))
		}
	}
	app.Logger.Println("Docker Client version: ", ping.APIVersion)
	app.client = apiClient
	defer apiClient.Close()
}
func (app *App) setupSocketServer() {
	app.Subscribers = map[string]*types.Subscriber{}
	ss := socket.NewServer(nil, nil)
	ss.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		fmt.Println("[newclient]: ", string(client.Id()), client.Nsp().Name())
		client.On("message", func(args ...any) {})
	})
	ss.On("disconnection", func(clients ...any) {
		fmt.Println("client has disconnected")
	})
	ss.Of("/sub", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("subscribe", func(args ...any) {
			params := args[0].(types.Record)
			id := params["id"].(string)
			room := socket.Room(id[:8])
			fmt.Println("[subscriber]:", room, id)
			fmt.Println("[rooms]:", client.Rooms().Keys())
			client.Join(room)
			app.Subscribers[id] = &types.Subscriber{
				ID:         string(client.Id()),
				Connection: client,
			}
			client.Emit("subbed", id, client.Id())
		})
		client.On("start", func(a ...any) {
			client.Emit("started", "1", "1")
		})
		client.On("stop", func(a ...any) {
			client.Emit("stopped", "1", "1")
		})
		client.On("status", func(args ...any) {
			arg := args[0].(types.Record)
			id := arg["id"].(string)
			fmt.Println("checking status for container: ", id)
			params := types.ContainerRequestParams{}
			params.ID = id
			j, _ := app.ContainerInspect(&params)
			client.Emit("status", id, j.State.Status)
		})
	})
	ss.Of("/containers", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("subscribe", func(args ...any) {
			params := args[0].(types.Record)
			id := params["id"].(string)
			room := socket.Room(id[:8])
			client.Join(room)
			fmt.Println("[rooms]:", client.Rooms().Keys())
			app.Subscribers[id] = &types.Subscriber{
				ID:         string(client.Id()),
				Connection: client,
			}
			client.Emit("subbed", id, client.Id())
		})
	})
	ss.Of("/images", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("subscribe", func(args ...any) {
			params := args[0].(types.Record)
			id := params["id"].(string)
			room := socket.Room(id[:8])
			client.Join(room)
			fmt.Println("[rooms]:", client.Rooms().Keys())
			app.Subscribers[id] = &types.Subscriber{
				ID:         string(client.Id()),
				Connection: client,
			}
			client.Emit("subbed", id, client.Id())
		})
	})
	ss.Of("/networks", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("subscribe", func(args ...any) {
			params := args[0].(types.Record)
			id := params["id"].(string)
			room := socket.Room(id[:8])
			client.Join(room)
			fmt.Println("[rooms]:", client.Rooms().Keys())
			app.Subscribers[id] = &types.Subscriber{
				ID:         string(client.Id()),
				Connection: client,
			}
			client.Emit("subbed", id, client.Id())
		})
	})
	ss.Of("/volumes", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("subscribe", func(args ...any) {
			params := args[0].(types.Record)
			id := params["id"].(string)
			room := socket.Room(id[:8])
			client.Join(room)
			fmt.Println("[rooms]:", client.Rooms().Keys())
			app.Subscribers[id] = &types.Subscriber{
				ID:         string(client.Id()),
				Connection: client,
			}
			client.Emit("subbed", id, client.Id())
		})
	})
	app.SocketServer = ss
}
func (app *App) beforeInitHooks() {
	app.Subscribers = map[string]*types.Subscriber{}

	app.initLogger()
	app.initDefaultSettings()
	app.initDb()
}
func (app *App) afterInitHooks() {
	go app.SetupDaemonEventListeners()
	app.SetupAppEventListeners()
}
func (app *App) initDefaultSettings() {
	app.configManager = utils.DefaultConfigurationManager()
	app.configManager.InitDefaults()
}
func (app *App) initLogger() {
	logger := log.New(os.Stdout, fmt.Sprintf("[%s]: ", utils.APP_NAME), log.LstdFlags)
	logger.Println("setting up logger")
	app.Logger = logger
}
func (app *App) initDb() {
	app.connectionManager = models.DefaultConnectionManager()
	app.connectionManager.InitDefaults()
}
func (app *App) SetupAppEventListeners() {}
func (app *App) SetupDaemonEventListeners() {
	cli := app.client
	ping, err := cli.Ping(context.Background())
	if err != nil {
		fmt.Println("[ping#error]:", err.Error())
		ok, sub := app.GetSub("sub")
		fmt.Println("[sub]:", sub != nil)
		if ok {
			fmt.Println("sending error message")
			cc := sub.Connection
			cc.Emit("apierror", err.Error(), fmt.Sprintf("%v", http.StatusInternalServerError))
		}
		return
	}
	fmt.Println("[ping]:", ping.APIVersion, ping.OSType)
	chn, _ := cli.Events(context.Background(), events.ListOptions{})
	for {
		msg := <-chn
		fmt.Println("[event]:", msg)
		id := msg.Actor.ID
		switch msg.Type {
		case "container":
			switch msg.Action {
			case "start":
				j, err := cli.ContainerInspect(context.Background(), id)
				if err != nil {
					fmt.Println("[daemon] error:", err.Error())
					continue
				}
				sum := app.ContainerGet(&types.ContainerGetParams{ID: id})
				f, s := app.GetSub(id)
				if f {
					fmt.Println("[started]:", id, j.Name, j.State.Status)
					c := s.Connection
					fmt.Println("rooms:", c.Rooms().Keys())
					if sum != nil {
						c.Broadcast().Emit("started", j.Name, id, sum.State, sum.Status)
					}
				} else {
					fmt.Println("[start] no connection found for", id)
				}
				ff, ss := app.GetSub("sub")
				fmt.Println("[sub] found:", ff, sum.ID)
				if ff {
					cc := ss.Connection
					if sum != nil {
						cc.Broadcast().Emit("started", j.Name, "sub")
					}
				} else {
					fmt.Println("[sub#start] no connection found for 'sub'")
				}
				break
			case "die":
				j, err := cli.ContainerInspect(context.Background(), id)
				if err != nil {
					fmt.Println("[daemon] error:", err.Error())
					continue
				}
				sum := app.ContainerGet(&types.ContainerGetParams{ID: id})
				f, s := app.GetSub(id)
				if f {
					fmt.Println("[stopped]:", id, j.Name, j.State.Status)
					c := s.Connection
					if sum != nil {
						c.Broadcast().Emit("stopped", j.Name, id, sum.State, sum.Status)
					} else {
						fmt.Println("[stopped] no container found with id: ", id)
					}
				} else {
					fmt.Println("[die] no connection found for", id)
				}
				ff, ss := app.GetSub("sub")
				if ff {
					cc := ss.Connection
					if sum != nil {
						cc.Broadcast().Emit("stopped", j.Name, "sub")
					}
				} else {
					fmt.Println("[sub#stop] no connection found for 'sub'")
				}
				break
			case "kill":
				j, err := cli.ContainerInspect(context.Background(), id)
				if err != nil {
					fmt.Println("[daemon] error:", err.Error())
					continue
				}
				sum := app.ContainerGet(&types.ContainerGetParams{ID: id})
				f, s := app.GetSub(id)
				if f {
					fmt.Println("[killed]:", j.Name, id, j.State.Status)
					c := s.Connection
					c.Broadcast().Emit("killed", j.Name, id, sum.State, sum.Status)
				} else {
					fmt.Println("[kill] no connection found for", id)
				}
				ff, ss := app.GetSub("sub")
				if ff {
					cc := ss.Connection
					if sum != nil {
						cc.Broadcast().Emit("killed", j.Name, "sub")
					}
				} else {
					fmt.Println("[sub#kill] no connection found for 'sub'")
				}
				break
			case "restart":
				j, err := cli.ContainerInspect(context.Background(), id)
				if err != nil {
					fmt.Println("[daemon] error:", err.Error())
					continue
				}
				sum := app.ContainerGet(&types.ContainerGetParams{ID: id})
				f, s := app.GetSub(id)
				if f {
					fmt.Println("[restarted]:", j.Name, id, j.State.Status)
					c := s.Connection
					c.Broadcast().Emit("restarted", j.Name, id, j.State.Status)
				} else {
					fmt.Println("[restart] no connection found for", id)
				}
				ff, ss := app.GetSub("sub")
				if ff {
					cc := ss.Connection
					if sum != nil {
						cc.Broadcast().Emit("restarted", j.Name, "sub")
					}
				} else {
					fmt.Println("[sub#restart] no connection found for 'sub'")
				}
				break
			case "pause":
				j, err := cli.ContainerInspect(context.Background(), id)
				if err != nil {
					fmt.Println("[daemon] error:", err.Error())
					continue
				}
				sum := app.ContainerGet(&types.ContainerGetParams{ID: id})
				f, s := app.GetSub(id)
				if f {
					fmt.Println("[paused]:", j.Name, id, j.State.Status)
					c := s.Connection
					c.Broadcast().Emit("paused", j.Name, id, j.State.Status)
				} else {
					fmt.Println("[pause] no connection found for", id)
				}
				ff, ss := app.GetSub("sub")
				if ff {
					cc := ss.Connection
					if sum != nil {
						cc.Broadcast().Emit("paused", j.Name, "sub")
					}
				} else {
					fmt.Println("[sub#pause] no connection found for 'sub'")
				}
				break
			case "unpause":
				j, err := cli.ContainerInspect(context.Background(), id)
				if err != nil {
					fmt.Println("[daemon] error:", err.Error())
					continue
				}
				sum := app.ContainerGet(&types.ContainerGetParams{ID: id})
				f, s := app.GetSub(id)
				if f {
					fmt.Println("[unpaused]:", j.Name, id, j.State.Status)
					c := s.Connection
					c.Broadcast().Emit("unpaused", j.Name, id, j.State.Status)
					c.Broadcast().Emit("sub_unpaused", j.Name, "sub")
				} else {
					fmt.Println("[unpause] no connection found for", id)
				}
				ff, ss := app.GetSub("sub")
				if ff {
					cc := ss.Connection
					if sum != nil {
						cc.Broadcast().Emit("unpaused", j.Name, "sub")
					}
				} else {
					fmt.Println("[sub#unpause] no connection found for 'sub'")
				}
				break
			case "rename":
				j, err := cli.ContainerInspect(context.Background(), id)
				if err != nil {
					fmt.Println("[daemon] error:", err.Error())
					continue
				}
				sum := app.ContainerGet(&types.ContainerGetParams{ID: id})
				f, s := app.GetSub(id)
				if f {
					fmt.Println("[rename]:", id, j.Name, j.State.Status)
					fmt.Println("[rename] sending:", id, j.Name)
					c := s.Connection
					c.Broadcast().Emit("renamed", j.Name, id)
					c.Broadcast().Emit("sub_renamed", j.Name, "sub")
				} else {
					fmt.Println("[rename] no connection found for ", id)
				}
				ff, ss := app.GetSub("sub")
				if ff {
					cc := ss.Connection
					if sum != nil {
						cc.Broadcast().Emit("renamed", j.Name, "sub")
					}
				} else {
					fmt.Println("[sub#rename] no connection found for 'sub'")
				}
				break
			case "destroy":
				j, err := cli.ContainerInspect(context.Background(), id)
				if err != nil {
					fmt.Println("[daemon] error:", err.Error())
					continue
				}
				sum := app.ContainerGet(&types.ContainerGetParams{ID: id})
				f, s := app.GetSub(id)
				if f {
					name := msg.Actor.Attributes["name"]
					fmt.Println("[destroy]:", id, name)
					c := s.Connection
					c.Broadcast().Emit("removed", name, id)
					c.Broadcast().Emit("sub_removed", j.Name, "sub")
				} else {
					fmt.Println("[destroy] no connection found for ", id)
				}
				ff, ss := app.GetSub("sub")
				if ff {
					cc := ss.Connection
					if sum != nil {
						cc.Broadcast().Emit("removed", j.Name, "sub")
					}
				} else {
					fmt.Println("[destroy] no connection found for 'sub'")
				}
				break
			}
		case "image":
			switch msg.Action {
			case "pull":
				f, s := app.GetSub("sub")
				if f {
					c := s.Connection
					c.Broadcast().Emit("pulled")
				}
				break
			}
		}
	}
}

func (app *App) ContainerListRunning() []*types.ContainerSummary {
	cli := app.client
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: false})
	csummary := make([]*types.ContainerSummary, 0)
	if err != nil {
		return csummary
	}
	for _, container := range containers {
		csummary = append(csummary, &types.ContainerSummary{
			ID:      container.ID,
			Name:    strings.ReplaceAll(container.Names[0], "/", ""),
			Command: container.Command,
			Image:   container.Image,
			Created: time.Unix(container.Created, 0).Format(time.UnixDate),
			State:   container.State,
			Status:  container.Status,
		})
	}
	return csummary
}
func (app *App) ContainerList() []*types.ContainerSummary {
	cli := app.client
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: false,
		Filters: filters.NewArgs(
			filters.Arg("status", "created"),
			filters.Arg("status", "exited"),
			filters.Arg("status", "dead"),
			filters.Arg("status", "removing"),
			filters.Arg("status", "restarting"),
			filters.Arg("status", "paused"),
		),
	})
	csummary := make([]*types.ContainerSummary, 0)
	if err != nil {
		return csummary
	}
	for _, container := range containers {
		csummary = append(csummary, &types.ContainerSummary{
			ID:      container.ID,
			Name:    strings.ReplaceAll(container.Names[0], "/", ""),
			Command: container.Command,
			Image:   container.Image,
			Created: time.Unix(container.Created, 0).Format(time.UnixDate),
			State:   container.State,
			Status:  container.Status,
		})
	}
	return csummary
}
func (app *App) ContainerGet(params *types.ContainerGetParams) *types.ContainerSummary {
	cli := app.client
	filterKV := make([]filters.KeyValuePair, 0)
	filterKV = append(filterKV, filters.Arg("id", params.ID))
	nameb := make([]byte, 0)
	if params.Name != nil {
		params.Name.UnmarshalJSON(nameb)
		filterKV = append(filterKV, filters.Arg("name", string(nameb)))
	}
	args := filters.NewArgs(filterKV...)
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return nil
	}
	if len(containers) == 0 {
		return nil
	}
	container := containers[0]
	csummary := &types.ContainerSummary{
		ID:      container.ID,
		Name:    container.Names[0],
		Command: container.Command,
		Image:   container.Image,
		Created: time.Unix(container.Created, 0).Format(time.UnixDate),
		State:   container.State,
		Status:  container.Status,
	}
	return csummary
}
func (app *App) ImageList() []types.ImageSummary {
	cli := app.client
	images, err := cli.ImageList(context.Background(), image.ListOptions{})
	isummary := make([]types.ImageSummary, 0)
	if err != nil {
		return isummary
	}
	for _, image := range images {
		tags := image.RepoTags
		id, _, _ := strings.Cut(image.ID, ":")
		repo := id
		if len(tags) > 0 {
			repo = tags[0]
		} else {
			repo = fmt.Sprintf("%s:%s", id, "latest")
		}
		isummary = append(isummary, types.ImageSummary{
			ID:      image.ID,
			Repo:    repo,
			Created: time.Unix(image.Created, 0).Format(time.UnixDate),
			Size:    humanize.IBytes(uint64(image.Size)),
		})
	}
	return isummary
}
func (app *App) VolumeList() []types.VolumeSummary {
	cli := app.client
	volumes, err := cli.VolumeList(context.Background(), volume.ListOptions{})
	vsummary := make([]types.VolumeSummary, 0)
	if err != nil {
		return vsummary
	}
	for _, volume := range volumes.Volumes {
		vsummary = append(vsummary, types.VolumeSummary{
			ID:         volume.Name,
			Name:       volume.Name,
			Created:    volume.CreatedAt,
			MountPoint: volume.Mountpoint,
		})
	}
	return vsummary
}
func (app *App) NetworkList() []types.NetworkSummary {
	cli := app.client
	networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
	nsummary := make([]types.NetworkSummary, 0)
	if err != nil {
		return nsummary
	}
	for _, network := range networks {
		nsummary = append(nsummary, types.NetworkSummary{
			ID:      network.ID,
			Name:    network.Name,
			Created: network.Created.Format(time.UnixDate),
			Ports:   []string{":80/tcp"},
		})
	}
	return nsummary
}

func (app *App) ContainerCreate(params *types.ContainerCreateParams) (*types.ContainerActionResult, error) {
	img, _, err := app.client.ImageInspectWithRaw(context.Background(), params.Image)
	if err != nil {
		return nil, err
	}
	cfg := img.Config
	cfg.Entrypoint = nil
	cfg.OpenStdin = true
	cfg.StdinOnce = true
	cmd := params.Cmd
	if cmd != nil {
		cfg.Cmd = cmd
	}
	entrypoint := params.Entrypoint
	if len(entrypoint) > 0 {
		// cfg.Entrypoint = entrypoint
	}
	name := params.Name
	wd := params.WorkingDir
	if wd != "" {
		cfg.WorkingDir = wd
	}
	env := params.Env
	if env != nil {
		cfg.Env = env
	}

	res, err := app.client.ContainerCreate(context.Background(), &container.Config{
		Image:        params.Image,
		AttachStdin:  params.Stdin,
		AttachStdout: params.Stdout,
		AttachStderr: params.Stderr,
		Tty:          params.Tty,
		OpenStdin:    params.Interactive,
		Cmd:          params.Cmd,
		Env:          env,
		Hostname:     params.Name,
		User:         params.User,
		Shell:        params.Shell,
		WorkingDir:   params.WorkingDir,
	}, nil, nil, nil, name)
	if err != nil {
		return nil, err
	}
	return &types.ContainerActionResult{
		ID:        res.ID,
		Container: nil,
	}, nil
}
func (app *App) ContainerRun(params *types.ContainerCreateParams) (*types.ContainerRunResponse, error) {
	res, err := app.ContainerCreate(params)
	if err != nil {
		return nil, err
	}
	/* err = app.client.ContainerRename(context.Background(), res.ID, params.Name)
	if err != nil {
		return nil, err
	} */
	go func() {
		err = app.client.ContainerStart(context.Background(), res.ID, container.StartOptions{})
		if err != nil {
			fmt.Println("error running container:", err.Error())
		}
	}()
	return &types.ContainerRunResponse{
		ID: res.ID,
	}, nil
}
func (app *App) ContainerStart(params *types.ContainerRequestParams) error {
	err := app.client.ContainerStart(context.Background(), params.ID, container.StartOptions{})
	return err
}
func (app *App) ContainerStop(params *types.ContainerRequestParams) error {
	noWaitTimeout := 0
	err := app.client.ContainerStop(context.Background(), params.ID, container.StopOptions{Timeout: &noWaitTimeout})
	return err
}
func (app *App) ContainerRestart(params *types.ContainerRequestParams) error {
	err := app.client.ContainerRestart(context.Background(), params.ID, container.StopOptions{})
	return err
}
func (app *App) ContainerKill(params *types.ContainerRequestParams) error {
	err := app.client.ContainerKill(context.Background(), params.ID, "SIGKILL")
	return err
}
func (app *App) ContainerPause(params *types.ContainerRequestParams) error {
	err := app.client.ContainerPause(context.Background(), params.ID)
	return err
}
func (app *App) ContainerUnpause(params *types.ContainerRequestParams) error {
	err := app.client.ContainerUnpause(context.Background(), params.ID)
	return err
}
func (app *App) ContainerExport(params *types.ContainerExportParams) ([]byte, error) {
	rc, err := app.client.ContainerExport(context.Background(), params.ID)
	buf := make([]byte, 0)
	br := bytes.NewBuffer(buf)
	io.Copy(br, rc)
	return br.Bytes(), err
}
func (app *App) ContainerDiff(params *types.ContainerDiffParams) ([]container.FilesystemChange, error) {
	diff, err := app.client.ContainerDiff(context.Background(), params.ID)
	return diff, err
}
func (app *App) ContainerTop(params *types.ContainerTopParams) (container.ContainerTopOKBody, error) {
	top, err := app.client.ContainerTop(context.Background(), params.ID, []string{})
	return top, err
}
func (app *App) ContainerStats(params *types.ContainerStatsParams) (string, error) {
	res, err := app.client.ContainerStatsOneShot(context.Background(), params.ID)
	if err != nil {
		fmt.Println("[stats#error]:", err.Error())
	}
	buf := new(strings.Builder)
	n, err := io.Copy(buf, res.Body)
	body := buf.String()
	fmt.Println("[stats]:", n, body)
	return body, err
}
func (app *App) ContainerPutArchive(params *types.ContainerRequestParams) {}
func (app *App) ContainerGetArchive(params *types.ContainerRequestParams) {}
func (app *App) ContainerRename(params *types.ContainerRequestParams, body *types.ContainerRenameParams) error {
	err := app.client.ContainerRename(context.Background(), params.ID, body.NewName)
	return err
}
func (app *App) ContainerRemove(params *types.ContainerRemoveParams) error {
	err := app.client.ContainerRemove(context.Background(), params.ID, container.RemoveOptions{})
	return err
}
func (app *App) ContainerInspect(params *types.ContainerRequestParams) (*dockertypes.ContainerJSON, error) {
	cjson, err := app.client.ContainerInspect(context.Background(), params.ID)
	if err != nil {
		return nil, err
	}
	return &cjson, nil
}
func (app *App) ContainerAttach(params *types.ContainerAttachParams) error {
	hj, err := app.client.ContainerAttach(context.Background(), params.ID, container.AttachOptions{
		Stream: params.Stream,
		Stdin:  params.Stdin,
		Stdout: params.Stdout,
		Stderr: params.Stderr,
		Logs:   params.Logs,
	})
	app.AttachedContainers[params.ID] = &hj
	return err
}
func (app *App) ContainerLogs(params *types.ContainerLogsParams, opts *types.ContainerLogsQuery) (string, error) {
	r, err := app.client.ContainerLogs(context.Background(), params.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		/* ShowStdout: params.ShowStdout,
		ShowStderr: params.ShowStderr,
		Details:    params.Details,
		Timestamps: params.Timestamps,
		Tail:       params.Tail,
		Follow:     params.Follow,
		Since:      params.Since,
		Until:      params.Until, */
	})
	if err != nil {
		return "", err
	}
	buf := bytes.NewBufferString("")
	_, err = io.Copy(buf, r)
	logs := buf.String()
	return logs, err
}
func (app *App) ContainerTerminal(ctx *gin.Context) {}
func (app *App) ContainerExec(params *types.ContainerExecParams, body *types.ContainerExecBody) error {
	cmd := strings.Split(body.Cmd, " ")
	env := strings.Split(body.Env, " ")
	exec, err := app.client.ContainerExecCreate(context.Background(), params.ID, container.ExecOptions{
		Tty:          body.Tty,
		AttachStdout: body.Stdout,
		AttachStdin:  body.Stdin,
		AttachStderr: body.Stderr,
		Detach:       body.Detach,
		Cmd:          cmd,
		Env:          env,
		User:         body.User,
		WorkingDir:   body.WorkingDir,
		Privileged:   body.Privileged,
	})
	if err != nil {
		fmt.Println("[exec]:", err.Error())
		return err
	}
	err = app.client.ContainerExecStart(context.Background(), exec.ID, container.ExecStartOptions{
		// Detach: body.Detach,
		Tty: body.Tty,
	})
	if err != nil {
		fmt.Println("[execstart]:", err.Error())
		return err
	}
	keepAlive := body.Tty && body.Stdin
	if keepAlive {
		hj, err := app.client.ContainerExecAttach(context.Background(), exec.ID, container.ExecStartOptions{})
		if err != nil {
			return err
		}
		fmt.Println("connection established:", params.ID)
		attachedExec := app.AttachedExecs[params.ID]
		if attachedExec != nil {
			attachedExec.Close()
		}
		app.AttachedExecs[params.ID] = &hj

		buf := bytes.Buffer{}
		go func() {
			defer hj.Conn.Close()

			mw := io.MultiWriter(&buf, os.Stdout)
			log.SetOutput(mw)
			_, err := stdcopy.StdCopy(log.Writer(), os.Stderr, hj.Reader)
			if err != nil {
				fmt.Println("[EXECERROR]:", err.Error())
			}
		}()
	}

	return nil
}
func (app *App) ContainerExecCommand(params *types.ContainerExecCommandParams) error {
	hj := app.AttachedExecs[params.ID]
	if hj == nil {
		return dockertypes.ErrorResponse{
			Message: "container is not connected",
		}
	}
	hj.Conn.Write([]byte(params.Cmd))
	return nil
}
func (app *App) ContainersPrune() {
	app.client.ContainersPrune(context.Background(), filters.NewArgs(filters.Arg("", "")))
}

func (app *App) ImagePull(params *types.ImagePullParams) (string, error) {
	rc, err := app.client.ImagePull(context.Background(), params.Repo, image.PullOptions{})
	if err != nil {
		return "", err
	}
	buf := bytes.NewBufferString("")
	io.Copy(buf, rc)
	return buf.String(), nil
}
func (app *App) ImageCreate(ref string) error {
	_, err := app.client.ImageCreate(context.Background(), ref, image.CreateOptions{})
	return err
}
func (app *App) ImageBuild(src string, tags ...string) error {
	fmt.Println("reading archive file...")
	ff, err := os.Open(src)
	if err != nil {
		fmt.Println("error reading file: ", err)
		return err
	}
	defer ff.Close()

	fmt.Println("initiating build instance")
	_, err = app.client.ImageBuild(context.Background(), ff, dockertypes.ImageBuildOptions{
		Tags: tags,
	})
	if err != nil {
		fmt.Println("error while building image: ", err.Error())
		return err
	}
	fmt.Println("build completed successfully")
	return nil
}
func (app *App) ImageInspect(id string) (*dockertypes.ImageInspect, error) {
	i, _, err := app.client.ImageInspectWithRaw(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (app *App) VolumeInspect(id string) (*volume.Volume, error) {
	v, err := app.client.VolumeInspect(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
func (app *App) NetworkInspect(id string) (*network.Inspect, error) {
	n, err := app.client.NetworkInspect(context.Background(), id, network.InspectOptions{})
	if err != nil {
		return nil, err
	}
	return &n, nil
}
func (app *App) ListConnections() []models.ConnectionConfig {
	return app.connectionManager.ListConnections()
}
func (app *App) SaveConnection(c *models.ConnectionConfig) bool {
	res := app.connectionManager.SaveConnection(c)
	return res == 1
}
func (app *App) UpdateConnection(p *types.CommonRequestParams, b *models.ConnectionConfig) bool {
	res := app.connectionManager.UpdateConnection(b)
	return res == 1
}
func (app *App) SetDefaultConnection(id string) string {
	return app.connectionManager.SetDefaultConnection(id)
}
func (app *App) GetConnection(id string) *models.ConnectionConfig {
	res, k := app.connectionManager.GetConnection(id)
	if k {
		return res
	} else {
		return nil
	}
}
func (app *App) GetDefaultConnection() (*models.ConnectionConfig, string) {
	return app.connectionManager.GetDefaultConnection()
}
func (app *App) DeleteConnection(id string) bool {
	res := app.connectionManager.DeleteConnection(id)
	return res == 1
}
func (app *App) TestConnection(idOrConnStr string, exact bool) (bool, string) {
	if exact {
		return app.connectionManager.TestConnectionString(idOrConnStr)
	}
	return app.connectionManager.TestConnection(idOrConnStr)
}
