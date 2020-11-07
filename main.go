package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"gwell-poc/user"
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

// DrillFloat ...
type DrillFloat struct {
	Depth          float32 `json:"depth"`
	E              float32 `json:"e"`
	PoissonsRation float32 `json:"poissons_ration"`
	FrictionAngle  float32 `json:"friction_angle"`
	UCS            float32 `json:"ucs"`
	CPSG           float32 `json:"cp_sg" pg:"cp_sg"`
	CPTimeSG       float32 `json:"cp_time_sg" pg:"cp_time_sg"`
	LCPSG          float32 `json:"lcp_sg" pg:"lcp_sg"`
	FPSG           float32 `json:"fp_sg" pg:"fp_sg"`
	PPSG           float32 `json:"pp_sg" pg:"pp_sg"`
	CPPSI          float32 `json:"cp_psi" pg:"cp_psi"`
	CPTimePSI      float32 `json:"cp_time_psi" pg:"cp_time_psi"`
	LCPPSI         float32 `json:"lcp_psi" pg:"lcp_psi"`
	FPPSI          float32 `json:"fp_psi" pg:"fp_psi"`
	PPPSI          float32 `json:"pp_psi" pg:"pp_psi"`
	DepthName      string  `json:"depth_name" pg:"depth_name"`
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
	csvFile, err := os.Open("./file/data-main-update.csv")
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
		if i <= 1 {
			continue
		}
		var depthName string
		maxRange := 15
		floatRec := make([]float32, 0)
		for idx, recData := range rec {
			if idx > maxRange {
				depthName = recData
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
			PPSG:           floatRec[2],
			CPSG:           floatRec[3],
			CPTimeSG:       floatRec[4],
			LCPSG:          floatRec[5],
			FPSG:           floatRec[6],
			PPPSI:          floatRec[7],
			CPPSI:          floatRec[8],
			CPTimePSI:      floatRec[9],
			LCPPSI:         floatRec[10],
			FPPSI:          floatRec[11],
			E:              floatRec[12],
			PoissonsRation: floatRec[13],
			FrictionAngle:  floatRec[14],
			UCS:            floatRec[15],
			DepthName:      depthName,
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
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var payload user.LoginRequest
		payloadDecoder := json.NewDecoder(r.Body)
		payloadDecoder.DisallowUnknownFields()

		err := payloadDecoder.Decode(&payload)
		if err != nil {
			Error(w, http.StatusBadRequest, err)
			return
		}
		success := user.Login(&payload)
		if !success {
			Error(w, http.StatusUnauthorized, nil)
			return
		}
		Success(w, http.StatusOK, nil)
	}).Methods("POST")

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
