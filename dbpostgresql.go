package dbpostgresql

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

var (
	Hostname = ""
	Port     = 2345
	Username = ""
	Password = ""
	Database = ""
)

type Userdata struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

func openConnection() (*sql.DB, error) {
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", Hostname, Port, Username, Password, Database)

	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func exists(username string) int {
	username = strings.ToLower(username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}

	defer db.Close()

	userID := -1
	statement := fmt.Sprintf(`SELECT "id" FROM "users" where username = '%s'`, username)

	rows, err := db.Query(statement)

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			fmt.Println("Scan(): ", err)
			return -1
		}
		userID = id
	}

	defer rows.Close()
	return userID
}

func AddUser(d Userdata) int {
	d.Username = strings.ToLower(d.Username)
	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userId := exists(d.Username)
	if userId != -1 {
		fmt.Println("User already exists:", d.Username)
		return -1
	}

	insertStatement := `insert into "users" ("username") values ($1)`
	_, err = db.Exec(insertStatement, d.Username)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	userId = exists(d.Username)
	if userId == -1 {
		return userId
	}

	insertStatement = `insert into "userdata" ("userId", "name", "surname", "description") values ($1, $2, $3, $4)`

	_, err = db.Exec(insertStatement, userId, d.Name, d.Surname, d.Description)
	if err != nil {
		fmt.Println("db.Exec(): ", err)
		return -1
	}
	return userId

}

func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}

	defer db.Close()

	statement := fmt.Sprintf(`SELECT "username" FROM "users" where id = %d`, id)

	rows, err := db.Query(statement)
	if err != nil {
		return err
	}

	var username string
	for rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			return err
		}
	}

	defer rows.Close()

	if exists(username) != id {
		return fmt.Errorf("user with ID %d does not exist", id)
	}

	deleteStatement := `delete from "userdata" where userid=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}
	deleteStatement = `delete from "users" where id = $1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	return nil

}



