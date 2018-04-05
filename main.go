package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"

	"gopkg.in/ini.v1"

	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
)

var (
	fConfigFile            string
	fDebug                 bool
	file                   string
	fInit, fWatch, fStatus bool
	fApiUrl                string
	fFile                  string
	fRefresh               bool
)

func init() {
	flag.BoolVar(&fDebug, "debug", false, "Debug mode")
	flag.StringVar(&fConfigFile, "config", "config.ini", "Configuration file")
	flag.StringVar(&file, "file", "", "File for export")
	flag.BoolVar(&fWatch, "watch", false, "Search for transactions for existing addresses")
	flag.BoolVar(&fInit, "init", false, "DB Init")
	flag.BoolVar(&fStatus, "status", false, "Show key statuses")
	flag.BoolVar(&fRefresh, "refresh", false, "Refresh data from database")
}

func HttpQuery(url string) ([]byte, error) {
	if fDebug {
		log.Printf("HTTP: Doing HTTP Query: %s\n", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	if fDebug {
		log.Printf("HTTP: Got result %s\n", string(body))
	}

	return body, nil
}

func Watch(db mysql.Conn) error {
	var current_received big.Int
	var pubkey string

	if fDebug {
		log.Println("Watch()")
	}

	query := "SELECT pub, received FROM xrpkeys WHERE tx_value > received AND NOW() < started_ts + INTERVAL 1 DAY AND completed = false AND received < tx_value;"

	if fRefresh {
		query = "SELECT pub, received FROM xrpkeys WHERE used = true and completed = false;"
	}

	rows, _, err := db.Query(query)
	if err != nil {
		return err
	}

	if fDebug && len(rows) == 0 {
		log.Printf("No record to look after.")
	}

	for _, row := range rows {
		pubkey = row.Str(0)

		if fDebug {
			log.Printf("Looking for key 0x%s\n", pubkey)
		}

		value, err := QueryBalance(pubkey)
		if err != nil {
			return err
		}

		current_received.UnmarshalText([]byte(row.Str(1)))

		if fDebug {
			log.Printf("Current: pub:%s received:%s\n", pubkey, current_received.Text(10))

		}

		if value.Cmp(&current_received) != 0 || fRefresh {
			if fDebug {
				log.Printf("Storing new received value (%s) in database.\n", value.Text(10))
			}

			err := UpdateValue(db, pubkey, value)
			if err != nil {
				return err
			}

			if fDebug {
				log.Printf("Query for TX for 0x%s\n", pubkey)
			}

			txs, err := QueryTx(pubkey)
			if err != nil {
				return err
			}

			err = StoreTXs(db, txs)
			if err != nil {
				return err
			}
		} else {
			if fDebug {
				log.Printf("No balance change.\n")
			}
		}
	}

	if fDebug {
		log.Println("Watch() done successfully!")
	}

	return nil
}

func ShowStatus(db mysql.Conn) error {
	rows, _, err := db.Query("SELECT id, pub, used, tx_value, received, started_ts FROM xrpkeys WHERE completed = false ORDER BY ID ASC;")
	if err != nil {
		return err
	}

	for _, row := range rows {
		var started_ts_str string
		if row.Str(5) != "2018-01-01 00:00:00" {
			started_ts_str = fmt.Sprintf("started_ts:'%s'", row.Str(5))
		}

		log.Printf("id:%d 0x%s used:%v waited:%d received:%d %s\n", row.Int(0), row.Str(1), row.Bool(2), row.Int(3), row.Int(4), started_ts_str)
	}

	return nil
}

func main() {
	var db mysql.Conn
	var fd *os.File
	var err error

	flag.Parse()

	cfg, err := ini.Load(fConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	if !fDebug {
		fDebug = cfg.Section("general").Key("debug").MustBool(false)
	}

	pool_num := cfg.Section("keys").Key("num").MustInt(20)

	if fDebug {
		log.Printf("Key pool size is %d.\n", pool_num)
	}

	fApiUrl = cfg.Section("general").Key("api_url").MustString("https://s.altnet.rippletest.net:51234")

	if fDebug {
		log.Printf("Using API url: %s\n", fApiUrl)
	}

	if file == "" {
		file = cfg.Section("general").Key("file").MustString("./private-keys")
	}

	if fDebug {
		log.Printf("Using file: %s\n", file)
	}

	is_disabled := cfg.Section("db").Key("disabled").MustBool(false)
	if err != nil {
		fmt.Printf("Invalid value for disabled field: %s\n", cfg.Section("db").Key("disabled").String())
		os.Exit(1)
	}

	if is_disabled == false {
		db = DbConnect(
			cfg.Section("db").Key("host").String(),
			cfg.Section("db").Key("name").String(),
			cfg.Section("db").Key("user").String(),
			cfg.Section("db").Key("pass").String(),
		)
		defer db.Close()

		if fInit {
			err := CreateTable(db)
			if err != nil {
				log.Fatal(err)
			} else {
				log.Println("Tables created.")
			}

			return
		}
	}

	if file != "" {
		fd, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Panic(err)
		}
		defer fd.Close()
	}

	if fStatus {
		ShowStatus(db)
		return
	}

	if fWatch {
		err := Watch(db)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	db_num := GetStoreStatus(db)

	if pool_num <= db_num {
		log.Printf("No need to insert new records (%d keys in DB).\n", db_num)
		return
	}

	required_num := pool_num - db_num

	if fDebug {
		log.Printf("Required to create %d new keys.\n", required_num)
	}

	for i := 0; i < required_num; i++ {
		key, seed, err := GenerateKey()
		if err != nil {
			log.Panic(err)
		}

		address, err := GetAddress(key)
		if err != nil {
			log.Panic(err)
		}

		if fDebug {
			log.Printf("Seed: %s\n", seed)
			log.Printf("Pub: %s\n", address)
			// log.Printf("Pub:  0x%x\n", hash[12:])
		}

		if fd != nil {
			fd.WriteString(fmt.Sprintf("%s;%s\n", address, seed))
		}

		if db != nil {
			Store(db, address)
		}
	}
}
