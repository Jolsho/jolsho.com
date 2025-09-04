package server

import (
    "net/http"
)

// TODO -- also just setup the server with TLS...
const CERT = "/var/certs/cert."
const KEY = "/var/certs/key."

func main() {

	// TODO -- add a rate limiting ip check thing
	// check ips, and increment a viewer count based on 
	// suffeceintly different IPs or something
    http.HandleFunc("/live/main.m3u8", servePlaylist)
    http.HandleFunc("/live/", serveSegment)

    http.HandleFunc("/vod/main.m3u8", serveVodPlaylist)
    http.HandleFunc("/vod/", serveVod)

    http.ListenAndServeTLS(":8000", CERT, KEY, nil)
}
