package db

import (
	"log"

	"github.com/BurntSushi/migration"
)

var Migrator []migration.Migrator

func initMigrator() {
	log.Printf("init db.Migrator")
	Migrator = []migration.Migrator{
		func(tx migration.LimitedTx) (err error) {
			_, err = tx.Exec(`
create table if not exists requests(
  created_at datetime default current_timestamp,
  status int not null,
  ip text not null,
  alias text not null,
  target text not null
)`)
			return
		},
	}
}
