package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type ShortLink struct {
	Key          string `json:"key"`
	Original     string `json:"original"`
	Clicks       int64  `json:"clicks"`
	CreationTime int64  `json:"creationTime"`
}

func (link *ShortLink) Fields() map[string]interface{} {
	// mapstructure is overkill
	return map[string]interface{}{
		"key":          link.Key,
		"original":     link.Original,
		"creationTime": link.CreationTime,
		"clicks":       link.Clicks,
	}
}

func NewShortLinkFromStringMap(r map[string]string) *ShortLink {
	link := NewShortLink(r["key"], r["original"])
	link.Clicks, _ = strconv.ParseInt(r["clicks"], 10, 64)
	link.CreationTime, _ = strconv.ParseInt(r["creationTime"], 10, 64)
	return link
}

func NewShortLink(key string, original string) *ShortLink {
	link := &ShortLink{
		CreationTime: time.Now().Unix(),
		Key:          key,
		Original:     original,
	}
	return link
}

type Server struct {
	redisClient *redis.Client
}

func NewServer() *Server {
	redisDest := fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port"))
	if viper.GetBool("debug") {
		log.Println("creating redis client for:", redisDest)
	}
	client := redis.NewClient(&redis.Options{
		Addr: redisDest,
		DB:   viper.GetInt("redis.db"),
	})

	return &Server{redisClient: client}
}

func (s *Server) ListenAndServe() error {
	http.HandleFunc("/", s.handleRequest)

	serverDest := fmt.Sprintf(":%d", viper.GetInt("port"))
	if viper.GetBool("debug") {
		log.Println("server starting on: ", serverDest)
	}
	return http.ListenAndServe(serverDest, nil)
}

func (s *Server) saveShortLink(link *ShortLink) error {
	return s.redisClient.HMSet(link.Key, link.Fields()).Err()
}

func (s *Server) findShortLink(key string) (*ShortLink, error) {
	k, err := s.redisClient.HExists(key, "original").Result()
	if err != nil {
		return nil, err
	}
	if !k {
		return nil, errors.New("key does not exist: " + key)
	}

	r, err := s.redisClient.HGetAll(key).Result()
	if err != nil {
		return nil, err
	}

	link := NewShortLinkFromStringMap(r)
	return link, nil
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		key := r.URL.Path[1:]
		link, err := s.findShortLink(key)
		if err != nil {
			if viper.GetBool("debug") {
				log.Println("error getting key:", key, err)
			}
			http.NotFound(w, r)
			return
		}

		s.redisClient.HIncrBy(link.Key, "clicks", 1)
		http.Redirect(w, r, link.Original, http.StatusMovedPermanently)
	case "TRACE":
		if !viper.GetBool("show_info") {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Path[1:]
		link, err := s.findShortLink(key)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(link)
		w.Write(data)

	case "POST":
		theURL := r.FormValue("url")
		valid, err := url.Parse(theURL)
		if err != nil || theURL == "" {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		b := make([]byte, viper.GetInt("entropy_bytes"))
		_, err = rand.Read(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		key := base64.URLEncoding.EncodeToString(b)
		link := NewShortLink(key, valid.String())
		if s.saveShortLink(link) != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(link)
		w.Write(data)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	viper.SetEnvPrefix("kaon")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// server
	viper.SetDefault("debug", false)
	viper.SetDefault("port", 8080)
	viper.SetDefault("entropy_bytes", 10)
	viper.SetDefault("show_info", true)

	// redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)

	// config file
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if viper.GetBool("debug") {
				log.Println("no config file provided; using environment or defaults")
			}
		} else {
			log.Fatalf("could not parse provided config file: %v", err)
		}
	}

	server := NewServer()
	server.ListenAndServe()
}
