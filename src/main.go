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

type item struct {
	ID   int
	Text string
}

type list struct {
	Identifier string
	Title      string
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

	app.router = gin.Default()

	app.router.LoadHTMLGlob("templates/*")

	app.router.StaticFS("/static", http.Dir("./static"))

	app.router.GET("/", func(c *gin.Context) {
		var (
			object    gin.H
			listEmpty = false
		)

		getListsQuery := sq.Select("identifier", "title").From("lists")
		rows, err := getListsQuery.RunWith(app.db).Query()
		if err != nil {
			app.writeToLog("Failed to retrieve lists")
		}

		var lists []list

		for rows.Next() {
			var list list
			rows.Scan(&list.Identifier, &list.Title)
			lists = append(lists, list)
		}

		if len(lists) == 0 {
			listEmpty = true
		}

		log.Println(listEmpty)

		object = gin.H{
			"list_not_empty": !listEmpty,
			"lists":          lists,
		}

		c.HTML(http.StatusOK, "index.html", object)
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

		var items []item

		getItemsQuery := sq.Select("id", "text").From("items").Where(sq.Eq{"list_id": id})
		rows, err = getItemsQuery.RunWith(app.db).Query()
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var item item
			rows.Scan(&item.ID, &item.Text)
			items = append(items, item)
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

		c.Redirect(301, "/list/"+c.PostForm("list-identifier"))
	})

	app.router.POST("/confirmRemoveItem", func(c *gin.Context) {
		var (
			id     = c.PostForm("item-id")
			listID = c.PostForm("list-identifier")
		)

		c.HTML(http.StatusOK, "remove.html", gin.H{
			"id":     id,
			"listID": listID,
		})
	})

	app.router.POST("/removeItem", func(c *gin.Context) {
		var id = c.PostForm("item-id")
		if "" == id {
			return
		}

		deleteQuery := sq.Delete("items").Where(sq.Eq{"id": id})
		_, err := deleteQuery.RunWith(app.db).Query()
		if err != nil {
			// Log error, not fatal
			app.writeToLog("Error: failed to delete item with id: " + id)
		}

		c.Redirect(301, "/list/"+c.PostForm("list-identifier"))
	})

	app.router.POST("/editingItem", func(c *gin.Context) {
		c.HTML(http.StatusOK, "edit.html", gin.H{
			"previousValue": c.PostForm("item-value"),
			"listID":        c.PostForm("list-identifier"),
			"id":            c.PostForm("item-id"),
		})
	})

	app.router.POST("/editItem", func(c *gin.Context) {
		/*
			item-id
			item-name
		*/
		var id = c.PostForm("item-id")
		updateQuery := sq.Update("items").Set("text", c.PostForm("item-name")).Where(sq.Eq{"id": id})
		_, err := updateQuery.RunWith(app.db).Query()
		if err != nil {
			// Log error, not fatal
			app.writeToLog("Error: failed to edit item with id: " + id)
		}

		c.Redirect(301, "/list/"+c.PostForm("list-identifier"))

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
