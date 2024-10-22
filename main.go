package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/actor-staged-form/commands"
	"github.com/anthdm/hollywood/actor"
	"github.com/golang-module/carbon"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const MAX_STAGES = 2

type responseWriter struct {
	http.ResponseWriter
	http.Hijacker
	StatusCode int
}

func NewResponseWriter(w http.ResponseWriter, h http.Hijacker) *responseWriter {
	return &responseWriter{w, h, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

var totalReqs = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "http_request_total",
	Help: "Total number of HTTP requests",
}, []string{"path"})

var responseStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "response_status",
	Help: "Status of response",
}, []string{"status", "path"})

var httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Response time of HTTP requests.",
}, []string{"path"})

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			path, _ := route.GetPathTemplate()
			timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))

			rw := NewResponseWriter(w, w.(http.Hijacker))
			next.ServeHTTP(rw, r)

			statusCode := rw.StatusCode

			responseStatus.WithLabelValues(strconv.Itoa(statusCode), path).Inc()
			totalReqs.WithLabelValues(path).Inc()

			timer.ObserveDuration()
		},
	)
}

func init() {
	prometheus.Register(totalReqs)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
}

func GenerateId() string {
	// Create a new ULID
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	id, err := ulid.New(ms, entropy)
	if err != nil {
		fmt.Printf("error while generating ULID: %v\n", err)
	}
	return id.String()
}

type TestServer struct {
	ctx    *actor.Context
	stages map[*actor.PID]struct{}
}

func NewTestServer() actor.Receiver {
	return &TestServer{
		stages: make(map[*actor.PID]struct{}),
	}
}

func (s *TestServer) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		_ = msg
		s.ctx = ctx
		s.start(":2222")
	}
}

func (s *TestServer) start(port string) {

	go func() {
		r := mux.NewRouter()
		r.Use(prometheusMiddleware)

		r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("This is the homepage"))
		})

		r.HandleFunc("/on", s.ConnectionHandler)
		r.Handle("/metrics", promhttp.Handler())

		log.Println("APPLICATION STARTED AT PORT=", port)
		log.Fatal(http.ListenAndServe(port, r))
	}()
}

func (s *TestServer) ConnectionHandler(w http.ResponseWriter, r *http.Request) {
	if len(s.stages) == MAX_STAGES {
		w.Header().Set("FAILURE", "Stage cannot be created at this time")
		http.Error(w, "stage cannot be created at this time", 500)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("An error occured upgrading connection: ", err)
		return
	}

	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		fmt.Println("An error occured: ", err)
		return
	}

	sid := GenerateId()

	pid := s.ctx.SpawnChild(NewStage(conn, sid), fmt.Sprintf("stage_%s", sid))
	// Lines after spawning an actor child not working
	s.stages[pid] = struct{}{}
}

type Stage struct { //represent an actor
	origin     string
	resolution string
	content    []interface{}
	conn       *websocket.Conn
	stageId    string
	state      string
}

func NewStage(conn *websocket.Conn, stageId string) actor.Producer {
	return func() actor.Receiver {
		return &Stage{
			conn:       conn,
			stageId:    stageId,
			origin:     carbon.Now().ToDateTimeString(),
			resolution: carbon.Now().ToDateTimeString(),
			state:      "start_tiny_test",
		}
	}
}

func (s *Stage) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case actor.Started:
		go s.Read()
	}
}

func (s *Stage) CheckState(current_state string) error {
	if s.state == current_state {
		return nil
	}
	return errors.New("state not accessible at this time")
}

func (s *Stage) MoveToNextState(next_state string) {
	s.state = next_state
}

func (s *Stage) Read() {
	var msg commands.WSMessage

	for {
		err := s.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error occurred reading data: ", err)
			return
		}
		go s.MessageHandler(msg)
	}
}

func (s *Stage) AddInput(tinyTestInput *commands.TinyTestInput) {
	input, err := json.MarshalIndent(tinyTestInput, "", " ")
	if err != nil {
		panic(err)
	}
	s.content = append(s.content, string(input))
}

func Decode(encoded string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		fmt.Println("Unable to decode data")
		return nil, fmt.Errorf("unable to decode data")
	}
	return b, nil
}

func Encode(v interface{}) (string, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("error encoding json data")
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)
	return encoded, nil
}

// State is also msg.Type

func (s *Stage) MessageHandler(msg commands.WSMessage) {
	switch msg.Type {
	case "start_tiny_test":
		if err := s.CheckState(msg.Type); err != nil {
			resp := commands.NewErrResponse(msg.Type, err)
			err = s.conn.WriteJSON(resp)
			if err != nil {
				fmt.Println("Unable to send error response: ", err)
				return
			}
			return
		}

		s.ProcessTinyTest(msg)
		s.MoveToNextState("lawrence_test")

	case "lawrence_test":
		if err := s.CheckState(msg.Type); err != nil {
			resp := commands.NewErrResponse(msg.Type, err)
			err = s.conn.WriteJSON(resp)
			if err != nil {
				fmt.Println("Unable to send error response: ", err)
				return
			}
			return
		}

		s.conn.WriteJSON(commands.NewSuccessResponse("Lawrence is !HIM actually"))
		s.MoveToNextState("sad")
	case "content":
		if err := s.CheckState(msg.Type); err != nil {
			resp := commands.NewErrResponse(msg.Type, err)
			err = s.conn.WriteJSON(resp)
			if err != nil {
				fmt.Println("Unable to send error response: ", err)
				return
			}
			return
		}

		if err := s.conn.WriteJSON(s.content); err != nil {
			fmt.Println("Error occured sending json data")
			return
		}

	default:
		s.conn.WriteJSON(commands.NewSuccessResponse("Command not supported yet!"))
	}
}

func (s *Stage) ProcessTinyTest(msg commands.WSMessage) {
	var tinyTestInput *commands.TinyTestInput
	var stringToSend string
	_, ok := msg.Data.(string)
	if !ok {
		encodedString, err := Encode(msg.Data)
		if err != nil {
			fmt.Println(err)
			return
		}

		stringToSend = encodedString
	} else {
		stringToSend = msg.Data.(string)
	}

	b, err := Decode(stringToSend)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(b, &tinyTestInput)
	if err != nil {
		fmt.Println("Unable to unmarshal: ", err)
		return
	}

	output, err := commands.SerializeTinyTestInput(tinyTestInput)
	if err != nil {
		fmt.Println("Unable to serialize tiny test input: ", err)
		return
	}
	if err = s.conn.WriteJSON(output); err != nil {
		fmt.Println("Unable to send response to user: ", err)
		return
	}
	go s.AddInput(tinyTestInput)
}

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		fmt.Println(err)
		return
	}

	engine.Spawn(NewTestServer, "test_server")
	select {}
}
