package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/fzzy/radix/redis"
    "net/http"
    "net/url"
    "os"
    "runtime"
    "strconv"
    "strings"
    "time"
)

const (
    COUNTER = "__kaon_counter__"
    INFO_SUFFIX = "_"
)

var (
    client *redis.Client
)

type Kaon struct {
    Key          string
    Original     string
    Clicks       int64
    CreationTime int64
}

func (k Kaon) Json() []byte {
    m, _ := json.Marshal(k)
    return m
}

func NewKaon(key string, original string) *Kaon {
    kaon := new(Kaon)
    kaon.CreationTime = time.Now().UnixNano()
    kaon.Key = key
    kaon.Original = original
    kaon.Clicks = 0
    return kaon
}

func FindKaon(key string) (*Kaon, error) {
    k, _ := client.Cmd("HEXISTS", key, "Original").Bool()
    if !k {
        return nil, errors.New("key does not exist: " + key)
    } else {
        kaon := new(Kaon)
        kaon.Key = key
        r, _ := client.Cmd("HGETALL", key).Hash()
        kaon.Original = r["Original"]
        kaon.Clicks, _ = strconv.ParseInt(r["Clicks"], 10, 64)
        kaon.CreationTime, _ = strconv.ParseInt(r["CreationTime"], 10, 64)
        return kaon, nil
    }
}

func SaveKaon(key string, original string) *Kaon {
    kaon := NewKaon(key, original)
    go client.Cmd("HMSET", kaon.Key,
                  "Original", kaon.Original,
                  "CreationTime", kaon.CreationTime,
                  "Clicks", kaon.Clicks)
    return kaon
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
        case "GET":
            key := r.URL.Path[1:]
            info := strings.HasSuffix(key, INFO_SUFFIX)
            key = strings.Replace(key, INFO_SUFFIX, "", 1)
            if key == "" {
                w.WriteHeader(200)
            } else {
                kaon, err := FindKaon(key)
                if err == nil {
                    if info {
                        w.Header().Set("Content-Type", "application/json")
                        r := kaon.Json()
                        w.Write(r)
                    } else {
                        client.Cmd("HINCRBY", kaon.Key, "Clicks", 1)
                        http.Redirect(w, r, kaon.Original, http.StatusMovedPermanently)
                    }
                } else {
                    http.NotFound(w, r)
                }
            }
        case "POST":
            theUrl := r.FormValue("url")
            valid, err := url.Parse(theUrl)
            if err != nil || theUrl == "" {
                http.Error(w, "Invalid URL", http.StatusBadRequest)
            } else {
                counter, _ := client.Cmd("INCR", COUNTER).Int64()
                key := strconv.FormatInt(counter, 36)
                kaon := SaveKaon(key, valid.String())

                w.Header().Set("Content-Type", "application/json")
                r := kaon.Json()
                w.Write(r)
            }
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU()+1)

    redisHost := os.Getenv("KAON_REDIS_HOST")
    if redisHost == "" {
        redisHost = "localhost"
    }

    redisPort := os.Getenv("KAON_REDIS_PORT")
    if redisPort == "" {
        redisPort = "6379"
    }

    redisDb, err := strconv.Atoi(os.Getenv("KAON_REDIS_DB"))
    if err != nil {
        redisDb = 0
    }

    redisDest := fmt.Sprintf("%s:%s", redisHost, redisPort)

    listenPort := os.Getenv("KAON_PORT")
    if listenPort == "" {
        listenPort = "8080"
    }

    client, _ = redis.DialTimeout("tcp", redisDest, time.Duration(10)*time.Second)
    client.Cmd("SELECT", redisDb)

    http.HandleFunc("/", HandleRequest)
    http.ListenAndServe(":"+listenPort, nil)
}
