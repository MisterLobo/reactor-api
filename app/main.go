package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"reactor/app"
	"reactor/models"
	"reactor/types"
	"syscall"

	ginGzip "github.com/gin-contrib/gzip"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	engineiotypes "github.com/zishang520/engine.io/types"
	"github.com/zishang520/socket.io/socket"
)

func setupSocketServer(app *app.App) *socket.Server {
	app.Subscribers = map[string]*types.Subscriber{}
	ss := socket.NewServer(nil, nil)
	ss.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		fmt.Println("[newclient]: ", string(client.Id()), client.Nsp().Name())
		client.On("error", func(args ...any) {})
	})
	ss.On("disconnect", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		fmt.Println("[disconnected]: ", string(client.Id()), client.Nsp().Name())
	})
	ss.Of("/sub", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("subscribe", func(args ...any) {
			params := args[0].(types.Record)
			id := params["id"].(string)
			// room := socket.Room(id[:8])
			// client.Join(room)
			fmt.Println("[sub#client]:", id)
			fmt.Println("[rooms]:", client.Rooms().Keys())
			app.Subscribers["sub"] = &types.Subscriber{
				ID:         string(client.Id()),
				Connection: client,
			}
			client.Emit("subbed", id, client.Id())
			// client.Disconnect(false)
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
		client.On("ping", func(args ...any) {
			fmt.Println("[ping]:", args)
			arg := args[0].(types.Record)
			fmt.Println("arg:", arg)
			exact := arg["exact"].(bool)
			fmt.Println("exact:", exact)
			if exact && arg["connection_string"] != nil {
				connStr := arg["connection_string"].(*string)
				statusOk, err := app.TestConnection(*connStr, true)
				fmt.Println("[ping#result]:", statusOk, err)
				client.Broadcast().Emit("pong", types.Record{"ok": statusOk, "error": err})
			} else {
				if arg["id"] != nil {
					id := arg["id"].(*string)
					statusOk, err := app.TestConnection(*id, false)
					fmt.Println("[ping#result]:", statusOk, err)
					client.Broadcast().Emit("pong", types.Record{"ok": statusOk, "error": err})
				} else {
					conn, _ := app.GetDefaultConnection()
					statusOk, err := app.TestConnection(conn.ID, false)
					fmt.Println("[ping#result]:", statusOk, err)
					client.Broadcast().Emit("pong", types.Record{"ok": statusOk, "error": err})
				}
			}
		})
		client.On("test", func(args ...any) {
			fmt.Println("[ping]:", args)
			arg := args[0].(types.Record)
			fmt.Println("arg:", arg)
			exact := arg["exact"].(bool)
			fmt.Println("exact:", exact)
			if exact && arg["connection_string"] != nil {
				connStr := arg["connection_string"].(*string)
				statusOk, err := app.TestConnection(*connStr, true)
				fmt.Println("[ping#result]:", statusOk, err)
				client.Broadcast().Emit("pong", types.Record{"ok": statusOk, "error": err})
			} else {
				if arg["id"] != nil {
					id := arg["id"].(*string)
					statusOk, err := app.TestConnection(*id, false)
					fmt.Println("[ping#result]:", statusOk, err)
					client.Broadcast().Emit("pong", types.Record{"ok": statusOk, "error": err})
				} else {
					conn, _ := app.GetDefaultConnection()
					statusOk, err := app.TestConnection(conn.ID, false)
					fmt.Println("[ping#result]:", statusOk, err)
					client.Broadcast().Emit("pong", types.Record{"ok": statusOk, "error": err})
				}
			}
		})
	})
	ss.Of("/container", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("subscribe", func(args ...any) {
			params := args[0].(types.Record)
			id := params["id"].(string)
			room := socket.Room(id[:8])
			client.Join(room)
			fmt.Println("[container#client]:", id)
			fmt.Println("[rooms]:", client.Rooms().Keys())
			app.Subscribers[id] = &types.Subscriber{
				ID:         string(client.Id()),
				Connection: client,
			}
			client.Emit("subbed", id, client.Id())
		})
	})
	ss.Of("/image", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("subscribe_image", func(args ...any) {
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
	ss.Of("/network", func(clients ...any) {
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
	ss.Of("/volume", func(clients ...any) {
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
	return ss
}

func main() {
	r := gin.Default()
	app := app.DefaultApp()
	ss := setupSocketServer(app)

	c := socket.DefaultServerOptions()
	c.SetServeClient(true)
	c.SetCors(&engineiotypes.Cors{
		Origin:      "*",
		Credentials: true,
	})

	if ss != nil {
		app.Setup()
	}
	// ss := app.SocketServer

	r.MaxMultipartMemory = 150 << 20
	r.
		Use(cors.New(cors.Config{
			AllowAllOrigins: true,
			/* AllowHeaders: []string{
				"Origin",
				"Authorization",
				"x-SECRET",
				"Content-Type",
				"Accept-Encoding",
				"Content-Encoding",
			}, */
			// ExposeHeaders:    []string{"Content-Type", "Content-Encoding"},
			AllowCredentials: true,
		})).
		Use(ginGzip.Gzip(ginGzip.BestCompression))

	r.
		GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"error": "pong",
			})
		}).
		GET("/info", func(ctx *gin.Context) {}).
		GET("/version", func(ctx *gin.Context) {})

	r.GET("/socket.io/*any", gin.WrapH(ss.ServeHandler(c)))
	r.POST("/socket.io/*any", gin.WrapH(ss.ServeHandler(c)))

	r.
		GET("/connections", func(ctx *gin.Context) {
			list := app.ListConnections()
			fmt.Println("list", list)
			ctx.JSON(http.StatusOK, gin.H{"list": list})
		}).
		POST("/connections", func(ctx *gin.Context) {
			var body models.ConnectionConfig
			err := ctx.ShouldBindJSON(&body)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			app.SaveConnection(&body)
			ctx.JSON(http.StatusOK, gin.H{"data": body})
		}).
		PUT("/connections/:id", func(ctx *gin.Context) {
			var params types.CommonRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			var body models.ConnectionConfig
			err = ctx.ShouldBindJSON(&body)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.Status(http.StatusOK)
		}).
		PATCH("/connections/:id/default", func(ctx *gin.Context) {
			var params types.CommonRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res := app.SetDefaultConnection(params.ID)
			if res != params.ID {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to set deafult"})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"id": res})
		}).
		GET("/connections/:id", func(ctx *gin.Context) {
			var params types.CommonRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			data := app.GetConnection(params.ID)
			ctx.JSON(http.StatusOK, gin.H{"data": data})
		}).
		GET("/connections/:id/default", func(ctx *gin.Context) {
			var params types.CommonRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			data, _ := app.GetDefaultConnection()
			ctx.JSON(http.StatusOK, gin.H{"data": data})
		}).
		POST("/connections/test", func(ctx *gin.Context) {
			var params types.ConnectionTestParams
			err := ctx.ShouldBindJSON(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			statusOk, errStr := app.TestConnection(params.Connection, params.Exact)
			ctx.JSON(http.StatusOK, gin.H{"ok": statusOk, "error": errStr})
		})

	r.
		GET("/containers", func(ctx *gin.Context) {
			var query types.ContainerListQueryParams
			ctx.ShouldBind(&query)
			var c []*types.ContainerSummary
			if query.All {
				c = app.ContainerList()
			} else {
				c = app.ContainerListRunning()
			}
			ctx.JSON(200, c)
		}).
		GET("/images", func(ctx *gin.Context) {
			i := app.ImageList()
			ctx.JSON(200, i)
		}).
		GET("/volumes", func(ctx *gin.Context) {
			v := app.VolumeList()
			ctx.JSON(200, v)
		}).
		GET("/networks", func(ctx *gin.Context) {
			n := app.NetworkList()
			ctx.JSON(200, n)
		})

	r.POST("/containers/create", func(ctx *gin.Context) {
		var params types.ContainerCreateParams
		ctx.ShouldBindJSON(&params)
		r, e := app.ContainerCreate(&params)
		if e != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
			return
		}
		ctx.JSON(http.StatusOK, r)
	}).
		DELETE("/containers/prune", func(ctx *gin.Context) {})

	r.
		POST("/containers/run", func(ctx *gin.Context) {
			var body types.ContainerCreateParams
			err := ctx.ShouldBindJSON(&body)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			fmt.Println("[params]: ", body.Name, body.Name == "", body.WorkingDir, body.WorkingDir == "")
			res, err := app.ContainerRun(&body)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"id": res.ID})
		}).
		GET("/container/:id", func(ctx *gin.Context) {
			var params types.ContainerRequestParams
			var query types.ContainerGetParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			err = ctx.ShouldBindQuery(&query)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c := app.ContainerGet(&query)
			if c == nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"data": c})
		}).
		GET("/container/:id/inspect", func(ctx *gin.Context) {
			var params types.ContainerRequestParams
			err := ctx.ShouldBindUri(&params)
			fmt.Println("[params]: ", params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			inspectJson, err := app.ContainerInspect(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			j, _ := json.Marshal(inspectJson)
			ctx.String(http.StatusOK, string(j))
		}).
		POST("/container/:id/start", func(ctx *gin.Context) {
			var params types.ContainerRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err = app.ContainerStart(&params)
			if err != nil {
				fmt.Println("[start] error", err.Error())
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			con := app.ContainerGet(&types.ContainerGetParams{ID: params.ID})

			ctx.JSON(http.StatusOK, gin.H{"id": con.ID, "state": con.State, "status": con.Status})
		}).
		PUT("/container/:id/stop", func(ctx *gin.Context) {
			var params types.ContainerRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err = app.ContainerStop(&params)
			if err != nil {
				fmt.Println("[stop]: error", err.Error())
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"status": "exited"})
		}).
		POST("/container/:id/restart", func(ctx *gin.Context) {
			var params types.ContainerRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			err = app.ContainerRestart(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.Status(http.StatusOK)
		}).
		PUT("/container/:id/kill", func(ctx *gin.Context) {
			var params types.ContainerRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			go func() {
				err = app.ContainerKill(&params)
				if err != nil {
					fmt.Println("[kill]: error", err.Error())
					ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
			}()
			ctx.Status(http.StatusOK)
		}).
		PUT("/container/:id/pause", func(ctx *gin.Context) {
			var params types.ContainerRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			err = app.ContainerPause(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.Status(http.StatusOK)
		}).
		PUT("/container/:id/unpause", func(ctx *gin.Context) {
			var params types.ContainerRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			err = app.ContainerUnpause(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.Status(http.StatusOK)
		}).
		GET("/container/:id/diff", func(ctx *gin.Context) {
			var params types.ContainerDiffParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			diffs, err := app.ContainerDiff(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"diffs": diffs})
		}).
		GET("/container/:id/stats", func(ctx *gin.Context) {
			var params types.ContainerStatsParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			stats, err := app.ContainerStats(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"stats": stats})
		}).
		GET("/container/:id/top", func(ctx *gin.Context) {
			var params types.ContainerTopParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			top, err := app.ContainerTop(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			// ctx.JSON(http.StatusOK, gin.H{"top": top})
			ctx.JSON(http.StatusOK, gin.H{"top": types.Record{"titles": top.Titles, "processes": top.Processes}})
		}).
		POST("/container/:id/put_archive", func(ctx *gin.Context) {}).
		GET("/container/:id/get_archive", func(ctx *gin.Context) {}).
		POST("/container/:id/export", func(ctx *gin.Context) {
			var params types.ContainerExportParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			chb := make(chan []byte)
			go func() {
				bytes, err := app.ContainerExport(&params)
				if err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				chb <- bytes
			}()
			bytes := <-chb
			_, _ = ctx.Writer.Write(bytes)
			ctx.Status(http.StatusOK)
			/* ctx.Stream(func(w io.Writer) bool {
				_, e := w.Write(bytes)
				return e == nil
			})
			ctx.JSON(http.StatusOK, gin.H{"error": "ok"}) */
		}).
		GET("/container/:id/files", func(ctx *gin.Context) {}).
		GET("/container/:id/logs", func(ctx *gin.Context) {
			var params types.ContainerLogsParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			var query types.ContainerLogsQuery
			ctx.BindQuery(&query)
			logs, err := app.ContainerLogs(&params, &query)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"logs": logs})
		}).
		POST("/container/:id/exec", func(ctx *gin.Context) {
			var params types.ContainerExecParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				fmt.Println("cannot bind params:", err.Error())
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			var body types.ContainerExecBody
			err = ctx.BindJSON(&body)
			if err != nil {
				fmt.Println("cannot bind body:", err.Error())
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			fmt.Println("[exec#params]:", params, body)
			err = app.ContainerExec(&params, &body)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.Status(http.StatusOK)
		}).
		PATCH("/container/:id/rename", func(ctx *gin.Context) {
			var params types.ContainerRequestParams
			err := ctx.ShouldBindUri(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			var body types.ContainerRenameParams
			err = ctx.ShouldBindJSON(&body)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			err = app.ContainerRename(&params, &body)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.Status(http.StatusOK)
		}).
		DELETE("/container/:id", func(ctx *gin.Context) {
			var params types.ContainerRemoveParams
			err := ctx.ShouldBindJSON(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			err = app.ContainerRemove(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.Status(http.StatusOK)
		}).
		POST("/images/create", func(ctx *gin.Context) {}).
		POST("/images/pull", func(ctx *gin.Context) {
			var params types.ImagePullParams
			err := ctx.ShouldBindJSON(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			logs, err := app.ImagePull(&params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			var progress types.ImagePullProgress
			fmt.Println("[logs]:", logs)
			json.Unmarshal([]byte(logs), &progress)
			fmt.Println("[progress]:", progress)
			ctx.JSON(http.StatusOK, gin.H{"logs": logs, "status": "ok"})
		}).
		POST("/images/build", func(ctx *gin.Context) {
			ct := ctx.GetHeader("Content-Type")
			enc := ctx.GetHeader("Accept-Encoding")
			fmt.Println("[encoding]: ", enc, ct)
			f, _ := ctx.FormFile("file")
			tag := ctx.PostForm("tag")
			fmt.Println("building image with tag:", tag)
			// ctx.SaveUploadedFile(f, f.Filename)
			_, err := os.Stat(".tmp")
			if err != nil {
				fmt.Println("stat error: ", err.Error())
			}
			if os.IsNotExist(err) {
				os.Mkdir(".tmp", os.ModePerm)
			}
			tmpdir, err := os.MkdirTemp(".tmp", "build")
			if err != nil {
				fmt.Println("error processing file:", err.Error())
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			_, err = os.Stat(tmpdir)
			if err != nil {
				fmt.Println("error stat: ", err.Error(), os.IsNotExist(err))
			}
			savePath := path.Join(tmpdir, f.Filename)
			ctx.SaveUploadedFile(f, savePath)

			ch := make(chan bool)
			che := make(chan error)
			done := make(chan bool)
			go func() {
				r, err := os.Open(savePath)
				if err != nil {
					fmt.Println("error reading file:", err)
					che <- err
					return
				}
				unc, err := gzip.NewReader(r)
				if err != nil {
					che <- err
					return
				}

				defer unc.Close()

				sp := savePath
				fmt.Println("[build#result]: ", sp)
				if savePath != "" {
					e := app.ImageBuild(savePath, tag)
					if e != nil {
						fmt.Println("build error: ", e.Error())
						che <- e
						done <- true
						return
					}
					ch <- e == nil
					os.RemoveAll(tmpdir)
				}
				fmt.Println("exiting goroutine")
				che <- nil
				done <- true
			}()
			ok := <-ch
			err = <-che
			d := <-done
			if d {
				fmt.Println("closing channels")
				close(ch)
				close(che)
			}
			if err != nil {
				fmt.Println("error processing image build:", err.Error())
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			fmt.Println("[ok]: ", ok)

			ctx.JSON(http.StatusOK, gin.H{"filename": savePath})
		}).
		PUT("/images/build/:id/cancel", func(ctx *gin.Context) {}).
		DELETE("/images/prune", func(ctx *gin.Context) {}).
		GET("/image/:id/inspect", func(ctx *gin.Context) {
			var params types.ImageRequestParams
			err := ctx.ShouldBindUri(&params)
			fmt.Println("[params]: ", params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			inspectJson, err := app.ImageInspect(params.ID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			j, _ := json.Marshal(inspectJson)
			ctx.String(http.StatusOK, string(j))
		})

	r.GET("/volume/:id/inspect", func(ctx *gin.Context) {
		var params types.VolumeRequestParams
		err := ctx.ShouldBindUri(&params)
		fmt.Println("[params]: ", params)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		inspectJson, err := app.VolumeInspect(params.ID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		j, _ := json.Marshal(inspectJson)
		ctx.String(http.StatusOK, string(j))
	})

	r.
		GET("/network/:id/inspect", func(ctx *gin.Context) {
			var params types.NetworkRequestParams
			err := ctx.ShouldBindUri(&params)
			fmt.Println("[params]: ", params)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			inspectJson, err := app.NetworkInspect(params.ID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			j, _ := json.Marshal(inspectJson)
			ctx.String(http.StatusOK, string(j))
		})

	go func() {
		if err := r.Run(":8080"); err != nil {
			app.Logger.Fatalf("Failed to start server: %s", err)
		}
	}()

	exit := make(chan struct{})
	SignalC := make(chan os.Signal, 5)

	signal.Notify(SignalC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		for s := range SignalC {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
				close(exit)
			}
		}
	}()

	<-exit
	app.SocketServer.Close(nil)
	os.Exit(0)
}
