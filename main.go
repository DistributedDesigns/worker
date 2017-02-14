package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

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

	consoleLog = logging.MustGetLogger("console")
	done       = make(chan struct{})

	redisBaseKey string
	redisPool    *redis.Pool
	rmqConn      *amqp.Connection
)

const (
	// Named RMQ queues / exchanges
	auditEventQ      = "audit_event"
	dumplogQ         = "dumplog"
	quoteRequestQ    = "quote_req"
	quoteBroadcastEx = "quote_broadcast"
)

func main() {
	// Load the config
	kingpin.Parse()
	initConsoleLogging()
	loadConfig()
	redisBaseKey = fmt.Sprintf("%s%d:", config.Redis.KeyPrefix, *workerNum)

	// Connect to external services
	initRMQ()
	defer rmqConn.Close()
	initRedis()

	// Start internal services
	initQuoteCacheRMQ()

	// Start concurrent actions
	go catchQuoteBroadcasts()
	// watch auto transactions
	// watch transaction queue
	// handle transactions

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
		`%{time:15:04:05.000} %{color}▶ %{level:8s}%{color:reset} %{id:03d} %{message}  %{shortfile}`,
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
		BaseTTL    int `yaml:"base ttl"`
		BackoffTTL int `yaml:"backoff ttl"`
	} `yaml:"quote policy"`
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
	// closed in main()
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