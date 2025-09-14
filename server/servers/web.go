package servers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"golang.org/x/time/rate"
)


func BuildBasicHttp() func() {
	return func() {
		server := &http.Server{
			Addr: ":80",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				target := "https://" + r.Host + r.URL.RequestURI()
				http.Redirect(w, r, target, http.StatusMovedPermanently)
			}),
		}

		go func() {
			log.Println("HTTP server starting on :80")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}()

		watchStopSever(server)
	}
}

func BuildHttps(state *ServerState) func() {
	mux := http.NewServeMux()

	// Serve static files from ./dst
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./dst/index.html")
	})

	mux.HandleFunc("/image/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./dst/me.jpg")
	})

	mux.Handle("/assets/", http.FileServer(http.Dir("./dst")))
	mux.HandleFunc("/chat", state.HandleChat)
	mux.HandleFunc("/isLive", state.IsLive)
	mux.Handle("/hls/", hlsHandler())

	// Wrap with rate limiting
	return func() {
		server := &http.Server{
			Addr:    ":443",
			Handler: state.rateLimitedHandler(mux),
		}
	
		go func () {
			log.Println("HTTPS server starting on :443")
			if err := server.ListenAndServeTLS("../assets/server.crt", "../assets/server.key"); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}()

		watchStopSever(server)
	}
}

func (s *ServerState)rateLimitedHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter, exists := s.visitors[r.RemoteAddr]
		if !exists {
			limiter = rate.NewLimiter(2, 12) // 1 req/sec, burst 5
			s.visitors[r.RemoteAddr] = limiter
		}

        if !limiter.Allow() {
            w.WriteHeader(http.StatusTooManyRequests)
            fmt.Fprint(w, "Too Many Requests")
            return
        }
        next.ServeHTTP(w, r)
    })
}

// hlsHandler returns a handler that serves HLS files under the given root directory
func hlsHandler() http.Handler {
    fs := http.FileServer(http.Dir("/var/hls"))

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/hls")

        // Set Cache-Control
        w.Header().Set("Cache-Control", "no-cache")
        // Set CORS
        w.Header().Set("Access-Control-Allow-Origin", "*")

        // Set MIME type for HLS files explicitly
        ext := filepath.Ext(r.URL.Path)
        switch ext {
        case ".m3u8":
            w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
        case ".ts":
            w.Header().Set("Content-Type", "video/mp2t")
        }

        // Serve the file
        fs.ServeHTTP(w, r)
    })
}

