package store

import (
	"database/sql"
	"fmt"
	"errors" 
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

type DBConfig struct {
	Host     string 
	User     string
	Password string
	Database string
}

type MySqlStore struct {
	db *sql.DB
}

func ConnectToStore(config DBConfig) (MySqlStore, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local",
		config.User, config.Password, config.Host, config.Database)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		// Use fmt.Errorf to wrap the original error for better debugging
		return MySqlStore{}, fmt.Errorf("failed to open database connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return MySqlStore{}, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	store := MySqlStore{
		db: db,
	}
	return store, nil
}

func (store *MySqlStore) Close() error {
	if store.db != nil {
		return store.db.Close()
	}
	return nil
}

func (store *MySqlStore) CreateBucket(bucketName string) error {
	if bucketName == "" {
		return errors.New("bucketName cannot be empty")
	}

	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			file_name VARCHAR(255) NOT NULL,
			file_data LONGBLOB
		)
	`, bucketName) 

	_, err := store.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create bucket '%s': %w", bucketName, err)
	}
	fmt.Printf("Ensured bucket/table '%s' exists.\n", bucketName)
	return nil
}

func (store *MySqlStore) DeleteBucket(bucketName string) error {
	if bucketName == "" {
		return errors.New("bucketName cannot be empty")
	}
	if store.db == nil {
		return errors.New("database connection is not initialized")
	}

	// First check if the bucket exists
	checkSQL := "SELECT COUNT(*) FROM information_schema.tables WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?"
	var count int
	err := store.db.QueryRow(checkSQL, bucketName).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if bucket '%s' exists: %w", bucketName, err)
	}

	if count == 0 {
		return fmt.Errorf("bucket '%s' not found", bucketName)
	}

	// Drop the table (bucket)
	dropTableSQL := fmt.Sprintf("DROP TABLE %s", bucketName)
	_, err = store.db.Exec(dropTableSQL)
	if err != nil {
		return fmt.Errorf("failed to delete bucket '%s': %w", bucketName, err)
	}

	fmt.Printf("Successfully deleted bucket '%s'.\n", bucketName)
	return nil
}

func (store *MySqlStore) InsertFile(bucketName string, fileName string, fileData []byte) error {
	if bucketName == "" {
		return errors.New("bucketName cannot be empty")
	}
	if fileName == "" {
		return errors.New("fileName cannot be empty")
	}
	if store.db == nil {
		return errors.New("database connection is not initialized")
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (file_name, file_data) VALUES (?, ?)", bucketName)
	stmt, err := store.db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement for bucket '%s': %w", bucketName, err)
	}
	defer stmt.Close() // Ensure the prepared statement is closed

	result, err := stmt.Exec(fileName, fileData)
	if err != nil {
		return fmt.Errorf("failed to insert file '%s' into bucket '%s': %w", fileName, bucketName, err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("Warning: Could not get last insert ID for file '%s' in bucket '%s': %v\n", fileName, bucketName, err)
	} else {
		fmt.Printf("File '%s' inserted successfully into bucket '%s' with ID: %d\n", fileName, bucketName, lastInsertID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("Warning: Could not get rows affected for file '%s' in bucket '%s': %v\n", fileName, bucketName, err)
	} else {
		fmt.Printf("Rows affected for '%s': %d\n", fileName, rowsAffected)
	}

	return nil // No error
}

func (store *MySqlStore) DeleteFile(bucketName string, fileName string) error {
	if bucketName == "" {
		return errors.New("bucketName cannot be empty")
	}
	if fileName == "" {
		return errors.New("fileName cannot be empty")
	}
	if store.db == nil {
		return errors.New("database connection is not initialized")
	}

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE file_name = ?", bucketName)
	stmt, err := store.db.Prepare(deleteSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement for bucket '%s': %w", bucketName, err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(fileName)
	if err != nil {
		return fmt.Errorf("failed to delete file '%s' from bucket '%s': %w", fileName, bucketName, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for file '%s' in bucket '%s': %w", fileName, bucketName, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file '%s' not found in bucket '%s'", fileName, bucketName)
	}

	fmt.Printf("Successfully deleted file '%s' from bucket '%s'. Rows affected: %d\n", fileName, bucketName, rowsAffected)
	return nil
}

func (store *MySqlStore) GetFile(bucketName string, fileName string) ([]byte, error) {
	if bucketName == "" {
		return nil, errors.New("bucketName cannot be empty")
	}
	if fileName == "" {
		return nil, errors.New("fileName cannot be empty")
	}
	if store.db == nil {
		return nil, errors.New("database connection is not initialized")
	}

	var fileData []byte
	querySQL := fmt.Sprintf("SELECT file_data FROM %s WHERE file_name = ?", bucketName)

	err := store.db.QueryRow(querySQL, fileName).Scan(&fileData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file '%s' not found in bucket '%s'", fileName, bucketName)
		}
		return nil, fmt.Errorf("failed to retrieve file '%s' from bucket '%s': %w", fileName, bucketName, err)
	}

	fmt.Printf("Successfully retrieved file '%s' from bucket '%s'.\n", fileName, bucketName)
	return fileData, nil
}

func (store *MySqlStore) GetFiles(bucketName string) ([]string, error) {
	if bucketName == "" {
		return nil, errors.New("bucketName cannot be empty")
	}
	if store.db == nil {
		return nil, errors.New("database connection is not initialized")
	}

	var fileNames []string
	querySQL := fmt.Sprintf("SELECT file_name FROM %s", bucketName)

	rows, err := store.db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query files from bucket '%s': %w", bucketName, err)
	}
	defer rows.Close() // Ensure rows are closed

	for rows.Next() {
		var fileName string
		if err := rows.Scan(&fileName); err != nil {
			return nil, fmt.Errorf("failed to scan file name from bucket '%s': %w", bucketName, err)
		}
		fileNames = append(fileNames, fileName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for bucket '%s': %w", bucketName, err)
	}

	fmt.Printf("Successfully retrieved %d files from bucket '%s'.\n", len(fileNames), bucketName)
	return fileNames, nil
}

func (store *MySqlStore) GetBuckets() ([]string, error) {
	if store.db == nil {
		return nil, errors.New("database connection is not initialized")
	}

	var bucketNames []string
	querySQL := fmt.Sprintf("SELECT TABLE_NAME FROM information_schema.tables WHERE TABLE_SCHEMA = DATABASE()")

	rows, err := store.db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query bucket names: %w", err)
	}
	defer rows.Close() // Ensure rows are closed

	for rows.Next() {
		var bucketName string
		if err := rows.Scan(&bucketName); err != nil {
			return nil, fmt.Errorf("failed to scan bucket name: %w", err)
		}
		bucketNames = append(bucketNames, bucketName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for bucket names: %w", err)
	}

	fmt.Printf("Successfully retrieved %d buckets.\n", len(bucketNames))
	return bucketNames, nil
}