package lib

import (
	"os"

	"github.com/jmoiron/sqlx"
)

// GetDBPath get db path
func GetDBPath() string {
	p, _ := os.Executable()
	return p + "/data.db"
}

// ExistDB gets whether there is db or not
func ExistDB(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CreateDBTable creates tokens Table
func CreateDBTable(db *sqlx.DB) error {
	q := `CREATE TABLE IF NOT EXISTS tokens(
		id INTEGER AUTO_INCREMENT PRIMARY KEY,
		token TEXT,
		is_refresh BOOLEAN,
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime'))
		);
		CREATE TRIGGER IF NOT EXISTS trigger_tokens_updated_at AFTER UPDATE ON tokens
		BEGIN
			UPDATE test SET updated_at = DATETIME('now', 'localtime') WHERE rowid == NEW.rowid;
		END;`
	_, err := db.Exec(q)
	return err
}

// InsertToken inserts token
func InsertToken(db *sqlx.DB, token string, isRefresh bool) error {
	q := `DELETE FROM tokens WHERE is_refresh = ?`
	if _, err := db.Exec(q, isRefresh); err != nil {
		return err
	}
	q = `INSERT INTO tokens(token, is_refresh) 
	VALUES(?, ?)`
	_, err := db.Exec(q, token, isRefresh)
	return err
}

// GetToken gets token
func GetToken(db *sqlx.DB, isRefresh bool) (string, error) {
	var token string
	q := `SELECT token FROM tokens WHERE is_refresh = ?`
	err := db.Get(&token, q, isRefresh)
	return token, err
}
