package db

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB

const (
	db_host    = "DB_HOST"
	db_port    = "DB_PORT"
	db_user    = "DB_USER"
	db_pass    = "DB_PASS"
	db_name    = "DB_NAME"
	table_name = "TABLE_NAME"
)

type producer struct {
	Producer       string `json:"producer"`
	Num_Benchmarks int    `json:"num_benchmarks"`
}

type producers struct {
	Total_Producers  int        `json:"total_producers"`
	Total_Benchmarks int        `json:"total_benchmarks"`
	Producers        []producer `json:"producers"`
}

type benchmark struct {
	Producer  string  `json:"producer"`
	Mean_ms   float32 `json:"mean_ms"`
	Median_ms float32 `json:"median_ms"`
	Timestamp string  `json:"timestamp"`
}

type benchmarks struct {
	Epoch      string      `json:"epoch"`
	Interval   string      `json:"interval"`
	Benchmarks []benchmark `json:"benchmarks"`
}

func init() {
	config := dbConfig()
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[db_host], config[db_port],
		config[db_user], config[db_pass], config[db_name])

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
}

func dbConfig() map[string]string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conf := make(map[string]string)

	conf[db_host] = os.Getenv(db_host)
	conf[db_port] = os.Getenv(db_port)
	conf[db_user] = os.Getenv(db_user)
	conf[db_pass] = os.Getenv(db_pass)
	conf[db_name] = os.Getenv(db_name)
	conf[table_name] = os.Getenv(table_name)

	return conf
}

func AllProducers() (producers, error) {
	response := producers{}
	total_producers := 0
	total_benchmarks := 0

	rows, err := db.Query(`
        SELECT
            producer,
            COUNT (producer) as num_benchmarks
        FROM benchmarks
        GROUP BY producer
        ORDER BY producer ASC`)

	if err != nil {
		return response, err
	}

	defer rows.Close()
	for rows.Next() {
		prod := producer{}
		err = rows.Scan(
			&prod.Producer,
			&prod.Num_Benchmarks,
		)
		if err != nil {
			return response, err
		}

		total_producers += 1
		total_benchmarks += prod.Num_Benchmarks

		response.Producers = append(response.Producers, prod)
	}

	response.Total_Producers = total_producers
	response.Total_Benchmarks = total_benchmarks

	err = rows.Err()
	if err != nil {
		return response, err
	}

	return response, err
}

func AllBenchmarks(epoch string) (benchmarks, error) {
	response := benchmarks{}
	response.Epoch = epoch
	var rows *sql.Rows
	var err error

	replacer := strings.NewReplacer("-", " ", "s", "")
	interval := replacer.Replace(epoch)

	parts := strings.Split(epoch, "-")
	timestamp := ""
	where := ""

	if strings.Contains(epoch, "day") {
		response.Interval = "1 hour"
		timestamp = "date_trunc( 'hour', created_on )"
		where = fmt.Sprintf("WHERE created_on >= (CURRENT_DATE - INTERVAL '%s')", interval)
	} else if epoch == "all" {
		response.Interval = "1 day"
		timestamp = "date_trunc( 'day', created_on )"
		where = ""
	}

	query := fmt.Sprintf(`
        SELECT producer, AVG(ALL cpu_usage_us) / 1000 as mean_ms, percentile_cont(0.5) within group (order by cpu_usage_us) / 1000 as median_ms, %s as timestamp
        FROM benchmarks
        %s
        GROUP BY producer, timestamp
        ORDER BY timestamp;`, timestamp, where)

	rows, err = db.Query(query)

	if err != nil {
		return response, err
	}

	defer rows.Close()
	for rows.Next() {
		bench := benchmark{}
		err = rows.Scan(
			&bench.Producer,
			&bench.Mean_ms,
			&bench.Median_ms,
			&bench.Timestamp,
		)
		if err != nil {
			return response, err
		}

		response.Benchmarks = append(response.Benchmarks, bench)
	}

	err = rows.Err()
	if err != nil {
		return response, err
	}

	return response, err
}
