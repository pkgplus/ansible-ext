package gateway

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"ansible-ext/proto/ansible"
	"ansible-ext/proto/hostmanager"

	"ansible-ext/server"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

func RestfulServer(ctx context.Context, grpc_addr, listen_addr string) error {
	swgger_dir := "./"

	conn, err := dial(ctx, "tcp", grpc_addr)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", swaggerServer(swgger_dir))
	mux.HandleFunc("/healthz", healthzServer(conn))
	mux.HandleFunc("/api/v1/ansible/playbooks/", server.PlayBookServer())

	gw, err := newGateway(ctx, conn)
	if err != nil {
		return err
	}
	mux.Handle("/", gw)

	s := &http.Server{
		Addr:    listen_addr,
		Handler: allowCORS(mux),
	}
	go func() {
		<-ctx.Done()
		log.Printf("Shutting down the http server")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown http server: %v", err)
		}
	}()

	log.Printf("Starting listening at %s", listen_addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("Failed to listen and serve: %v", err)
		return err
	}

	return nil
}

func swaggerServer(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
			log.Printf("Not Found: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		log.Printf("Serving %s", r.URL.Path)
		p := strings.TrimPrefix(r.URL.Path, "/swagger/")
		p = path.Join(dir, p)
		http.ServeFile(w, r, p)
	}
}

func healthzServer(conn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		s := conn.GetState()
		if s != connectivity.Ready {
			http.Error(w, fmt.Sprintf("grpc server is %s", s), http.StatusBadGateway)
			return
		}
		fmt.Fprintln(w, "ok")
	}
}

func dial(ctx context.Context, network, addr string) (*grpc.ClientConn, error) {
	switch network {
	case "tcp":
		return dialTCP(ctx, addr)
	case "unix":
		return dialUnix(ctx, addr)
	default:
		return nil, fmt.Errorf("unsupported network type %q", network)
	}
}

// dialTCP creates a client connection via TCP.
// "addr" must be a valid TCP address with a port number.
func dialTCP(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, addr, grpc.WithInsecure())
}

// dialUnix creates a client connection via a unix domain socket.
// "addr" must be a valid path to the socket.
func dialUnix(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	d := func(addr string, timeout time.Duration) (net.Conn, error) {
		return net.DialTimeout("unix", addr, timeout)
	}
	return grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithDialer(d))
}

func newGateway(ctx context.Context, conn *grpc.ClientConn) (http.Handler, error) {
	mux := runtime.NewServeMux()

	if err := Ansible.RegisterAnsibleHandler(ctx, mux, conn); err != nil {
		return nil, err
	}

	if err := HostManager.RegisterHostManagerHandler(ctx, mux, conn); err != nil {
		return nil, err
	}

	return mux, nil
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	log.Printf("preflight request for %s", r.URL.Path)
}
