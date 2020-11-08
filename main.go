package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"gwell-poc/user"
	"io"
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
	// Init Database
	db := pg.Connect(&pg.Options{
		Addr:     os.Getenv("PG_ADDRESS"),
		User:     os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		Database: os.Getenv("PG_DATABASE"),
	})
	fmt.Println("Success connect to db")

	router := mux.NewRouter()

	// Routes and Functions
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
	router.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("file")
		if err != nil {
			Error(w, http.StatusInternalServerError, err)
			return
		}
		if handler.Header["Content-Type"][0] != "text/csv" {
			Error(w, http.StatusNotAcceptable, err)
			return
		}
		defer file.Close()

		f, err := os.OpenFile("./uploads/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("Error Upload")
			Error(w, http.StatusInternalServerError, err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		if err != nil {
			fmt.Println("Error Download")
			Error(w, http.StatusInternalServerError, err)
			return
		}
		reader := csv.NewReader(f)
		reader.Comma = ';'
		records, err := reader.ReadAll()
		if err != nil {
			Error(w, http.StatusInternalServerError, err)
			return
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
		tx, err := db.Begin()
		if err != nil {
			Error(w, http.StatusInternalServerError, err)
			return
		}
		defer tx.Rollback()

		tx.Exec("TRUNCATE TABLE drill_floats CASCADE")
		for _, drillFloat := range drillFloats {
			_, err = tx.Model(&drillFloat).Insert()
			if err != nil {
				Error(w, http.StatusInternalServerError, err)
				return
			}
		}
		err = tx.Commit()
		if err != nil {
			Error(w, http.StatusInternalServerError, err)
			return
		}
		Success(w, http.StatusOK, "Success Upload File")
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
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
