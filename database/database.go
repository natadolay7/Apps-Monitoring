package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB adalah variabel global untuk koneksi database
var DB *gorm.DB

// ConnectDatabase membuat koneksi ke PostgreSQL
func ConnectDatabase() error {
	// Konfigurasi dari environment variable dengan fallback
	config := getDBConfig()

	// Buat DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName,
	)

	// Konfigurasi logger GORM
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  true,        // Enable color
		},
	)

	// Koneksi ke database
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Printf("✅ Connected to PostgreSQL database: %s@%s:%s",
		config.User, config.Host, config.Port)

	return nil
}

// DBConfig menyimpan konfigurasi database
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// getDBConfig mengambil konfigurasi dari env atau default
func getDBConfig() DBConfig {
	return DBConfig{
		Host:     getEnvOrDefault("DB_HOST", "127.0.0.1"),
		Port:     getEnvOrDefault("DB_PORT", "5434"),
		User:     getEnvOrDefault("DB_USERNAME", "admin"),
		Password: getEnvOrDefault("DB_PASSWORD", "secret123"),
		DBName:   getEnvOrDefault("DB_DATABASE", "sigesitlocal_db"),
	}
}

// getEnvOrDefault helper function
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDB mengembalikan instance database
func GetDB() *gorm.DB {
	return DB
}
