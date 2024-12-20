package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	wishlog "github.com/charmbracelet/log"
	wissh "github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/logging"
	"github.com/google/uuid"
)

type Payload struct {
	Container string
	Path      string
}
type JsonObject map[string]any
type FileInfo struct {
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

const HOST = ""
const PORT = "45500"
const AUTH_USER = "user"
const AUTH_PASS = "pass"
const AUTH_SECRET = "secret"

func listFiles(p string) ([]byte, error) {
	_, err := os.Stat(p)
	if err != nil {
		return []byte{}, err
	}
	res := JsonObject{}
	folders := make([]FileInfo, 0)
	files := make([]FileInfo, 0)
	fileInfo, err := os.ReadDir(p)
	if err != nil {
		return []byte{}, err
	}
	for _, file := range fileInfo {
		inf, err := file.Info()
		if err != nil {
			log.Println("[error]: ", err.Error())
			continue
		}
		isDir := false
		abs, err := filepath.Abs(path.Join(p, inf.Name()))
		if err != nil {
			log.Println("[error]: ", err.Error())
			continue
		}
		stat, err := os.Lstat(abs)
		if err != nil {
			log.Println("[error]: ", err.Error())
			continue
		}
		isLink := stat.Mode()&os.ModeSymlink != 0
		if !isLink && !inf.IsDir() {
			continue
		}

		isDir = stat.IsDir()
		log.Println("[stat]:", stat.Name(), isLink)
		stat_u := stat.Sys().(*syscall.Stat_t)
		filetype := "dir"
		realPath := abs
		if isLink {
			link, err := filepath.EvalSymlinks(abs)
			if err != nil {
				log.Println("[error]:", err.Error())
				continue
			}
			stat, err = os.Lstat(link)
			if err != nil {
				log.Println("[error]: ", err.Error())
				continue
			}
			if !stat.IsDir() {
				continue
			}
			isDir = true
			realPath = link
			log.Println("[dir#link]:", stat.Name(), link, stat.IsDir())
		}

		id, err := uuid.NewV7()
		if err != nil {
			log.Println("[error]:", err.Error())
			continue
		}
		folders = append(folders, FileInfo{
			Name:     inf.Name(),
			Type:     filetype,
			IsDir:    isDir,
			IsLink:   isLink,
			Created:  fmt.Sprintf("%+v", time.Unix(stat_u.Ctim.Sec, stat_u.Ctim.Nsec)),
			Modified: fmt.Sprintf("%+v", time.Unix(stat_u.Atim.Sec, stat_u.Atim.Nsec)),
			Size:     int(stat.Size()),
			Perm:     stat.Mode().String(),
			Owner:    fmt.Sprintf("%+v", stat_u.Uid),
			Group:    fmt.Sprintf("%+v", stat_u.Gid),
			RealPath: realPath,
			ID:       id.String(),
		})
	}
	res["folders"] = folders
	res["folder_count"] = len(folders)
	for _, file := range fileInfo {
		inf, err := file.Info()
		isLink := false
		if inf.IsDir() {
			continue
		}
		if err != nil {
			log.Println("error: ", err.Error())
			continue
		}
		abs, err := filepath.Abs(path.Join(p, inf.Name()))
		if err != nil {
			log.Println("error: ", err.Error())
			continue
		}

		stat, err := os.Lstat(abs)
		if err != nil {
			log.Println("error: ", err.Error())
			continue
		}
		isDir := stat.IsDir()
		isLink = stat.Mode()&os.ModeSymlink != 0
		log.Println("[stat]:", stat.Name(), isLink)
		stat_u := stat.Sys().(*syscall.Stat_t)
		if isLink {
			link, err := filepath.EvalSymlinks(abs)
			if err != nil {
				log.Println("[error]:", err.Error())
				continue
			}
			stat, err = os.Lstat(link)
			if err != nil {
				log.Println("[error]: ", err.Error())
				continue
			}
			if stat.IsDir() {
				continue
			}
			log.Println("[file#link]:", link, stat.IsDir())
		}

		id, err := uuid.NewV7()
		if err != nil {
			log.Println("[error]:", err.Error())
			continue
		}
		filetype := mime.TypeByExtension(path.Ext(stat.Name()))
		realPath := abs
		files = append(files, FileInfo{
			Name:     inf.Name(),
			Type:     filetype,
			IsDir:    isDir,
			IsLink:   isLink,
			Created:  fmt.Sprintf("%+v", time.Unix(stat_u.Ctim.Sec, stat_u.Ctim.Nsec)),
			Modified: fmt.Sprintf("%+v", time.Unix(stat_u.Atim.Sec, stat_u.Atim.Nsec)),
			Size:     int(stat.Size()),
			Perm:     stat.Mode().String(),
			Owner:    fmt.Sprintf("%+v", stat_u.Uid),
			Group:    fmt.Sprintf("%+v", stat_u.Gid),
			RealPath: realPath,
			ID:       id.String(),
		})
	}
	// log.Printf("len: %d\n", len(files))
	res["files"] = files
	res["file_count"] = len(files)

	return json.MarshalIndent(res, "", "  ")
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		user, pass, ok := req.BasicAuth()
		if ok {
			userHash := sha256.Sum256([]byte(user))
			passHash := sha256.Sum256([]byte(pass))
			realUserHash := sha256.Sum256([]byte(AUTH_USER))
			realPassHash := sha256.Sum256([]byte(AUTH_PASS))

			userMatch := (subtle.ConstantTimeCompare(userHash[:], realUserHash[:]) == 1)
			passMatch := (subtle.ConstantTimeCompare(passHash[:], realPassHash[:]) == 1)

			if userMatch && passMatch {
				next.ServeHTTP(w, req)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func checkAuthSecret(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		secret := req.Header.Get("X-Secret")
		secretHash := sha256.Sum256([]byte(secret))
		realSecretHash := sha256.Sum256([]byte(AUTH_SECRET))

		secretMatch := (subtle.ConstantTimeCompare(secretHash[:], realSecretHash[:]) == 1)
		if secretMatch {
			next.ServeHTTP(w, req)
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func lsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var p Payload
	err := json.NewDecoder(req.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("path:", p.Path)
	response, err := listFiles(p.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%+v", string(response))
}

const (
	host        = "localhost"
	port        = "2222"
	hostKeyPath = "/EXTRA/dev/docker/ssh-keys/id_rsa"
)

func main() {
	s := http.Server{
		Addr: fmt.Sprintf("%s:%s", HOST, PORT),
	}
	http.HandleFunc("/ls", checkAuthSecret(basicAuth(lsHandler)))
	http.HandleFunc("/down", checkAuthSecret(basicAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		log.Println("Shut down signal received. Exiting")
		os.Exit(0)
		w.Header().Set("Content-Type", "application/json")
		// w.WriteHeader(200)
		// fmt.Fprintf(w, "%+v", string([]byte{}))
		s.Shutdown(context.Background())
	})))

	fwdh := &wissh.ForwardedTCPHandler{}
	srv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithPublicKeyAuth(func(ctx wissh.Context, key wissh.PublicKey) bool {
			// ssh.ParseKnownHosts([]byte("/home/algae/.ssh/known_hosts"))
			return true
		}),
		func(s *wissh.Server) error {
			s.LocalPortForwardingCallback = func(ctx wissh.Context, destinationHost string, destinationPort uint32) bool {
				return true
			}
			s.ReversePortForwardingCallback = func(ctx wissh.Context, destinationHost string, destinationPort uint32) bool {
				wishlog.Info("reverse port forwarding allowed", "host", destinationHost, "port", destinationPort)
				return true
			}
			s.RequestHandlers = map[string]wissh.RequestHandler{
				"tcpip-forward":        fwdh.HandleSSHRequest,
				"cancel-tcpip-forward": fwdh.HandleSSHRequest,
			}
			return nil
		},
		wissh.AllocatePty(),
		wish.WithSubsystem("bash", func(s wissh.Session) {
			wish.Println(s, "user", s.User())
		}),
		wish.WithMiddleware(
			func(next wissh.Handler) wissh.Handler {
				return func(s wissh.Session) {
					wishlog.Info("Commands:", s.Command())
					cmd := wish.Command(s, "su", "-", s.User())
					if err := cmd.Run(); err != nil {
						wish.Fatalln(s, err)
					}

					next(s)
				}
			},

			activeterm.Middleware(),
			logging.Middleware(),
		),
	)

	if err != nil {
		wishlog.Error("Could not start server", "error", err)
	}

	go func() {
		if err = http.ListenAndServe(":45500", nil); err != nil {
			log.Fatalf("Could not start server: %s\n", err.Error())
			return
		}
		log.Println("Server listening on port 45500")
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	wishlog.Info("Starting HTTP server", "host", "host", "port", "45500")
	wishlog.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, wissh.ErrServerClosed) {
			wishlog.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	wishlog.Info("Stopping SSH server")
	if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, wissh.ErrServerClosed) {
		wishlog.Error("could not stop server", "error", err)
	}
	wishlog.Info("Stopping HTTP server")
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println("Could not stop server", "error", err.Error())
	}
}
