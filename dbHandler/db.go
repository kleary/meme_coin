package dbHandler

import (
	_ "database/sql"
	"fmt"
	"log"
	_ "strings"
	"time"

	_ "github.com/bmizerany/pq"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

// User is a struct that maps 1 to 1 with 'money' db table
type User struct {
	ID        int       `db:"id"`
	DID       string    `db:"discord_id"`
	Username  string    `db:"name"`
	CurMoney  int       `db:"current_money"`
	TotMoney  int       `db:"total_money"`
	WonMoney  int       `db:"won_money"`
	LostMoney int       `db:"lost_money"`
	GiveMoney int       `db:"given_money"`
	RecMoney  int       `db:"received_money"`
	EarMoney  int       `db:"earned_money"`
	MineTime  time.Time `db:"mine_time"`
}

func DbGet() *sqlx.DB {
	db, err := sqlx.Connect("postgres", "host=localhost user=memebot dbname=money password=password sslmode=disable parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func createUser(user *discordgo.User, db *sqlx.DB) {
	var newUser User
	newUser.DID = user.ID
	newUser.Username = user.Username
	_, err := db.NamedExec(`INSERT INTO money (discord_id, name) VALUES (:discord_id, :name)`, newUser)
	if err != nil {
		log.Fatal(err)
	}
}

func UserGet(discordUser *discordgo.User, db *sqlx.DB) User {
	var users []User
	//fmt.Println(discordUser.ID)
	err := db.Select(&users, `SELECT id, discord_id, name, current_money, total_money, won_money, lost_money, given_money, received_money, earned_money, mine_time FROM money WHERE discord_id = $1`, discordUser.ID)
	if err != nil {
		log.Fatal(err)
	}
	var user User
	if len(users) == 0 {
		fmt.Println("creating user: " + discordUser.ID)
		createUser(discordUser, db)
		user = UserGet(discordUser, db)
	} else {
		user = users[0]
	}
	return user
}

func MoneyDeduct(user *User, amount int, deduction string, db *sqlx.DB) {
	newCurrentMoney := user.CurMoney - amount
	newDeductionAmount := -1
	dbString := ``
	deductionRecord := -1

	if deduction == "tip" {
		dbString = `UPDATE money SET (current_money, given_money) = ($1, $2) WHERE discord_id = `
		deductionRecord = user.GiveMoney
		newDeductionAmount = user.GiveMoney + amount
		user.CurMoney = newCurrentMoney
		user.GiveMoney = newDeductionAmount
	}
	if deduction == "gamble" {
		dbString = `UPDATE money SET (current_money, lost_money) = ($1, $2) WHERE discord_id = `
		deductionRecord = user.LostMoney
		newDeductionAmount = user.LostMoney + amount
		user.CurMoney = newCurrentMoney
		user.LostMoney = newDeductionAmount
	}

	if dbString != `` && deductionRecord != -1 && newDeductionAmount != -1 {
		dbString = dbString + `'` + user.DID + `'`
		db.MustExec(dbString, newCurrentMoney, newDeductionAmount)
	}
}

func MoneyAdd(user *User, amount int, addition string, db *sqlx.DB) {
	newCurrentMoney := user.CurMoney + amount
	newAdditionAmount := -1
	dbString := ``
	additionRecord := -1

	if addition == "tip" {
		dbString = `UPDATE money SET (current_money, received_money) = ($1, $2) WHERE discord_id = `
		additionRecord = user.RecMoney
		newAdditionAmount = user.RecMoney + amount
		user.CurMoney = newCurrentMoney
		user.RecMoney = newAdditionAmount
	}
	if addition == "gamble" {
		dbString = `UPDATE money SET (current_money, won_money) = ($1, $2) WHERE discord_id = `
		additionRecord = user.WonMoney
		newAdditionAmount = user.WonMoney + amount
		user.CurMoney = newCurrentMoney
		user.WonMoney = newAdditionAmount
	}
	if addition == "mined" {
		dbString = `UPDATE money SET (current_money, earned_money, mine_time) = ($1, $2, $3) WHERE discord_id = `
		additionRecord = user.EarMoney
		newAdditionAmount = user.EarMoney + amount
		user.CurMoney = newCurrentMoney
		user.EarMoney = newAdditionAmount
		// bindvars can only be used as values so we have to concat the user.DID onto the db string
		dbString = dbString + `'` + user.DID + `'`
		db.MustExec(dbString, newCurrentMoney, newAdditionAmount, time.Now())
	} else {
		if dbString != `` && additionRecord != -1 && newAdditionAmount != -1 {
			// bindvars can only be used as values so we have to concat the user.DID onto the db string
			dbString = dbString + `'` + user.DID + `'`
			db.MustExec(dbString, newCurrentMoney, newAdditionAmount)
		}
	}
}