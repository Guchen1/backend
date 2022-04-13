package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Passage struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Time    string `json:"time"`
}

var db *sqlx.DB

func main() {
	db, _ = sqlx.Connect("mysql", "my:123456@tcp(127.0.0.1:3307)/test?charset=utf8")
	db.SetConnMaxLifetime(100)
	db.SetMaxIdleConns(10)
	r := gin.Default()
	r.Use(cors.Default())
	var part []int
	rows, _ := db.Query("SELECT id FROM passages")
	for rows.Next() {
		var p int
		rows.Scan(&p)
		part = append(part, p)
	}
	//rows, _ = db.Query("SELECT id ,title,summary,time FROM passages")
	var resultsp sync.Map
	//resultsp = make(map[string][]int)
	r.GET("/", func(c *gin.Context) {
		key := c.Query("key")
		if key == "" {
			rows.Close()
			c.JSON(http.StatusOK, part)
		} else {
			if v, ok := resultsp.Load(key); ok {
				c.JSON(http.StatusOK, v)
			} else {
				var resultp []int
				row, _ := db.Query("SELECT id FROM passages WHERE title LIKE ? OR summary LIKE ? OR content LIKE ?", "%"+key+"%", "%"+key+"%", "%"+key+"%")
				for row.Next() {
					var p int
					row.Scan(&p)
					resultp = append(resultp, p)
				}
				resultsp.Store(key, resultp)
				c.JSON(http.StatusOK, resultp)
			}
		}
	})
	r.POST("/", func(c *gin.Context) {
		ids, _ := ioutil.ReadAll(c.Request.Body)
		var temp map[string][]int
		var idss []int
		json.Unmarshal(ids, &temp)
		idss = temp["id"]
		var full []Passage
		query, pra, _ := sqlx.In("SELECT id,title,summary,time FROM passages WHERE id IN (?)", idss)
		query = db.Rebind(query)
		rows, _ := db.Query(query, pra...)
		for rows.Next() {
			var fulls Passage
			rows.Scan(&fulls.Id, &fulls.Title, &fulls.Summary, &fulls.Time)
			full = append(full, fulls)
		}
		c.JSON(http.StatusOK, full)
	})
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
