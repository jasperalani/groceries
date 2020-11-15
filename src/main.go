package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
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
	log              *os.File
	db               *sqlx.DB
	router           *gin.Engine
	identifierLength int
}

func main() {

	app.identifierLength = 8

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

	app.router.GET("/list/:identifier", func(c *gin.Context) {

		identifier := c.Param("identifier")

		regex := `^[a-zA-Z0-9]{` + fmt.Sprintf("%d", app.identifierLength) + `}$`
		valid, err := regexp.MatchString(regex, identifier)
		if err != nil {
			log.Fatal("Fatal error when matching regexp.")
		}

		if !valid {
			c.HTML(http.StatusOK, "blank.html", gin.H{
				"text": "Invalid list identifier",
			})
		}

		getTitleQuery := sq.Select("id", "title").From("lists").Where(sq.Eq{"identifier": identifier})
		rows, err := getTitleQuery.RunWith(app.db).Query()
		if err != nil {
			log.Fatal(err)
		}

		var (
			id    int
			title string
		)
		rows.Next()
		rows.Scan(&id, &title)

		var items []string

		getItemsQuery := sq.Select("text").From("items").Where(sq.Eq{"list_id": id})
		rows, err = getItemsQuery.RunWith(app.db).Query()
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var tempItem string
			rows.Scan(&tempItem)
			items = append(items, tempItem)
		}

		c.HTML(http.StatusOK, "list.html", gin.H{
			"list_id":    id,
			"identifier": identifier,
			"title":      title,
			"list":       items,
		})

	})

	app.router.POST("/submitItem", func(c *gin.Context) {
		var listID = c.PostForm("list-id")
		var text = c.PostForm("item-name")
		if "" == text {
			return
		}

		insertNewItemQuery := sq.Insert("items").Columns("list_id", "text").Values(listID, text)
		_, err := insertNewItemQuery.RunWith(app.db).Query()
		if err != nil {
			log.Fatal(err)
		}

		c.Redirect(301, "/list/"+c.PostForm("identifier"))
	})

	app.router.POST("/createList", func(c *gin.Context) {
		var title = c.PostForm("item-name")
		if "" == title {
			return
		}

		identifier := GenerateIdentifier(app.identifierLength)

		createNewListQuery := sq.Insert("lists").Columns("identifier", "title").Values(identifier, title)
		_, err := createNewListQuery.RunWith(app.db).Query()
		if err != nil {
			// Log error, not fatal
			app.writeToLog("Error: failed to create new list with name: " + title)
		}

		c.Redirect(301, "/list/"+identifier)
	})

	app.router.POST("/goToList", func(c *gin.Context) {
		c.Redirect(301, "/list/"+c.PostForm("identifier"))
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

func formatLog(message string) string {
	return `[Groceries] ` + time.Now().Format("2006/01/02 - 15:04:05") + " " + message + "\n"
}

func (app application) writeToLog(message string) bool {
	_, err := gin.DefaultWriter.Write([]byte(formatLog(message)))
	if err != nil {
		log.Fatal("Failed to write to log: fatal.")
	}
	return true
}
