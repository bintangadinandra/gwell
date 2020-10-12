package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
)

// Drill ...
type Drill struct {
	Depth     string `json:"depth"`
	ParticleA string `json:"particle_a" pg:"particle_a"`
	ParticleB string `json:"particle_b" pg:"particle_b"`
	ParticleC string `json:"particle_c" pg:"particle_c"`
}

func main() {
	start := time.Now()
	csvFile, err := os.Open("dummy3000.csv")
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

	var drills []Drill
	for _, rec := range records {
		drill := Drill{
			Depth:     rec[0],
			ParticleA: rec[1],
			ParticleB: rec[2],
			ParticleC: rec[3],
		}
		drills = append(drills, drill)
	}

	// Init Database
	db := pg.Connect(&pg.Options{
		Addr:     os.Getenv("PG_ADDRESS"),
		User:     os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		Database: os.Getenv("PG_DATABASE"),
	})
	fmt.Println("Success connect to db")

	_, err = db.Exec("TRUNCATE TABLE drills CASCADE")
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer tx.Rollback()

	for _, drill := range drills {
		_, err = tx.Model(&drill).Insert()
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
}
