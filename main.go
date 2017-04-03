package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	types "github.com/distributeddesigns/shared_types"
	"github.com/garyburd/redigo/redis"
	logging "github.com/op/go-logging"
	"github.com/streadway/amqp"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	yaml "gopkg.in/yaml.v2"
)

// Globals
var (
	logLevels = []string{"CRITICAL", "ERROR", "WARNING", "NOTICE", "INFO", "DEBUG"}
	logLevel  = kingpin.
			Flag("log-level", fmt.Sprintf("Minimum level for logging to the console. Must be one of: %s", strings.Join(logLevels, ", "))).
			Default("WARNING").
			Short('l').
			Enum(logLevels...)
	workerNum = kingpin.
			Flag("worker-num", "Worker ID number").
			Required().
			Short('n').
			Int()
	configFile = kingpin.
			Flag("config", "YAML file with service config").
			Default("./config/dev.yaml").
			Short('c').
			ExistingFile()
	noAudit = kingpin.
		Flag("no-audit", "Do not send messages to the audit server").
		Short('a').
		Bool()

	accountStore = make(map[string]*account)
	workATXStore = make(map[types.AutoTxKey]types.AutoTxInit)

	consoleLog = logging.MustGetLogger("console")
	done       = make(chan struct{})

	redisBaseKey string
	pendingTxKey string
	redisPool    *redis.Pool
	rmqConn      *amqp.Connection
)

const (
	// Named RMQ queues / exchanges
	auditEventQ      = "audit_event"
	dumplogQ         = "dumplog"
	quoteRequestQ    = "quote_req"
	quoteBroadcastEx = "quote_broadcast"
	autoTxQueue      = "autoTx"
	autoTxExchange   = "autoTx_resolved"

	// Redis settings
	pendingTxTimeout = 3
)

var autoTxInitChan = make(chan types.AutoTxInit)
var autoTxCancelChan = make(chan types.AutoTxKey)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Load the config
	kingpin.Parse()
	initConsoleLogging()
	loadConfig()
	redisBaseKey = fmt.Sprintf("%s:%d", config.Redis.KeyPrefix, *workerNum)
	pendingTxKey = redisBaseKey + ":pendingTx"
	// Connect to external services
	initRMQ()
	defer rmqConn.Close()
	initRedis()

	// Start internal services
	initQuoteCacheRMQ()

	// Capped channel of tx pulled out of redis
	unprocessedTxs := make(chan string, 2)

	// open http connections
	go incomingTxWatcher()

	go sendAutoTx()
	go receiveAutoTx()

	// Start concurrent actions
	go catchQuoteBroadcasts()
	// watch auto transactions
	go fetchNewTx(unprocessedTxs)
	go txWorker(unprocessedTxs)
	go cleanAccountStore()

	// halt until channel is closed
	<-done
}

func failOnError(err error, msg string) {
	if err != nil {
		consoleLog.Fatalf("%s: %s", msg, err)
	}
}

func initConsoleLogging() {

	// Create a default backend
	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)

	// Add output formatting
	var consoleFormat = logging.MustStringFormatter(
		`%{time:15:04:05.000} %{color}▶ %{level:8s}%{color:reset} %{id:03d} %{message} (%{shortfile})`,
	)
	consoleBackendFormatted := logging.NewBackendFormatter(consoleBackend, consoleFormat)

	// Add leveled logging
	level, err := logging.LogLevel(*logLevel)
	if err != nil {
		fmt.Println("Bad log level. Using default level of ERROR")
	}
	consoleBackendFormattedAndLeveled := logging.AddModuleLevel(consoleBackendFormatted)
	consoleBackendFormattedAndLeveled.SetLevel(level, "")

	// Attach the backend
	logging.SetBackend(consoleBackendFormattedAndLeveled)
}

// Holds values from <config>.yaml.
// 'PascalCase' values come from 'pascalcase' in x.yaml
var config struct {
	Rabbit struct {
		Host string
		Port int
		User string
		Pass string
	}

	Redis struct {
		Host        string
		Port        int
		MaxIdle     int    `yaml:"max idle connections"`
		MaxActive   int    `yaml:"max active connections"`
		IdleTimeout int    `yaml:"idle timeout"`
		KeyPrefix   string `yaml:"key prefix"`
	}

	QuotePolicy struct {
		BaseTTL      int   `yaml:"base ttl"`
		BackoffTTL   int   `yaml:"backoff ttl"`
		MinTTL       int   `yaml:"min ttl"`
		UseInBuySell int64 `yaml:"use in buy sell"`
	} `yaml:"quote policy"`

	CleanupInterval int `yaml:"cleanup interval"`
}

func loadConfig() {
	// Load the yaml file
	data, err := ioutil.ReadFile(*configFile)
	failOnError(err, "Could not read file")

	err = yaml.Unmarshal(data, &config)
	failOnError(err, "Could not unmarshal config")
}

func initRMQ() {
	rabbitAddress := fmt.Sprintf("amqp://%s:%s@%s:%d",
		config.Rabbit.User, config.Rabbit.Pass,
		config.Rabbit.Host, config.Rabbit.Port,
	)

	var err error
	rmqConn, err = amqp.Dial(rabbitAddress)
	failOnError(err, "Failed to rmqConnect to RabbitMQ")
}

func initRedis() {
	redisAddress := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)

	redisPool = &redis.Pool{
		MaxIdle:     config.Redis.MaxIdle,
		MaxActive:   config.Redis.MaxActive,
		IdleTimeout: time.Second * time.Duration(config.Redis.IdleTimeout),
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", redisAddress) },
	}

	// Test if we can talk to redis
	conn := redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	failOnError(err, "Could not establish connection with Redis")
}

// cleanAccountStore removes expired buy/sells from all accounts
func cleanAccountStore() {
	cleanupTicker := time.NewTicker(time.Second * time.Duration(config.CleanupInterval))

	for {
		select {
		case <-cleanupTicker.C:
			consoleLog.Debug("Starting account cleanup")
			for _, acct := range accountStore {
				acct.PruneExpiredTxs()
			}
			consoleLog.Debug("Done account cleanup")
		default:
			// Waiting for tick
		}
	}
}
