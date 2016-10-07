package main

import (
	"os"
  "strconv"
	"strings"
	"regexp"
	"fmt"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
)

type MyData struct {
	id int
	class int
	code int
	name string
	op1 int
	op2 int
	op3 string
}

func connectDB() error {
	var err error
	dbURL := os.Getenv("CLEARDB_DATABASE_URL")
	rep := regexp.MustCompile(`mysql://([^@]+)@([^/]+)/([^?]+)?\S+`)
	dbURL = rep.ReplaceAllString(dbURL, "$1@tcp($2:3306)/$3")
	db, err = sql.Open("mysql", dbURL)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

func extractKeycodes(text string) ([]int, []int, []int, error) {
	songs := []int{-1}
	artists := []int{-1}
	animes := []int{-1}

	q := "SELECT class, code, name FROM asb_data;"
	rows, err := db.Query(q)
	if err != nil {
		fmt.Println(err)
		return nil, nil, nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var class, code int
		var name string
		err := rows.Scan(&class, &code, &name)
		if err != nil {
			fmt.Println(err)
			continue
		}
		
		if strings.Index(text, name) != -1 {
			if class == 0 {
				animes = append(animes, code)
			} else if class == 1 {
				artists = append(artists, code)
			} else if class == 2 {
				songs = append(songs, code)
			}
		}
	}
	return songs, artists, animes, err
}

func searchQuery(text string) (string, error) {
	var ret string
	
	songs, artists, animes, err := extractKeycodes(text)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	songlist := make([]string, 0)
	lyrics := make([]string, 0)
	
	for _, song := range songs {
		for _, artist := range artists {
			for _, anime := range animes {
				if len(songs) > 1 && song == -1 || len(artists) > 1 && artist == -1 || len(animes) > 1 && anime == -1 {
					continue
				}
				q := "SELECT code FROM asb_data WHERE "
				if song != -1 {
					q += "id = "+strconv.Itoa(song)+" AND "
				}
				if artist != -1 {
					q += "option2 = "+strconv.Itoa(artist)+" AND "					
				}
				if anime != -1 {
					q += "option1 = "+strconv.Itoa(anime)+" AND "					
				}
				if song != -1 || artist != -1 || anime != -1 {
					q += strconv.Itoa(1)
				} else {
					q += strconv.Itoa(0)
				}
				fmt.Println(q)
				rows, err := db.Query(q)
				if err != nil {
					fmt.Println(err)
					continue
				}
				defer rows.Close()

				for rows.Next() {
					var code int
					err := rows.Scan(&code)
					if err != nil {
						fmt.Println(err)
						continue
					}
					data, err := getOriginalData(code)
					if err != nil {
						fmt.Println(err)
						continue
					}
					an_data, err := getOriginalData(data.op1)
					if err != nil {
						fmt.Println(err)
						continue
					}
					ar_data, err := getOriginalData(data.op2)
					if err != nil {
						fmt.Println(err)
						continue
					}
					songlist = append(songlist, data.name+" "+an_data.name+" "+ar_data.name)
					lyrics = append(lyrics, data.op3)
				}
			}
		}
	}

	if len(songlist) == 0 {
		ret = strconv.Quote(text)+"に関連するアニソンは見つかりませんでした。"
	} else if len(songlist) == 1 && len(songs) > 1 {
		ret = songlist[0]+"\n"+lyrics[0]
	} else {
		ret = strconv.Itoa(len(songlist))+"件のアニソンが見つかりました。"
		for _, s := range songlist {
			ret += "\n"+s
		}
	}
	
	return ret, err
}

func getOriginalData(code int) (*MyData, error) {
	var ret MyData
	q := "SELECT * FROM asb_data WHERE id = " + strconv.Itoa(code) + ";"
	rows, err := db.Query(q)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, class, op1, op2 int
		var name, op3 string
		err := rows.Scan(&id, &class, &code, &name, &op1, &op2, &op3)
		if err != nil {
			fmt.Println(err)
			continue
		}
		ret = MyData{id, class, code, name, op1, op2, op3}
	}
	return &ret, err
}

