package dotenv

import (
	"log"
	"github.com/joho/godotenv"
)

func init() {
	// TODO: This should not be here, it should be loaded in a more intentional location
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Failed to read .env file")
	}

}
