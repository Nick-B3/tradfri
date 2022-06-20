package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Data die verzonden gaat worden en dus hier binnenkomt
type Reading struct {
	TimeStamp string
	Lamp      string
	Status    string
}

var db *sql.DB
var dbError error

func init() {
	// database opstellen
	fmt.Println("Verbonden! ")
	db, dbError = sql.Open("sqlite3", "./readings.db")
	Check(dbError)
	statement, prepError := db.Prepare("CREATE TABLE IF NOT EXISTS reading (TimeStamp TEXT, Lamp TEXT, Status TEXT)")
	Check(prepError)
	statement.Exec()

}

func main() {

	r := gin.Default()

	// Dit creëert een standaard Gin-router
	r.GET("/reading", func(c *gin.Context) {
		lastTen := getLastTen()
		// stopt het in een json object en returnt het
		c.JSON(200, gin.H{"message": lastTen})

	})
	// Dit legt alle POST-commando's vast in /reading
	r.POST("/reading", verwerkData)

	//start de API
	r.Run(":5000")

}

func verwerkData(c *gin.Context) {
	// haal uit de originele post en dan gaat het naar de struct

	if c.Request.Method == "POST" {

		fmt.Println("Bezig...")
		fmt.Println("")

		var r Reading
		c.BindJSON(&r)

		// opslaan in de database
		saveToDatabase(r.TimeStamp, r.Lamp, r.Status)

		c.JSON(http.StatusOK, gin.H{
			"status":  "Posted!",
			"Message": "This worked!",
		})

		TienWaardesZien(db)
	}
}

// opslaan in de database
func saveToDatabase(TimeStamp string, Lamp string, Status string) {

	statement, err := db.Prepare("INSERT INTO reading (TimeStamp, Lamp, Status) VALUES (?,?,?)")
	Check(err)

	_, err = statement.Exec(TimeStamp, Lamp, Status)
	Check(err)

}

func getLastTen() []Reading {

	// de database opvragen voor metingen
	rows, _ := db.Query("SELECT TimeStamp, Lamp, Status from reading LIMIT 200")

	// enkele tijdelijke variabelen maken
	var TimeStamp string
	var Lamp string
	var Status string

	// maak een "slice"
	lastTen := make([]Reading, 10)

	// hier gaat de data in de "slice"
	for rows.Next() {
		rows.Scan(&TimeStamp, &Lamp, &Status)
		lastTen = append(lastTen, Reading{TimeStamp: TimeStamp, Lamp: Lamp, Status: Status})

	}
	// return
	return lastTen
}

// checkt of er errors zijn en laat dat weten
func Check(e error) {

	if e != nil {
		panic(e)
	}
}

func TienWaardesZien(db *sql.DB) {

	// pak de laatste 10 waarden van alles, de nieuwste bovenaan
	row, err := db.Query("SELECT * FROM reading ORDER BY rowid DESC LIMIT 10")
	if err != nil {
		Check(err)
	}
	defer row.Close()
	for row.Next() { // de waarden ophalen
		var TimeStamp string
		var Lamp string
		var Status string

		row.Scan(&TimeStamp, &Lamp, &Status)

		//print de laatste 10 waarden van alles uit
		fmt.Println("Moment: ", TimeStamp, " ", "Lamp: ", Lamp, " ", "Status: ", Status)
		fmt.Println("")

	}

}
