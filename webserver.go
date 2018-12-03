package main

import (
	pb "./protos/messages"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
)

func serveTemplate(w http.ResponseWriter, r *http.Request, client pb.MessageServiceClient) {
	layoutPath := filepath.Join("templates", "layout.html")
	templatePath := filepath.Join("templates", fmt.Sprintf("%s.html", filepath.Clean(r.URL.Path)))
	tpl, err := template.ParseFiles(layoutPath, templatePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something bad happened"))
		return
	}

	q := r.URL.Query()
	keys, keysExist := q["key"]
	if !keysExist {
		tpl.ExecuteTemplate(w, "layout", "")
		return
	}

	value, err := getValueFromKey(client, keys[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error retrieving key"))
		return
	}

	tpl.ExecuteTemplate(w, "layout", value)
}

func getValueFromKey(client pb.MessageServiceClient, key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &pb.MessageRequest{
		Key: key,
	}
	resp, err := client.GetMessage(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.GetValue(), nil
}

func serve(client pb.MessageServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, r, client)
	}
}

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := pb.NewMessageServiceClient(conn)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", serve(client))

	log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", nil))
}
