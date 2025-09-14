package servers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func mainSwitch(state *ServerState, w http.ResponseWriter, r *http.Request) {
	// Parse form data (from POST body)
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
	switch r.URL.String() {
		case "/auth": auth(state, w, r)
		case "/new_viewer": newViewer(state, w, r)
		case "/viewer_left": viewerLeft(state, w, r)
		case "/publish": publish(state, w, r)
		case "/publish_done": publishDone(state, w, r)
		default: 
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
	}
}

func BuildNginxServer(state *ServerState) func() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mainSwitch(state, w, r)
	})
	return func() {
		server := &http.Server{
			Addr: ":8081",
			Handler: state.rateLimitedHandler(mux),
		}

		go func() {
			log.Println("HTTP server starting on :8081")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}()

		watchStopSever(server)
	}
}


func auth(state *ServerState, w http.ResponseWriter, r *http.Request) {
    // Extract values
    addr := r.FormValue("addr")
	name := r.FormValue("name")

	if !state.allowedIps[addr] {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	dir := "/var/hls/" + name
	if err := ClearDir(dir); err != nil {
		if !os.IsNotExist(err) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ClearDir Error"))
			return
		} else {
			err := os.Mkdir(dir, 0777)
			if err == nil {
				err = os.Chmod(dir, 0777)
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("ClearDir Error"))
				return
			}
		}
	}

    // Respond OK so RTMP continues
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

// ClearDir removes all contents of the specified directory but keeps the directory itself
func ClearDir(dirPath string) error {
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        return fmt.Errorf("failed to read directory: %w", err)
    }

    for _, entry := range entries {
        entryPath := filepath.Join(dirPath, entry.Name())
        err := os.RemoveAll(entryPath)
        if err != nil {
            return fmt.Errorf("failed to remove %s: %w", entryPath, err)
        }
    }
    return nil
}

func publish(state *ServerState, w http.ResponseWriter, r *http.Request) {
	stream := &Stream{
		Name: r.FormValue("name"),
		IsLive: true,
		Title: "Live",
		Viewers: 0,
	}
	state.streams[stream.Name] = stream

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func publishDone(state *ServerState, w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	go RebuildM3U8("/var/hls/" + name)
	state.streams[name].IsLive = false
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

// RebuildM3U8 rebuilds the index.m3u8 file with all .ts files in numeric order
func RebuildM3U8(dir string) error {
	time.Sleep(30 * time.Second)
    files, err := os.ReadDir(dir)
    if err != nil { return err }

    var tsFiles []string

    // Collect .ts files and extract numbers for sorting
    for _, f := range files {
        if !f.IsDir() && strings.HasSuffix(f.Name(), ".ts") {
            tsFiles = append(tsFiles, f.Name())
        }
    }

    // Sort numerically based on the filename number
    sort.Slice(tsFiles, func(i, j int) bool {
        iNum, _ := strconv.Atoi(strings.TrimSuffix(tsFiles[i], ".ts"))
        jNum, _ := strconv.Atoi(strings.TrimSuffix(tsFiles[j], ".ts"))
        return iNum < jNum
    })

    // Build new index content
    var sb strings.Builder
    sb.WriteString("#EXTM3U\n")
    sb.WriteString("#EXT-X-VERSION:3\n")
    sb.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n") // reset sequence
    sb.WriteString("#EXT-X-TARGETDURATION:10\n") // you can adjust

    for _, ts := range tsFiles {
        sb.WriteString("#EXTINF:10,\n") // you can calculate duration if needed
        sb.WriteString(ts + "\n")
    }

    indexPath := filepath.Join(dir, "index.m3u8")
    return os.WriteFile(indexPath, []byte(sb.String()), 0644)
}

func newViewer(state *ServerState, w http.ResponseWriter, r *http.Request) {
	state.streams[r.FormValue("name")].Viewers++
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func viewerLeft(state *ServerState, w http.ResponseWriter, r *http.Request) {
	state.streams[r.FormValue("name")].Viewers--
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

