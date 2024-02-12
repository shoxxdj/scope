package main 

import (
	"fmt"
	"os"
	"errors"
	"flag"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main(){

	//Get config

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.scope/")
	err := viper.ReadInConfig()
	if err != nil {
		if _, err := os.Stat(os.Getenv("HOME")+"/.scope");
		errors.Is(err,os.ErrNotExist){
			fmt.Println("need to create directory")
			os.Mkdir(os.Getenv("HOME")+"/.scope",0755)
			os.Create(os.Getenv("HOME")+"/.scope/config")


			viper.SetDefault("database.dblocation", os.Getenv("HOME")+"/.scope/"+"scope.db")
//			viper.SetDefault("database.dbname","default")

			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
			viper.AddConfigPath("$HOME/.scope/")

			err :=viper.WriteConfig();
			if err!=nil{
				fmt.Println(err)
			}
		}else{
			fmt.Println("no idea")
		}
	}

	//Main

	//fmt.Println(viper.Get("database.dblocation"))
	//fmt.Println(chalk.Red, "Writing in colors", chalk.Cyan, "is so much fun", chalk.Reset)
	dbLocation:=	flag.String("dl",viper.GetString("database.dbLocation"),"Database location")
	saveConfig:=	flag.Bool("save",false,"Save configuration (need dbLocation to be defined to be efficient)")
	reset:=			flag.Bool("reset",false,"Reset configuration to default")

	fullDetails:=  flag.Bool("full",false,"Get full details (Location, url and id)")
	urlToAdd:= 	flag.String("a","","The url to add in the database")
	urlToDel:=		flag.Int("d",-1,"The url ID to delete from the database")
	category := flag.String("c","","The category to set the url in / search in")
	query := flag.String("q","","The url to search in scope")
	

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Scope : a binary to store scope for attackers. v:0.1\n")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "\t-%v: %v\n", f.Name,f.Usage) // f.Name, f.Value
		})
	}

	flag.Parse()


	if *saveConfig {
		viper.Set("database.dblocation",*dbLocation)
		err :=viper.WriteConfig();
		if err!=nil{
			fmt.Println(err)
		}
	}

	if *reset{
		viper.Set("database.dblocation", os.Getenv("HOME")+"/.scope/"+"scope.db")
		err :=viper.WriteConfig();
		if err!=nil{
			fmt.Println(err)
		}
		fmt.Println(chalk.Green,"[+]",chalk.Reset,"Configuration reset")
		os.Exit(1)
	}


	db, err := sql.Open("sqlite3", *dbLocation)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS scope (id INTEGER PRIMARY KEY AUTOINCREMENT, value text, category text)
	`
	_,err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	if *urlToDel != -1{
		tx,err := db.Begin()
		if err!=nil{
			log.Fatal(err)
		}
		stmt,err := tx.Prepare("delete from scope where id=?")
		if err!=nil{
			log.Fatal(err)
		}
		defer stmt.Close()
		stmt.Exec(*urlToDel)
		tx.Commit()
		fmt.Println("Url deleted")
		sqlStmt = "ALTER TABLE scope RENAME to old_scope";
		_,err = db.Exec(sqlStmt)
		if err!=nil{
			log.Fatal(err)
		}
		sqlStmt = "CREATE TABLE scope (id INTEGER PRIMARY KEY AUTOINCREMENT, value text, category text)"
		_,err = db.Exec(sqlStmt)
		if err!=nil{
			log.Fatal(err)
		}
		sqlStmt = "INSERT INTO scope(value,category) SELECT value from old_scope"
		_,err = db.Exec(sqlStmt)
		if err!=nil{
			log.Fatal(err)
		}
		sqlStmt = "DROP TABLE old_scope"
		_,err = db.Exec(sqlStmt)
		if err!=nil{
			log.Fatal(err)
		}
		fmt.Println(chalk.Red, "[-]", chalk.Reset, "Item deleted")
	}

	if *urlToAdd != ""{
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		stmt, err := tx.Prepare("insert into scope(value,category) values(?,?)")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		if(*category==""){
			stmt.Exec(*urlToAdd,"nc")
		}else{
			stmt.Exec(*urlToAdd,*category)
		}
		
		tx.Commit()
		fmt.Println(chalk.Green, "[+]", chalk.Reset, "Item added")
	}

	if(*query!=""){
		rows,err := db.Query("Select id,value,category from scope where value=?",*query)
		if err!=nil{
			log.Fatal(err)
		}
		//defer stmt.Close()
		//rows,err:=stmt.Exec(*category)
		if err!=nil{
			log.Fatal(err)
		}
		defer rows.Close()

		if(*fullDetails){
			fmt.Println("|",*dbLocation,"|")
		}
		for rows.Next(){
			var id int
			var value string
			var scope string 
			err = rows.Scan(&id,&value,&scope)
			if err!=nil{
				log.Fatal(err)
			}
			if(*fullDetails){
				fmt.Println(id,value,scope)
			}else{
				fmt.Println(value)
			}
		}
	}

	if *urlToAdd == "" && *urlToDel ==-1 && *query==""{

		if *category ==""{
			sqlStmt = "Select id,value,category from scope"
			rows,err := db.Query(sqlStmt)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			if(*fullDetails){
				fmt.Println("|",*dbLocation,"|")
			}

			for rows.Next(){
				var id int
				var value string
				var category string

				err = rows.Scan(&id,&value,&category)
				if err != nil{
					log.Fatal(err)
				}

				if(*fullDetails){
					fmt.Println(id,"|",value,"|",category)
				}else{
					fmt.Println(value)
				}
			}
		}else{
			rows,err := db.Query("Select id,value,category from scope where category=?",*category)
			if err!=nil{
				log.Fatal(err)
			}
			//defer stmt.Close()
			//rows,err:=stmt.Exec(*category)
			if err!=nil{
				log.Fatal(err)
			}
			defer rows.Close()

			if(*fullDetails){
				fmt.Println("|",*dbLocation,"|")
			}
			for rows.Next(){
				var id int
				var value string
				var scope string 
				err = rows.Scan(&id,&value,&scope)
				if err!=nil{
					log.Fatal(err)
				}
				if(*fullDetails){
					fmt.Println(id,value,scope)
				}else{
					fmt.Println(value)
				}
			}


		}

	}
}
