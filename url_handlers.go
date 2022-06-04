package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

func programsHandler(w http.ResponseWriter, r *http.Request) {

	type Program struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	rows, err := database_conn.Query(context.Background(), "select code,name from program;")
	if err == nil {
		var progs []Program
		for rows.Next() {
			progs = append(progs, Program{})
			rows.Scan(&progs[len(progs)-1].Code, &progs[len(progs)-1].Name)
		}
		json.NewEncoder(w).Encode(progs)
	} else {
		fmt.Print(`{"detail": "Not found."}`)
	}

}

func collegeHandler(w http.ResponseWriter, r *http.Request) {

	type College struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	rows, err := database_conn.Query(context.Background(), "select code,name from college;")
	if err == nil {
		var colz []College
		for rows.Next() {
			colz = append(colz, College{})
			rows.Scan(&colz[len(colz)-1].Code, &colz[len(colz)-1].Name)
		}
		json.NewEncoder(w).Encode(colz)
	} else {
		fmt.Print(`{"detail": "Not found."}`)
	}

}

func collegeProgramHandler(w http.ResponseWriter, r *http.Request) {

	type collegeProgram struct {
		Type        string   `json:"type"`
		College     string   `json:"college"`
		Program     string   `json:"program"`
		Seats       int32    `json:"seats"`
		Cutoff      int32    `json:"cutoff"`
		Cutin       int32    `json:"cutin"`
		CutoffMarks *float32 `json:"cutoffMarks"`
		CutinMarks  *float32 `json:"cutinMarks"`
	}
	query := `SELECT A.seats,
				A.cutin,
				A.cutoff,
				A.type,
				D.NAME  AS college,
				E.NAME  AS program,
				B.score AS cutinMarks,
				C.score AS cutoffMarks
			FROM   collegeprogram AS A,
				admission AS B,
				admission AS C,
				college AS D,
				program AS E
			WHERE  B.rank = A.cutin
				AND C.rank = A.cutoff
				AND A.college_id = D.code
				AND A.program_id = E.code;`

	// query2 := `SELECT D.seats,
	// 		D.cutin,
	// 		D.cutoff,
	// 		D.type,
	// 		E.NAME AS college,
	// 		F.NAME AS program,
	// 		D.cutinmarks,
	// 		D.cutoffmarks
	// 	FROM   (SELECT A.seats,
	// 				A.cutin,
	// 				A.cutoff,
	// 				A.type,
	// 				A.college_id,
	// 				A.program_id,
	// 				B.score AS cutinMarks,
	// 				C.score AS cutoffMarks
	// 			FROM   collegeprogram AS A,
	// 				admission AS B,
	// 				admission AS C
	// 			WHERE  B.rank = A.cutin
	// 				AND C.rank = A.cutoff) AS D
	// 		INNER JOIN college AS E
	// 				ON D.college_id = E.code
	// 		INNER JOIN program AS F
	// 				ON D.program_id = F.code;`

	rows, err := database_conn.Query(context.Background(), query)

	if err == nil {
		var colz []collegeProgram
		for rows.Next() {
			colz = append(colz, collegeProgram{})
			rows.Scan(&colz[len(colz)-1].Seats, &colz[len(colz)-1].Cutin, &colz[len(colz)-1].Cutoff, &colz[len(colz)-1].Type, &colz[len(colz)-1].College, &colz[len(colz)-1].Program, &colz[len(colz)-1].CutinMarks, &colz[len(colz)-1].CutoffMarks)
		}

		json.NewEncoder(w).Encode(colz)
	} else {
		fmt.Print(`{"detail": "Not found."}`)
	}

}

