package config

import (
	"vilow-be/prisma/db"
)

// SetupDatabase is a function that sets up the database client
func SetupDatabase() (*db.PrismaClient, error) {
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		return nil, err
	}

	return client, nil
}
