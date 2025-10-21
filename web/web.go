package web

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/players"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type sshKeyCache struct {
	signer    ssh.Signer
	createdAt time.Time
}

type sessionMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
}

var sshKeyCacheMap = make(map[string]*sshKeyCache)
var sshKeyCacheMutex sync.RWMutex

func generateSSHKeyPair() (ssh.Signer, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func getOrCreateSSHKey(sessionID string) (ssh.Signer, error) {
	sshKeyCacheMutex.Lock()
	defer sshKeyCacheMutex.Unlock()

	if cached, exists := sshKeyCacheMap[sessionID]; exists {
		if time.Since(cached.createdAt) < 24*time.Hour {
			slog.Info("reusing cached ssh key", "sessionId", sessionID, "age", time.Since(cached.createdAt).String())
			return cached.signer, nil
		}

		delete(sshKeyCacheMap, sessionID)
		slog.Info("ssh key expired, generating new one", "sessionId", sessionID)
	}

	signer, err := generateSSHKeyPair()
	if err != nil {
		return nil, err
	}

	sshKeyCacheMap[sessionID] = &sshKeyCache{
		signer:    signer,
		createdAt: time.Now(),
	}

	slog.Info("generated and cached new ssh key", "sessionId", sessionID)
	return signer, nil
}

func cleanupExpiredKeys() {
	sshKeyCacheMutex.Lock()
	defer sshKeyCacheMutex.Unlock()

	now := time.Now()
	cleaned := 0
	for sessionID, cached := range sshKeyCacheMap {
		if now.Sub(cached.createdAt) > 24*time.Hour {
			delete(sshKeyCacheMap, sessionID)
			cleaned++
		}
	}

	if cleaned > 0 {
		slog.Info("cleaned up expired ssh keys", "count", cleaned, "remaining", len(sshKeyCacheMap))
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if config.GetWebAllowedOrigins() != "" {
			origin := r.Header.Get("Origin")
			allowedOrigins := strings.SplitSeq(config.GetWebAllowedOrigins(), ",")
			for o := range allowedOrigins {
				if strings.TrimSpace(o) == origin {
					return true
				}
			}
			slog.Warn("websocket connection rejected due to origin", "origin", origin)
			return false
		}

		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	slog.Info("websocket connection attempt", "address", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("websocket connection established successfully", "address", r.RemoteAddr)

	var sessionID string
	var signer ssh.Signer

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	messageType, message, err := conn.ReadMessage()
	if err != nil {
		slog.Error("error reading initial message", "error", err)
		return
	}

	sessionID = "fallback-" + strings.ReplaceAll(r.RemoteAddr, ":", "-")
	if messageType == websocket.TextMessage {
		var sessionMsg sessionMessage
		if err := json.Unmarshal(message, &sessionMsg); err == nil && sessionMsg.Type == "session" {
			sessionID = sessionMsg.SessionID
			slog.Info("received session id", "sessionId", sessionID)
		}
	}

	conn.SetReadDeadline(time.Time{})

	signer, err = getOrCreateSSHKey(sessionID)
	if err != nil {
		slog.Error("ssh key generation/retrieval failed", "error", err)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("SSH key generation failed: "+err.Error()))
		return
	}

	sshConfig := &ssh.ClientConfig{
		User: "web-client",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	serverAddr := net.JoinHostPort(config.GetServerHost(), config.GetServerPortSSH())
	slog.Info("attempting ssh connection", "address", serverAddr, "user", sshConfig.User)

	sshConn, err := ssh.Dial("tcp", serverAddr, sshConfig)
	if err != nil {
		slog.Error("ssh connection failed", "error", err)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("SSH connection failed: "+err.Error()))
		return
	}
	defer sshConn.Close()
	slog.Info("ssh connection established successfully")

	session, err := sshConn.NewSession()
	if err != nil {
		slog.Error("ssh session creation failed", "error", err)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("SSH session creation failed: "+err.Error()))
		return
	}
	defer session.Close()

	err = session.RequestPty("xterm-256color", 120, 33, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	})
	if err != nil {
		slog.Error("pty request failed", "error", err)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("PTY request failed: "+err.Error()))
		return
	}

	sshIn, err := session.StdinPipe()
	if err != nil {
		slog.Error("ssh stdin pipe failed", "error", err)
		return
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		slog.Error("ssh stdout pipe failed", "error", err)
		return
	}

	sshErr, err := session.StderrPipe()
	if err != nil {
		slog.Error("ssh stderr pipe failed", "error", err)
		return
	}

	err = session.Shell()
	if err != nil {
		slog.Error("ssh shell start failed", "error", err)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("SSH shell start failed: "+err.Error()))
		return
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		buffer := make([]byte, 1024)
		for {
			n, err := sshOut.Read(buffer)
			if err != nil {
				if err != io.EOF {
					slog.Error("SSH stdout read error", "error", err)
				}
				break
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
				slog.Error("WebSocket write error", "error", err)
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		buffer := make([]byte, 1024)
		for {
			n, err := sshErr.Read(buffer)
			if err != nil {
				if err != io.EOF {
					slog.Error("SSH stderr read error", "error", err)
				}
				break
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
				slog.Error("WebSocket write error", "error", err)
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				slog.Error("error reading message", "error", err)
				break
			}

			if messageType == websocket.TextMessage && len(message) > 7 && string(message[:7]) == "RESIZE:" {
				parts := strings.Split(string(message[7:]), ",")
				if len(parts) == 2 {
					if cols, err1 := strconv.Atoi(parts[0]); err1 == nil {
						if rows, err2 := strconv.Atoi(parts[1]); err2 == nil {
							_ = session.WindowChange(rows, cols)
							continue
						}
					}
				}
			}

			if _, err := sshIn.Write(message); err != nil {
				slog.Error("ssh stdin write error", "error", err)
				break
			}
		}
	}()

	_ = session.Wait()
	wg.Wait()
}