func predictionHandler(w http.ResponseWriter, r *http.Request) {

	getProbabilityString := func(rank float32, cutoff float32, total_seats float32) string {
		if rank < cutoff-0.4*total_seats {
			return "very high"
		} else if rank < cutoff-0.1*total_seats {
			return "high"
		} else if rank > cutoff-0.1*total_seats && rank < cutoff+0.1*total_seats {
			return "critical"
		} else if rank < cutoff+0.3*total_seats {
			return "low"
		} else {
			return "very low"
		}
	}

	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}
	if err := r.ParseMultipartForm(64); err != nil {
		fmt.Println(err)
	}

	type queryResult struct {
		College      string `json:"college"`
		College_name string `json:"college_name"`
		Program      string `json:"program"`
		Program_name string `json:"program_name"`
		Type         string `json:"type"`
		Probablity   string `json:"probablity"`
	}

	var query strings.Builder
	query.WriteString(`SELECT
	A.type,
	D.code AS college,
	E.code AS program,
	D.name AS college_name,
	E.name AS program_name,
	A.cutoff,
	A.seats
FROM   collegeprogram AS A,
	college AS D,
	program AS E
WHERE
	A.college_id = D.code
	AND A.program_id = E.code `)

	var rows pgx.Rows
	if r.Form["college"][0] == "All" && r.Form["faculty"][0] == "All" {
		query.WriteString(";")
		rows, err = database_conn.Query(context.Background(), query.String())
	} else if r.Form["college"][0] == "All" {
		fmt.Fprintf(&query, "AND E.code =$1;")
		rows, err = database_conn.Query(context.Background(), query.String(), r.Form["faculty"][0])
	} else if r.Form["faculty"][0] == "All" {
		fmt.Fprintf(&query, "AND D.code =$1;")
		rows, err = database_conn.Query(context.Background(), query.String(), r.Form["college"][0])
	} else {
		fmt.Fprintf(&query, "AND D.code =$1 AND E.code =$2;")
		rows, err = database_conn.Query(context.Background(), query.String(), r.Form["college"][0], r.Form["faculty"][0])
	}

	if err == nil {
		var colz []queryResult
		for rows.Next() {
			colz = append(colz, queryResult{})
			var (
				cutoff, seats int32
			)
			rows.Scan(&colz[len(colz)-1].Type, &colz[len(colz)-1].College, &colz[len(colz)-1].Program, &colz[len(colz)-1].College_name, &colz[len(colz)-1].Program_name, &cutoff, &seats)
			rank, err := strconv.ParseFloat(r.Form["rank"][0], 32)
			if err != nil {
				println("error in parsing rank to float")
			}
			colz[len(colz)-1].Probablity = getProbabilityString(float32(rank), float32(cutoff), float32(seats))
		}
		json.NewEncoder(w).Encode(colz)
	} else {
		fmt.Printf(`{"detail": "Not found."} %v`, err)
	}

}

func analysisHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}
	if err := r.ParseMultipartForm(64); err != nil {
		fmt.Println(err)
	}

	type queryResult struct {
		Faculty      string `json:"faculty"`
		Type         string `json:"type"`
		LowerLimit   int32  `json:"lowerLimit"`
		UpperLimit   int32  `json:"upperLimit"`
		Seats        int32  `json:"seats"`
		College      string `json:"college"`
		Program_name string `json:"program_name"`
		College_name string `json:"college_name"`
	}

	var query strings.Builder
	query.WriteString(`SELECT
	A.type,
	D.code AS college,
	E.code AS program,
	D.name AS college_name,
	E.name AS program_name,
	A.cutoff,
	A.cutin,
	A.seats,
	A.id
FROM   collegeprogram AS A,
	college AS D,
	program AS E
WHERE
	A.college_id = D.code
	AND A.program_id = E.code `)

	if r.Form["college"][0] == "All" && r.Form["faculty"][0] != "All" {
		fmt.Fprintf(&query, "AND E.code = $1;")
		rows, err := database_conn.Query(context.Background(), query.String(), r.Form["faculty"][0])
		if err != nil {
			fmt.Println("database error", err)
		}
		var colz []queryResult
		var collegeprogram_id int32
		for rows.Next() {
			colz = append(colz, queryResult{})
			rows.Scan(&colz[len(colz)-1].Type, &colz[len(colz)-1].College, &colz[len(colz)-1].Faculty, &colz[len(colz)-1].College_name, &colz[len(colz)-1].Program_name, &colz[len(colz)-1].UpperLimit, &colz[len(colz)-1].LowerLimit, &colz[len(colz)-1].Seats, &collegeprogram_id)
		}
		json.NewEncoder(w).Encode(colz)
	} else if r.Form["college"][0] == "All" || r.Form["faculty"][0] != "All" {
		fmt.Println("Filter should be one college and all faculty")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		fmt.Fprintf(&query, "AND D.code =$1;")
		rows, err := database_conn.Query(context.Background(), query.String(), r.Form["college"][0])
		if err != nil {
			fmt.Println("database error", err)
		}

		var colz []queryResult
		for rows.Next() {
			colz = append(colz, queryResult{})
			var collegeprogram_id int32
			rows.Scan(&colz[len(colz)-1].Type, &colz[len(colz)-1].College, &colz[len(colz)-1].Faculty, &colz[len(colz)-1].College_name, &colz[len(colz)-1].Program_name, &colz[len(colz)-1].UpperLimit, &colz[len(colz)-1].LowerLimit, &colz[len(colz)-1].Seats, &collegeprogram_id)
			query_new := `select MAX(rank) from admission where collegeprogram_id = $1 and rank is not null`
			err := database_conn.QueryRow(context.Background(), query_new, collegeprogram_id).Scan(&colz[len(colz)-1].LowerLimit)
			if err != nil {
				fmt.Println("database error", err)
			}
		}
		json.NewEncoder(w).Encode(colz)
	}

}

func rankHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}
	if err := r.ParseMultipartForm(64); err != nil {
		fmt.Println(err)
	}
	min_rank := r.Form["min_rank"][0]
	max_rank := r.Form["max_rank"][0]
	college := r.Form["college"][0]

	if college != "All" {
		type queryResult struct {
			Program      string `json:"collegeprogram__program"`
			College_name string `json:"collegeprogram__college__name"`
			Program_name string `json:"collegeprogram__program__name"`
			Count        int32  `json:"count"`
		}

		query := `SELECT B.program_id,C.NAME,D.NAME,Count(*)
		FROM   admission AS A,collegeprogram AS B,college AS C,program AS D
		WHERE  A.collegeprogram_id = B.id
		AND B.college_id = C.code
		AND B.program_id = D.code
		AND B.college_id = $1
		AND A.rank >= $2
		AND A.rank <= $3
		GROUP  BY B.program_id,C.NAME,D.NAME;`

		var colz []queryResult
		rows, err := database_conn.Query(context.Background(), query, college, min_rank, max_rank)
		if err != nil {
			fmt.Println("database error: ", err)
		}

		for rows.Next() {
			colz = append(colz, queryResult{})
			rows.Scan(&colz[len(colz)-1].Program, &colz[len(colz)-1].College_name, &colz[len(colz)-1].Program_name, &colz[len(colz)-1].Count)
		}
		json.NewEncoder(w).Encode(colz)
	}

	type queryResult struct {
		College      string `json:"collegeprogram__college"`
		College_name string `json:"collegeprogram__college__name"`
		Count        int32  `json:"count"`
	}

	query := `SELECT B.college_id,C.name,Count(*)
	FROM   admission AS A,collegeprogram AS B,college AS C
	WHERE  A.collegeprogram_id = B.id
	AND B.college_id = C.code
	AND A.rank >= $1
	AND A.rank <= $2
	GROUP  BY B.college_id,C.name;`

	var colz []queryResult
	rows, err := database_conn.Query(context.Background(), query, min_rank, max_rank)
	if err != nil {
		fmt.Println("database error: ", err)
	}

	for rows.Next() {
		colz = append(colz, queryResult{})
		rows.Scan(&colz[len(colz)-1].College, &colz[len(colz)-1].College_name, &colz[len(colz)-1].Count)
	}
	json.NewEncoder(w).Encode(colz)
}

func districtHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}
	if err := r.ParseMultipartForm(64); err != nil {
		fmt.Println(err)
	}
	faculty := r.Form["faculty"][0]
	college := r.Form["college"][0]

	// var responseData map[string]string
	type loc struct {
		Dist_code int32  `json:"district__code"`
		Dist_name string `json:"district__name"`
		Count     int32  `json:"count"`
	}

	type response struct {
		College  string `json:"college"`
		Program  string `json:"program"`
		Location []loc  `json:"location"`
	}

	resposeData := response{}
	var rows pgx.Rows

	if college == "All" && faculty == "All" {
		resposeData.College = "All colleges"
		resposeData.Program = "All programs"
		rows, err = database_conn.Query(context.Background(), "select B.code,B.name,count(B.code) from admission A,district B where A.district_id=B.code group by B.code,B.name order by count desc;")
		if err != nil {
			fmt.Println("database error: ", err)
		}

	} else if college == "All" {
		resposeData.College = "All colleges"
		if err = database_conn.QueryRow(context.Background(), "select name from program where code=$1;", faculty).Scan(&resposeData.Program); err != nil {
			fmt.Println("database error: ", err)
		}

		rows, err = database_conn.Query(context.Background(), "select B.code,B.name,count(B.code) from admission A,district B,collegeprogram C where A.district_id=B.code and A.collegeprogram_id=C.id and C.program_id=$1 group by B.code,B.name order by count desc;", faculty)
		if err != nil {
			fmt.Println("database error: ", err)
		}

	} else if faculty == "All" {
		resposeData.Program = "All programs"

		if err = database_conn.QueryRow(context.Background(), "select name from college where code=$1;", college).Scan(&resposeData.College); err != nil {
			fmt.Println("database error: ", err)
		}

		rows, err = database_conn.Query(context.Background(), "select B.code,B.name,count(B.code) from admission A, district B, collegeprogram C where A.collegeprogram_id=C.id and B.code=A.district_id and C.college_id= $1 group by B.code,B.name order by count desc;", college)
		if err != nil {
			fmt.Println("database error: ", err)
		}

	} else {

		if err = database_conn.QueryRow(context.Background(), "select name from college where code=$1;", college).Scan(&resposeData.College); err != nil {
			fmt.Println("database error: ", err)
		}
		if err = database_conn.QueryRow(context.Background(), "select name from program where code=$1;", faculty).Scan(&resposeData.Program); err != nil {
			fmt.Println("database error: ", err)
		}
		rows, err = database_conn.Query(context.Background(), "select B.code,B.name,count(B.code) from admission A, district B, collegeprogram C where A.collegeprogram_id=C.id and B.code=A.district_id and C.college_id = $1 and C.program_id = $2 group by B.code,B.name order by count desc;", college, faculty)
		if err != nil {
			fmt.Println("database error: ", err)
		}

	}

	for rows.Next() {
		resposeData.Location = append(resposeData.Location, loc{})
		err = rows.Scan(&resposeData.Location[len(resposeData.Location)-1].Dist_code, &resposeData.Location[len(resposeData.Location)-1].Dist_name, &resposeData.Location[len(resposeData.Location)-1].Count)
		if err != nil {
			fmt.Println("error while scanning ", err)
		}
	}
	json.NewEncoder(w).Encode(resposeData)

}

func zoneHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}
	if err := r.ParseMultipartForm(64); err != nil {
		fmt.Println(err)
	}
	faculty := r.Form["faculty"][0]
	college := r.Form["college"][0]

	type loc struct {
		Zone_code int32  `json:"district__zone__id"`
		Zone_name string `json:"district__zone__name"`
		Count     int32  `json:"count"`
	}

	var location []loc

	var rows pgx.Rows

	if college == "All" && faculty == "All" {
		rows, err = database_conn.Query(context.Background(), "select D.id,D.name,count(D.id) from admission A, district B, collegeprogram C, zone D where A.collegeprogram_id=C.id and B.code=A.district_id and B.zone_id=D.id group by D.id,D.name order by count desc;")

		if err != nil {
			fmt.Println("database error: ", err)
		}

	} else if college == "All" {
		rows, err = database_conn.Query(context.Background(), "select D.id,D.name,count(D.id) from admission A, district B, collegeprogram C, zone D where A.collegeprogram_id=C.id and B.code=A.district_id and B.zone_id=D.id and C.program_id = $1 group by D.id,D.name order by count desc;", faculty)

		if err != nil {
			fmt.Println("database error: ", err)
		}

	} else if faculty == "All" {
		rows, err = database_conn.Query(context.Background(), "select D.id,D.name,count(D.id) from admission A, district B, collegeprogram C, zone D where A.collegeprogram_id=C.id and B.code=A.district_id and B.zone_id=D.id and C.college_id = $1 group by D.id,D.name order by count desc;", college)

		if err != nil {
			fmt.Println("database error: ", err)
		}

	} else {
		rows, err = database_conn.Query(context.Background(), "select D.id,D.name,count(D.id) from admission A, district B, collegeprogram C, zone D where A.collegeprogram_id=C.id and B.code=A.district_id and B.zone_id=D.id and C.college_id = $1 and C.program_id = $2 group by D.id,D.name order by count desc;", college, faculty)

		if err != nil {
			fmt.Println("database error: ", err)
		}
	}
	for rows.Next() {
		location = append(location, loc{})
		err = rows.Scan(&location[len(location)-1].Zone_code, &location[len(location)-1].Zone_name, &location[len(location)-1].Count)
		if err != nil {
			fmt.Println("error while scanning ", err)
		}
	}
	json.NewEncoder(w).Encode(location)

}
