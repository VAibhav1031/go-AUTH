package cli

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

type Session struct {
	UserID     int64
	Username   string
	LoggedInAt time.Time
	ExpiresAt  time.Time
	MFAenabled bool
	LastLogin  *time.Time
}

func init(){
	db, err := sql.Open("sqlite", "app.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 2. Create a table
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT UNIQUE NOT NULL,
		lastlogin DATETIME,
		blockedtim DATETIME
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}


}

func Initiate() {

	timeout := time.Now().Add(5 * time.Minute)
	var currentSession *Session
	var handlers func()
	var ok bool

	// Readline instance creation for the current tErminal
	rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

	
	// App Object Construction
	app := &App{db: *Conn,Session: *currentSession,rl: rl}

	// make sure app struct populated with necessary things
	preLoginCmds := map[string]func(){"register": app.handleRegister, "login": app.handleLogin, "help": app.handleHelp, "exit": func() { os.Exit(0) }}
	postLoginCmds := map[string]func(){"whoami": app.handleWhoami, "enable-2fa": app.handleEnable2Fa, "disable-2fa": app.handleDisable2Fa, "logout": app.handleLogout, "help": app.handlePostHelp}

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		} // log
		cmd := strings.Fields(strings.TrimSpace(line))
		if len(cmd) > 1 {
			break
			//log
		}

		/*  Session Management */
		if currentSession == nil {
			//  do we have to say something or we have to show the comamnds or what
			// or lastLogin thing
			handlers, ok = preLoginCmds[cmd[0]]
		} else {
			currentSession.ExpiresAt = timeout.Add(currentSession.LoggedInAt.Sub(time.Time{}))

			if time.Now().After(currentSession.ExpiresAt) {
				currentSession = nil
				continue
			}

			handlers, ok = postLoginCmds[cmd[0]]

		}

		if !ok {
			fmt.Printf("Unknown command!! '%v'", cmd)
			continue
		}

		err = 
		if err != nil {
		}

	}

}
