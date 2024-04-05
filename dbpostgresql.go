package dbpostgresql

import (
	"database/sql"
	"errors"
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
	Sslmode  = ""
)

type Userdata struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

func openConnection() (*sql.DB, error) {
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", Hostname, Port, Username, Password, Database, Sslmode)

	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("DB connect")
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
	statement := fmt.Sprintf(`SELECT "id" FROM "users" where username='%s'`, username)

	rows, err := db.Query(statement)
	if err != nil {
		return -1
	}

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

	insertStatement = `insert into "userdata" ("userid", "name", "surname", "description") values ($1, $2, $3, $4)`

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

	statement := fmt.Sprintf(`SELECT "username" FROM "users" where id=%d`, id)

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
	deleteStatement = `delete from "users" where id=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	return nil

}

func ListUsers() ([]Userdata, error) {
	Data := []Userdata{}

	db, err := openConnection()
	if err != nil {
		return Data, err
	}

	defer db.Close()

	statement := `SELECT "id", "username", "name","surname", "description" FROM "users", "userdata" WHERE users.id=userdata.userid `
	rows, err := db.Query(statement)
	if err != nil {
		return Data, err
	}

	for rows.Next() {
		var id int
		var username string
		var name string
		var surname string
		var description string

		err = rows.Scan(&id, &username, &name, &surname, &description)
		temp := Userdata{ID: id, Name: name, Username: username, Surname: surname, Description: description}
		Data = append(Data, temp)
		if err != nil {
			return Data, err
		}
	}

	defer rows.Close()
	return Data, nil

}

func UpdateUser(d Userdata) error {
	db, err := openConnection()
	if err != nil {
		return err
	}

	defer db.Close()

	userID := exists(d.Username)
	if userID == -1 {
		return errors.New("user does not exist")
	}

	d.ID = userID
	updateStatement := `update "userdata" set "name"=$1, "surname"=$2, "description"=$3 where "userid"=$4`

	_, err = db.Exec(updateStatement, &d.Name, &d.Surname, &d.Description, &d.ID)
	if err != nil {
		return err
	}
	return nil

}
