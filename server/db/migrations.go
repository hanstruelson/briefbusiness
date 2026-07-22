package db

import "fmt"

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS User (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Email VARCHAR(255) UNIQUE NOT NULL,
		CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS AccessToken (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Token VARCHAR(255) UNIQUE NOT NULL,
		UserId INT NOT NULL,
		FOREIGN KEY (UserId) REFERENCES User(Id),
		CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS Draft (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Title VARCHAR(255) NOT NULL,
		Markdown TEXT NOT NULL,
		UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UserId INT NOT NULL,
		FOREIGN KEY (UserId) REFERENCES User(Id)
	)`,
	`CREATE TABLE IF NOT EXISTS DraftBackup (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		DraftId INT NOT NULL,
		Title VARCHAR(255) NOT NULL,
		Markdown TEXT NOT NULL,
		SavedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (DraftId) REFERENCES Draft(Id)
	)`,
}

func RunMigrations() error {
	_, err := DB.Exec("CREATE TABLE IF NOT EXISTS Migration (Id INTEGER PRIMARY KEY AUTOINCREMENT, Position INT NOT NULL)")
	if err != nil {
		fmt.Println("Error creating Migration table:", err)
		return err
	}
	_, err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_Position ON Migration (Position)")
	if err != nil {
		fmt.Println("Error creating Migration index:", err)
		return err
	}
	var rowCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM Migration").Scan(&rowCount)
	if err != nil {
		fmt.Println("Error running migration count query", err)
		return err
	}
	for index, migration := range migrations {
		if index >= rowCount {
			_, err := DB.Exec(migration)
			if err != nil {
				fmt.Println("Error running migration index", index, err)
				return err
			}
			_, err = DB.Exec("INSERT INTO Migration (Position) VALUES (?)", index)
			if err != nil {
				fmt.Println("Error inserting migration record", err)
				return err
			}
		}
	}
	return nil
}
