// main.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dmotylev/goproperties"
	"github.com/fimston/connection-string-builder"
	"github.com/fimston/logger-backend-go.git/dao"
	"github.com/fimston/logger-backend-go.git/destination"
	"github.com/go-martini/martini"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"time"
)

type AccountInfo struct {
	userId int64
}

type IndexHead struct {
	Index string `json:"_index"`
	Type  string `json:"_type"`
}

type LogRecord struct {
	Prog   string `json:"prog"`
	Text   string `json:"text"`
	Tstamp int    `json:"tstamp"`
	User   string `json:"user"`
	Pid    int    `json:"pid"`
}

type BulkRecord struct {
	Index IndexHead `json:"index"`
}

const AmqpReconnectionInterval = "60s"

var (
	apiKeysMap      dao.ApiKeyMap
	accountsDao     dao.AccountsDao
	props           properties.Properties
	reconnectDelay  time.Duration
	dest            destination.Destination
	configPath      = flag.String("config", "", "path to configuration file")
	showVersion     = flag.Bool("version", false, "show application version and exit")
	showHelp        = flag.Bool("help", false, "show help")
	version         string
	applicationName string
)

func usage() {
	fmt.Printf("%s parameters:\n", applicationName)
	fmt.Println("\t--config=XXX\t - specify path to config file (required)")
	fmt.Println("\t--help\t\t - show this message")
	fmt.Println("\t--version\t - show application version and exit")
}

func checkArgs() {
	if *showVersion {
		fmt.Printf("%s version %s\n", applicationName, version)
		os.Exit(0)
	}
	if *showHelp {
		usage()
		os.Exit(0)
	}
	if *configPath == "" {
		usage()
		log.Fatal("Please specify path to configuration file")
	}
}

func initDAO(config *properties.Properties) (*dao.PgAccountsDao, error) {
	connBuilder, err := connstring.CreateBuilder(connstring.ConnectionStringPg)
	if err != nil {
		return nil, err
	}
	connBuilder.Address(config.String("database.addr", ""))
	connBuilder.Port(uint16(config.Int("database.port", 5432)))
	connBuilder.Username(config.String("database.username", ""))
	connBuilder.Password(config.String("database.password", ""))
	connBuilder.Dbname(config.String("database.dbname", ""))

	accountsDao := dao.NewPgAccountsDao(connBuilder.Build())
	return accountsDao, nil
}

func initREST(config *properties.Properties) (*martini.Martini, error) {
	m := martini.New()
	os.Setenv("PORT", config.String("http.port", "3000"))
	m.Use(checkAccount)

	r := martini.NewRouter()
	r.Post("/:type", postHandler)

	m.Action(r.Handle)
	return m, nil

}

func checkAccount(res http.ResponseWriter, req *http.Request) {
	apiKey := req.Header.Get("X-Api-Key")
	account, ok := apiKeysMap[apiKey]
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
	}
	req.Header.Set("Index", account.IndexAlias())
}

func postRequestIsValid(body []byte) bool {
	return true
}

func postHandler(params martini.Params, res http.ResponseWriter, req *http.Request) string {
	data, err := ioutil.ReadAll(req.Body)
	log.Println(string(data))
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return fmt.Sprint(err)
	}
	if !postRequestIsValid(data) {
		res.WriteHeader(http.StatusBadRequest)
		return ""
	}
	head, err := json.Marshal(&BulkRecord{IndexHead{req.Header.Get("Index"), params["type"]}})

	buf := bytes.NewBuffer(head)
	buf.Write([]byte("\n"))
	buf.Write(data)
	buf.Write([]byte("\n"))
	if err = dest.Push(buf.Bytes()); err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return ""
	}

	return "OK"
}

func main() {
	flag.Parse()

	apiKeysMap = make(dao.ApiKeyMap)

	w, err := syslog.New(syslog.LOG_INFO, applicationName)
	if err != nil {
		log.Fatalf("connecting to syslog: %s", err)
	}

	log.SetOutput(w)
	log.SetFlags(0)

	checkArgs()

	props, err = properties.Load(*configPath)

	if err != nil {
		log.Fatal(err)
	}

	accountsDao, err = initDAO(&props)
	if err != nil {
		log.Fatal(err)
	}

	rest, err := initREST(&props)
	if err != nil {
		log.Fatal(err)
	}

	dest, err = destination.NewDestinations(&props)
	if err != nil {
		log.Fatal(err)
	}

	err = accountsDao.LoadAccountsByApiKey(apiKeysMap)
	if err != nil {
		log.Fatal(err)
	}
	rest.Run()
}
