package server

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Maps obfuscated name → real file path
var tsMap = map[string]string{}

const LIVEDIR = "/var/www/hls/"
const LIVEPLAYLIST = LIVEDIR + "mainstream.m3u8"

func servePlaylist(w http.ResponseWriter, r *http.Request) {
    file, err := os.Open(LIVEPLAYLIST)
    if err != nil {
        http.Error(w, "Playlist not found", 404)
        return
    }
    defer file.Close()

    w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")

    scanner := bufio.NewScanner(file)
    counter := 0
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasSuffix(line, ".ts") {
            obfuscated := fmt.Sprintf("segment-%d.ts", counter)
            tsMap[obfuscated] = LIVEDIR + line
            line = obfuscated
            counter++
        }
        fmt.Fprintln(w, line)
    }
}
func serveSegment(w http.ResponseWriter, r *http.Request) {
    obfuscated := strings.TrimPrefix(r.URL.Path, "/live/")
    realFile, ok := tsMap[obfuscated]
    if !ok {
        http.NotFound(w, r)
        return
    }
    http.ServeFile(w, r, realFile)
}

const VODPLAYLIST = "/var/www/vod/main.m3u8"
const VODDIR = "/var/www/vod/"

func serveVodPlaylist(w http.ResponseWriter, r *http.Request) { 
    w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
    http.ServeFile(w, r, VODPLAYLIST)
}
func serveVod(w http.ResponseWriter, r *http.Request) { 
    segment := strings.TrimPrefix(r.URL.Path, "/vod/")
    http.ServeFile(w, r, VODDIR + filepath.Clean(segment))
}
