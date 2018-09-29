package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

//go:generate go-bindata-assetfs ./dist/...

const (
	TTL = 60 * 60 * 24 * 14
)

var (
	CONFUSING_MULTI = []string{
		"ij1lI",
		"oO0Q",
		"uUvVwW",
		"nhm",
		"sS5",
	}

	CONFUSING_PAIR = []string{
		"b6",
		"q9",
		"B8",
		"cC",
		"zZ",
		"xX",
		"jJ",
	}

	CONFUSING_ALL = append(CONFUSING_MULTI, CONFUSING_PAIR...)

	CONFUSING_CHAR = flatten(CONFUSING_ALL...)
)

func flatten(xs ...string) string {
	return strings.Join(xs, "")
}

func selectAndBuild(as, bs, cs string) string {
	a := as[rand.Intn(len(as))]

	b := a
	for b == a {
		b = bs[rand.Intn(len(bs))]
	}

	c := b
	for b == c {
		c = cs[rand.Intn(len(cs))]
	}

	return string([]byte{a, b, c})
}

func ThreeConfusing() string {
	set := CONFUSING_MULTI[rand.Intn(len(CONFUSING_MULTI))]

	return selectAndBuild(set, set, set)
}

func TwoConfisuing() string {
	set := CONFUSING_ALL[rand.Intn(len(CONFUSING_ALL))]

	switch rand.Intn(3) {
	case 0:
		return selectAndBuild(CONFUSING_CHAR, set, set)

	case 1:
		return selectAndBuild(set, CONFUSING_CHAR, set)

	case 2:
		return selectAndBuild(set, set, CONFUSING_CHAR)

	default:
		return ""
	}
}

func MakeKey() string {
	if rand.Intn(1) == 1 {
		return ThreeConfusing()
	} else {
		return TwoConfisuing()
	}
}

type Store interface {
	Set(from, to string) error
	Get(from string) (to string, err error)
	TTL(from string) (seconds int, err error)
	Exists(from string) (bool, error)
	New(to string) (from string, err error)
}

type RedisStore struct {
	conn redis.Conn
}

func NewRedisStore(address string) (Store, error) {
	conn, err := redis.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &RedisStore{conn}, nil
}

func (s *RedisStore) Set(from, to string) error {
	s.conn.Send("MULTI")
	s.conn.Send("SET", from, to)
	s.conn.Send("EXPIRE", from, TTL)
	return s.conn.Send("EXEC")
}

func (s *RedisStore) Get(from string) (to string, err error) {
	return redis.String(s.conn.Do("GET", from))
}

func (s *RedisStore) TTL(from string) (seconds int, err error) {
	ttl, err := redis.Int(s.conn.Do("TTL", from))
	if err != nil {
		return 0, err
	}
	if ttl < 0 {
		return 0, fmt.Errorf("invalid TTL")
	}
	return ttl, nil
}

func (s *RedisStore) Exists(from string) (bool, error) {
	return redis.Bool(s.conn.Do("EXISTS", from))
}

func (s *RedisStore) FindAvailableKey() (from string, err error) {
	for {
		from = MakeKey()

		var exists bool
		if exists, err = s.Exists(from); err != nil || !exists {
			return
		}
	}
}

func (s *RedisStore) New(to string) (from string, err error) {
	from, err = s.FindAvailableKey()
	if err != nil {
		return
	}
	return from, s.Set(from, to)
}

type Handler struct {
	store      Store
	fileServer http.Handler
}

func NewHandler(store Store) (Handler, error) {
	return Handler{store, http.FileServer(assetFS())}, nil
}

func (h Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	buf := new(bytes.Buffer)

	_, err := io.Copy(buf, r.Body)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"bad request; failed read request.\"}")
		return
	}

	u := buf.String()

	_, err = url.ParseRequestURI(u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"bad request; invalid URI.\"}")
		return
	}

	key, err := h.store.New(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "{\"error\":\"internal server error.\"}")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"info\":{\"short_url\":\"https://%s/%s\",\"original_url\":\"%s\",\"stats\":\"https://%s/%s/stats\",\"ttl\":%d},\"error\":null}\n", r.Host, key, u, r.Host, key, TTL)
}

func (h Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	key := strings.ToLower(r.URL.Path[1:4])

	u, err := h.store.Get(key)
	if err != nil {
		h.NotFound(w, r)
		return
	}

	http.Redirect(w, r, u, http.StatusSeeOther)
}

func (h Handler) Stats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "method not allowed.")
		return
	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "sorry, not implemented yet.")
}

func (h Handler) StatsJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.MethodNotAllowed(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	key := strings.ToLower(r.URL.Path[1:4])

	u, err := h.store.Get(key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "{\"error\":\"not found\"}")
		return
	}

	ttl, err := h.store.TTL(key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "{\"error\":\"internal server error\"}")
		return
	}

	fmt.Fprintf(w, "{\"info\":{\"short_url\":\"https://%s/%s\",\"original_url\":\"%s\",\"stats\":\"https://%s/%s/stats\",\"ttl\":%d},\"stats\":\"not implemented yet\",\"error\":null}\n", r.Host, key, u, r.Host, key, ttl)
}

func (h Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "not found")
}

func (h Handler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(w, "{\"error\":\"method not allowed\"}")
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		switch r.Method {
		case "GET":
			h.fileServer.ServeHTTP(w, r)
		case "POST":
			h.Subscribe(w, r)
		default:
			h.MethodNotAllowed(w, r)
		}
	} else if r.URL.Path == "/favicon.svg" || strings.HasPrefix(r.URL.Path, "/_nuxt/") {
		h.fileServer.ServeHTTP(w, r)
	} else if len(r.URL.Path) == 4 {
		h.Redirect(w, r)
	} else if len(r.URL.Path) == 10 && strings.HasSuffix(r.URL.Path, "/stats") {
		h.Stats(w, r)
	} else if len(r.URL.Path) == 15 && strings.HasSuffix(r.URL.Path, "/stats.json") {
		h.StatsJSON(w, r)
	} else {
		h.NotFound(w, r)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	store, err := NewRedisStore("redis:6379")
	if err != nil {
		panic(err.Error())
	}

	handler, err := NewHandler(store)
	if err != nil {
		panic(err.Error())
	}

	http.ListenAndServe(":8080", handler)
}