func Run() error {
	mux := http.NewServeMux()

	// WebSocket endpoint for terminal connections
	log.Printf("Registering WebSocket handler at /ws")
	mux.HandleFunc("/ws", handleWebSocket)

	if config.GetDebug() {
		mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/heap", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/goroutine", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/threadcreate", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/block", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/cmdline", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/symbol", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/all", http.DefaultServeMux.ServeHTTP)
	}

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/favicon.ico")
	})

	mux.HandleFunc("/dist/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web"+r.URL.Path)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Eventually implement admin only information
		// params := r.URL.Query()
		// if params.Get("admin_key") == config.GetWebAdminKey() {}
		total := len(games.GetAll())
		totalStarted := 0
		for _, game := range games.GetAll() {
			if game.InProgress {
				totalStarted++
			}
		}

		totalAbandoned := 0
		for _, game := range games.GetAll() {
			if game.WinnerID == "" && !game.InProgress {
				totalAbandoned++
			}
		}

		t, err := template.ParseFiles("web/dist/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = t.Execute(w, struct {
			TotalGames            int
			TotalStartedGames     int
			TotalAbandonedGames   int
			TotalPlayers          int
			TotalVisitors         int
			TotalConnectedPlayers int
			ActiveGames           map[string]*games.Game
		}{
			TotalGames:            total,
			TotalStartedGames:     totalStarted,
			TotalAbandonedGames:   totalAbandoned,
			TotalPlayers:          players.GetUniquePlayerCount(),
			TotalVisitors:         players.GetVisitorPlayerCount(),
			TotalConnectedPlayers: players.GetConnectedPlayerCount(),
			ActiveGames:           games.GetAllActive(),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupExpiredKeys()
		}
	}()

	webPort := config.GetServerPortWeb()
	slog.Info("Starting web server", "port", webPort)
	return http.ListenAndServe(":"+webPort, mux)
}
