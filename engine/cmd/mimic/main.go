package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"openmimic/engine/internal/importer"
	"openmimic/engine/internal/service"
	pb "openmimic/engine/pkg/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	grpcPort := flag.Int("grpc-port", 50051, "gRPC server port")
	httpPort := flag.Int("http-port", 8080, "HTTP server port for Studio REST API")
	flag.Parse()

	srv := service.NewGeneratorServer()
	platformStore := service.NewPlatformStore()

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("failed to listen on gRPC port %d: %v", *grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGeneratorServiceServer(grpcServer, srv)

	go func() {
		log.Printf("OpenMimic Engine gRPC server listening on :%d", *grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Start HTTP JSON bridge for Visual Studio
	mux := http.NewServeMux()

	// CORS wrapper
	enableCORS := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	}

	mux.HandleFunc("/api/preview", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req pb.StreamPreviewRequest
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
			return
		}

		unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}
		if err := unmarshaler.Unmarshal(bodyBytes, &req); err != nil {
			http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		var events []map[string]any

		previewCount := req.GetPreviewCount()
		if previewCount <= 0 {
			previewCount = 10
		}

		eng := service.NewGeneratorServer()
		stream := &httpPreviewStream{ctx: r.Context()}
		if err := eng.StreamPreview(&req, stream); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, resp := range stream.results {
			var parsed map[string]any
			_ = json.Unmarshal([]byte(resp.JsonPayload), &parsed)
			events = append(events, parsed)
		}

		_ = json.NewEncoder(w).Encode(events)
	}))

	mux.HandleFunc("/api/generators/start", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req pb.StartGeneratorRequest
		unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}
		if err := unmarshaler.Unmarshal(bodyBytes, &req); err != nil {
			http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
			return
		}

		resp, err := srv.StartGenerator(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		platformStore.LogAudit("START_GENERATOR", resp.GeneratorId, "studio_user", fmt.Sprintf("Started generator %s", req.GetConfig().GetName()))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	mux.HandleFunc("/api/generators/stop", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req pb.StopGeneratorRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
			return
		}

		resp, err := srv.StopGenerator(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		platformStore.LogAudit("STOP_GENERATOR", req.GetGeneratorId(), "studio_user", "Stopped generator execution")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	mux.HandleFunc("/api/import/sql", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			DDL string `json:"ddl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.DDL == "" {
			http.Error(w, "Invalid request: 'ddl' field required", http.StatusBadRequest)
			return
		}

		cfg, err := importer.SQLDDLToGeneratorConfig(payload.DDL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse DDL: %v", err), http.StatusUnprocessableEntity)
			return
		}

		jsonBytes, err := protojson.Marshal(cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonBytes)
	}))

	mux.HandleFunc("/api/templates", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		templates := platformStore.ListTemplates()
		var list []map[string]any
		for _, tpl := range templates {
			configBytes, _ := protojson.Marshal(tpl.GetConfig())
			var cfgMap map[string]any
			_ = json.Unmarshal(configBytes, &cfgMap)
			list = append(list, map[string]any{
				"id":       tpl.GetId(),
				"title":    tpl.GetTitle(),
				"category": tpl.GetCategory(),
				"config":   cfgMap,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}))

	mux.HandleFunc("/api/audit", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		logs := platformStore.GetAuditLogs(50)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(logs)
	}))

	log.Printf("OpenMimic REST/HTTP bridge listening on :%d", *httpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), mux); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

type httpPreviewStream struct {
	grpc.ServerStream
	ctx     context.Context
	results []*pb.StreamPreviewResponse
}

func (s *httpPreviewStream) Context() context.Context {
	return s.ctx
}

func (s *httpPreviewStream) Send(r *pb.StreamPreviewResponse) error {
	s.results = append(s.results, r)
	return nil
}
