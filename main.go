package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

const database_url = "postgres://smloy:passwd@192.168.122.223:5432/ioe_admission"

var database_conn *pgx.Conn
var err error

func handleUrlEndpoints(mux *mux.Router) {
	mux.HandleFunc("/programs", programsHandler).Methods("GET")
	mux.HandleFunc("/colleges", collegeHandler).Methods("GET")
	mux.HandleFunc("/collegeprograms", collegeProgramHandler).Methods("GET")
	mux.HandleFunc("/prediction/", predictionHandler).Methods("POST")
	mux.HandleFunc("/analysis/", analysisHandler).Methods("POST")
	mux.HandleFunc("/rank/", rankHandler).Methods("POST")
	mux.HandleFunc("/district/", districtHandler).Methods("POST")
	mux.HandleFunc("/zone/", zoneHandler).Methods("POST")
}

func main() {

	database_conn, err = pgx.Connect(context.Background(), database_url)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database_conn.Close(context.Background())

	// query_func_example()

	router := mux.NewRouter()
	handleUrlEndpoints(router)

	server := http.Server{
		Addr:         ":8080",
		Handler:      router,
		WriteTimeout: 2000 * time.Millisecond,
		ReadTimeout:  2000 * time.Millisecond,
	}

	listen_err := server.ListenAndServe()
	if listen_err != nil {
		log.Fatal(listen_err)
		os.Exit(1)
	}

}

// func query_func_example() error {

// 	var (
// 		Dist_code int32
// 		Dist_name string
// 		Count     int32
// 	)

// 	rows, err := database_conn.Query(context.Background(), "select B.code,B.name,count(B.code) from admission A, district B, collegeprogram C where A.collegeprogram_id=C.id and B.code=A.district_id and C.college_id = 'PUR' and C.program_id = 'BCT' group by B.code,B.name order by count desc;")

// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
// 		return err
// 	} else {
// 		for rows.Next() {
// 			rows.Scan(&Dist_code, &Dist_name, &Count)
// 			fmt.Println(Dist_code, Dist_name, Count)
// 		}
// 	}
// 	return nil
// }
