package custom

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"testing"
)

func TestConnect(t *testing.T) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 4321, "postgres", "postgres", "postgres")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	rows, err := db.Query("SELECT 1 WHERE 1=1")
	if err != nil {
		panic(err)
	}

	if rows.Next() {
		var val int
		err = rows.Scan(&val)
		if err != nil {
			panic(err)
		}

		fmt.Printf("value is %d", 1)
	}

	defer db.Close()
}
