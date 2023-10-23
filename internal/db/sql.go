package db

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DatabaseConfig struct {
	URL      string
	User     string
	Password string
}

func getDB(config DatabaseConfig) (*sqlx.DB, error) {
	return sqlx.Connect("postgres", "postgresql://"+config.User+":"+config.Password+"@"+config.URL)
}

func ConnectToSql(configFrom DatabaseConfig, configTo DatabaseConfig, filter string, tableList []string) error {

	connectionTo := "postgresql://" + configTo.User + ":" + configTo.Password + "@" + configTo.URL
	dbTo, err := getDB(configTo)
	if err != nil {
		fmt.Fprint(os.Stderr, "Unable to connect to "+connectionTo, err)
		return err
	}
	connectionFrom := "postgresql://" + configFrom.User + ":" + configFrom.Password + "@" + configFrom.URL
	dbFrom, err := getDB(configFrom)

	if err != nil {
		fmt.Fprint(os.Stderr, "Unable to connect to "+connectionFrom, err)
		return err
	}
	defer dbFrom.Close()
	defer dbTo.Close()

	for _, tableName := range tableList {
		selectQuery := fmt.Sprintf("select * from %s ", tableName)
		if len(filter) > 0 {
			selectQuery += " where " + filter
		}
		rowsFrom, err := dbFrom.Queryx(selectQuery)
		if err != nil {
			fmt.Fprint(os.Stderr, "Error executing query ", err)
			return err
		}
		insertQueries := []string{}
		baseInsert := fmt.Sprintf("insert into %s ", tableName)

		columns, err := rowsFrom.Columns()
		if err != nil {
			fmt.Fprint(os.Stderr, "Error getting columns name ", err)
			return err
		}
		fmt.Printf("columns: %v,\n", columns)
		baseInsert += "(" + strings.Join(columns, ", ") + ") values ("
		fmt.Printf("baseInsert: %v\n", baseInsert)
		for i := 0; rowsFrom.Next(); i++ {

			values, err := rowsFrom.SliceScan()
			if err != nil {
				fmt.Fprint(os.Stderr, "Error values ", err)
				return err
			}
			entities := []string{}
			for _, v := range values {
				if byteSlice, ok := v.([]uint8); ok {
					entities = append(entities, string(byteSlice))
				} else if v == nil {
					entities = append(entities, "NULL")
				} else if t, ok := v.(time.Time); ok {
					formattedTime := t.Format("2006-01-02T15:04:05.0000")

					entities = append(entities, "'"+formattedTime+"'")
				} else {
					entities = append(entities, fmt.Sprintf("'%v'", v))
				}
			}
			insertQueries = append(insertQueries, baseInsert+strings.Join(entities, " ,")+");")
			fmt.Println()
		}
		fmt.Printf("insertQueries: %v\n", insertQueries)
		for _, insert := range insertQueries {
			_, err := dbTo.Exec(insert)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ListTables(config DatabaseConfig) ([]string, error) {
	db, err := getDB(config)
	if err != nil {
		return nil, err
	}

	r, err := db.Queryx("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;")
	if err != nil {
		return nil, err
	}
	tables := []string{}
	for r.Next() {
		values, _ := r.SliceScan()

		for _, v := range values {
			tables = append(tables, fmt.Sprintf("%s", v))
		}
	}
	return tables, err

}
