package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/gorilla/mux"
)

// GetAllDrill ...
func GetAllDrill(db *pg.DB) (drillFloats []*DrillFloat, err error) {
	start := time.Now()
	err = db.Model(&drillFloats).Select()
	end := time.Now()
	timeDiff := end.Sub(start)
	fmt.Println(timeDiff.Seconds())
	return
}

// Drill
type Drill struct {
	Depth          string  `json:"depth"`
	DT             string  `json:"dt"`
	RHOB           string  `json:"rhob"`
	OmegaV         string  `json:"omega_v"`
	OmegaHCaps     string  `json:"omega_h_caps"`
	OmegaH         string  `json:"omega_h"`
	E              float32 `json:"e"`
	PoissonsRation string  `json:"poissons_ration"`
	FrictionAngle  string  `json:"friction_angle"`
	UCS            string  `json:"ucs"`
	CP             string  `json:"cp"`
	CPTimer        string  `json:"cp_timer"`
	LCP            string  `json:"lcp"`
	FP             string  `json:"fp"`
}

// DrillFloat ...
type DrillFloat struct {
	Depth          float32 `json:"depth"`
	DT             float32 `json:"dt"`
	RHOB           float32 `json:"rhob"`
	OmegaV         float32 `json:"omega_v" pg:"omega_v"`
	OmegaHCaps     float32 `json:"omega_h_caps" pg:"omega_h_caps"`
	OmegaH         float32 `json:"omega_h" pg:"omega_h"`
	E              float32 `json:"e"`
	PoissonsRation float32 `json:"poissons_ration"`
	FrictionAngle  float32 `json:"friction_angle"`
	UCS            float32 `json:"ucs"`
	CP             float32 `json:"cp"`
	CPTimer        float32 `json:"cp_timer"`
	LCP            float32 `json:"lcp"`
	FP             float32 `json:"fp"`
}

// Success ...
func Success(w http.ResponseWriter, status int, data interface{}) {
	resp := map[string]interface{}{
		"data":  data,
		"error": nil,
	}
	js, err := json.Marshal(resp)
	if err != nil {
		resp := map[string]interface{}{
			"data":  nil,
			"error": fmt.Sprintf("%s", err),
		}
		js, _ = json.Marshal(resp)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Error ...
func Error(w http.ResponseWriter, status int, err error) {
	var errResp map[string]interface{}
	if err != nil {
		errCode := http.StatusInternalServerError
		errMsg := err.Error()

		errResp = map[string]interface{}{
			"code":    errCode,
			"message": errMsg,
		}
	}
	resp := map[string]interface{}{
		"data":  nil,
		"error": errResp,
	}
	js, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	start := time.Now()
	csvFile, err := os.Open("./file/data-main.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()
	r := csv.NewReader(csvFile)
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var drillFloats []DrillFloat
	for i, rec := range records {
		if i == 0 {
			continue
		}
		maxRange := 13
		floatRec := make([]float32, 0)
		for idx, recData := range rec {
			if idx > maxRange {
				break
			}
			trimmed := strings.TrimSpace(recData)
			floatData, err := strconv.ParseFloat(trimmed, 32)
			if err != nil {
				panic(err)
			}
			floatRec = append(floatRec, float32(floatData))
		}
		drillFloat := DrillFloat{
			Depth:          floatRec[0],
			DT:             floatRec[1],
			RHOB:           floatRec[2],
			OmegaV:         floatRec[3],
			OmegaHCaps:     floatRec[4],
			OmegaH:         floatRec[5],
			E:              floatRec[6],
			PoissonsRation: floatRec[7],
			FrictionAngle:  floatRec[8],
			UCS:            floatRec[9],
			CP:             floatRec[10],
			CPTimer:        floatRec[11],
			LCP:            floatRec[12],
			FP:             floatRec[13],
		}
		drillFloats = append(drillFloats, drillFloat)
	}

	// Init Database
	db := pg.Connect(&pg.Options{
		Addr:     os.Getenv("PG_ADDRESS"),
		User:     os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		Database: os.Getenv("PG_DATABASE"),
	})
	fmt.Println("Success connect to db")
	_, err = db.Exec("TRUNCATE TABLE drill_floats CASCADE")
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer tx.Rollback()

	for _, drillFloat := range drillFloats {
		_, err = tx.Model(&drillFloat).Insert()
		if err != nil {
			panic(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	end := time.Now()
	timeDiff := end.Sub(start)
	fmt.Println("Success")
	fmt.Println(timeDiff.Seconds())

	router := mux.NewRouter()
	router.HandleFunc("/get-all-drill", func(w http.ResponseWriter, r *http.Request) {
		drills, err := GetAllDrill(db)
		if err != nil {
			Error(w, http.StatusInternalServerError, err)
			return
		}
		Success(w, 200, drills)
	}).Methods("GET")

	port := fmt.Sprint(":", 9500)
	addr := flag.String("addr", port, "http service address")
	s := &http.Server{
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         *addr,
		Handler:      router,
	}

	fmt.Printf("Starting API server at %s\n", *addr)
	err = s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
