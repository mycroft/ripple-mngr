package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
)

func DbConnect(dbHost, dbName, dbUser, dbPass string) mysql.Conn {
	db := mysql.New("tcp", "", fmt.Sprintf("%s:3306", dbHost), dbUser, dbPass, dbName)
	err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func CreateTable(db mysql.Conn) error {
	query := `CREATE TABLE xrpkeys(
                id INT NOT NULL AUTO_INCREMENT,
                pub CHAR(40) NOT NULL,
                used BOOL NOT NULL DEFAULT false,
                completed BOOL NOT NULL DEFAULT false,
                tx_metadata TEXT,
                tx_value NUMERIC(32) DEFAULT 0,
                received NUMERIC(32) DEFAULT 0,
                started_ts TIMESTAMP DEFAULT '2018-01-01 00:00:00',
                PRIMARY KEY(id));`

	_, _, err := db.Query(query)
	if err != nil {
		return err
	}

	query = `CREATE TABLE xrptx(
                id INT NOT NULL AUTO_INCREMENT,
                hash TEXT,
                from_addr TEXT,
                to_addr TEXT,
                amount NUMERIC(32),
                PRIMARY KEY(id));`

	_, _, err = db.Query(query)
	if err != nil {
		return err
	}

	return nil
}

func Store(db mysql.Conn, pub string) error {
	if fDebug {
		log.Printf("DB Store: %s\n", pub)
	}

	stmt, err := db.Prepare("INSERT INTO xrpkeys(pub) VALUES(?)")
	if err != nil {
		return err
	}

	res, err := stmt.Run(pub)
	if err != nil {
		return err
	}

	if fDebug {
		log.Printf("DB Store: returned id:%d\n", res.InsertId())
	}

	return nil
}

func StoreTXs(db mysql.Conn, txs []TX) error {
	if fDebug {
		log.Printf("DB StoreTXs: %d records\n", len(txs))
	}

	stmt, err := db.Prepare(`INSERT IGNORE INTO xrptx(hash, from_addr, to_addr, amount)
        VALUE(?, ?, ?, ?)`)

	if err != nil {
		return err
	}

	for _, tx := range txs {
		if fDebug {
			log.Printf("DB StoreTXs: Storing tx hash:%s\n", tx.Hash)
		}

		res, err := stmt.Run(
			tx.Hash,
			tx.Account,
			tx.Destination,
			tx.Amount,
		)

		if err != nil {
			return err
		}

		if fDebug {
			log.Printf("DB StoreTXs: returned id:%d\n", res.InsertId())
		}

	}

	if fDebug {
		log.Printf("DB StoreTXs Done.\n")
	}

	return nil
}

func GetStoreStatus(db mysql.Conn) int {
	rows, _, err := db.Query("SELECT COUNT(*) as count FROM xrpkeys WHERE used = false;")
	if err != nil {
		log.Fatal(err)
	}

	if len(rows) != 1 {
		return 0
	}

	return rows[0].Int(0)
}

func UpdateValue(db mysql.Conn, pub string, value big.Int) error {
	stmt, err := db.Prepare("UPDATE xrpkeys SET received = ? WHERE pub = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Run(value.Text(10), pub)
	if err != nil {
		return err
	}

	if fDebug {
		log.Printf("UpdateValue pub:%s value:%s completed without error\n", pub, value.Text(10))
	}

	return nil
}
