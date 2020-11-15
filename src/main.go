package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	app application
)

type application struct {
	log    *os.File
	db     *sqlx.DB
	router *gin.Engine
}

func main() {

	gin.ForceConsoleColor()

	var logToFile = false
	if logToFile {
		gin.DisableConsoleColor()
	}

	app.log, _ = os.Create("groceries.log")
	gin.DefaultWriter = io.MultiWriter(app.log, os.Stdout)

	app.db = initDB()

	gin.SetMode(gin.DebugMode)

	var list = []string{
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam in turpis porta, feugiat lacus nec, ornare lacus. Fusce faucibus accumsan est, quis tempor urna accumsan quis. Nulla in pretium tellus, sed porttitor turpis. Vestibulum accumsan porta tristique. Nunc purus neque, sagittis eget eros sed, cursus vehicula orci. ",
		"poodle",
		"doodle",
	}

	app.router = gin.Default()

	app.router.LoadHTMLGlob("templates/*")

	app.router.StaticFS("/static", http.Dir("./static"))

	app.router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Title of list",
			"list":  list,
		})
	})

	app.router.GET("/createList", func(c *gin.Context) {
		c.HTML(http.StatusOK, "createList.html", gin.H{})
	})

	app.router.POST("/createList", func(c *gin.Context) {
		var title = c.PostForm("item-name")
		if "" == title {
			return
		}

		identifier := GenerateIdentifier(8)

		createNewListQuery := sq.Insert("lists").Columns("title").Values(title)
		_, err := createNewListQuery.RunWith(app.db).Query()
		if err != nil {
			// Log error, not fatal
			app.writeToLog("Error: failed to create new list with name: " + title)
		}

		c.Redirect(301, "/")
	})

	app.writeToLog("Application started successfully.")

	app.router.Run(":8080")
}

func initDB() *sqlx.DB {
	db, err := sqlx.Connect("mysql", "groceriesdbadmin:groceriesdbpass@tcp(127.0.0.1:3306)/groceries")
	if err != nil {
		log.Fatal(err)
	}
	app.writeToLog("Successfully connected to database.")
	return db
}

func (app application) writeToLog(message string) bool {
	var content = `[Groceries] ` + time.Now().Format("2006/01/02 - 15:04:05") + " " + message + "\n"
	_, err := gin.DefaultWriter.Write([]byte(content))
	if err != nil {
		log.Fatal("Failed to write to log: fatal.")
	}
	return true
}
