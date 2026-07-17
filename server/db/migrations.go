package db

import "fmt"

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS User (
		Id SERIAL PRIMARY KEY,
		Email VARCHAR(255) UNIQUE NOT NULL,
		CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS AccessToken (
		Id SERIAL PRIMARY KEY,
		Token VARCHAR(255) UNIQUE NOT NULL,
		UserId INT NOT NULL,
		FOREIGN KEY (UserId) REFERENCES User(Id),
		CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,
}

func RunMigrations() error {
	DB.Exec("CREATE TABLE IF NOT EXISTS Migration (Id SERIAL PRIMARY KEY, Position INT NOT NULL, INDEX idx_Position (Position))")
	var rowCount int
	err := DB.QueryRow("SELECT COUNT(*) FROM Migration").Scan(&rowCount)
	if err != nil {
		fmt.Println("Error running migration", err)
		return err
	}
	for index, migration := range migrations {
		if index >= rowCount {
			_, err := DB.Exec(migration)
			if err != nil {
				fmt.Println("Error running migration", err)
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
