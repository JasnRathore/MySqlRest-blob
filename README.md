# Blobstore - MySQL-Backed File Storage Service

Transform your MySQL database into a powerful file storage system! This Go-based REST API lets you create buckets, upload files, and retrieve them seamlessly - all while leveraging the reliability and familiarity of MySQL.

## Project Overview

Blobstore is a lightweight HTTP service that treats your MySQL database as a blob storage system. Instead of managing complex file systems or cloud storage configurations, you can store and retrieve files directly through a clean REST API. Each "bucket" becomes a MySQL table, and your files are stored as binary data with automatic content-type detection.

## Project Requirements

- **Go 1.22.1** or higher
- **MySQL Server** (any recent version)
- Network access to your MySQL instance

## Dependencies

The project uses minimal external dependencies to keep things lightweight:

```go
filippo.io/edwards25519 v1.1.0      // Cryptographic support
github.com/go-sql-driver/mysql v1.9.2  // MySQL driver
```

All dependencies are managed through Go modules, so installation is straightforward.

## Getting Started

### Database Setup

First, create your MySQL database and ensure your connection credentials are ready:

```sql
CREATE DATABASE mysqlrest;
```

Make sure your MySQL user has the necessary permissions to create tables and manage data.

### Configuration

The application connects to MySQL using these default settings:
- **Host**: localhost
- **User**: root  
- **Password**: root
- **Database**: mysqlrest

You can modify these values in the `main.go` file in the config section:

```go
config := store.DBConfig{
    Host:     "localhost", 
    User:     "your-username",
    Password: "your-password",
    Database: "your-database",
}
```

### Installation

Clone the repository and install dependencies:

```bash
git clone https://github.com/JasnRathore/MySqlRest-blob
cd MySqlRest-blob
go mod download
```

## How to Run the Application

Start the HTTP server with a simple command:

```bash
go run main.go
```

The server will start on port 8080 and you'll see output like:

```
MySQL Store connected successfully.
Starting HTTP server on port :8080...
```

Your API is now ready to handle requests at `http://localhost:8080`

## API Endpoints

### Bucket Management

**Create a new bucket:**
```bash
POST /buckets/create
Content-Type: application/json

{
    "bucketName": "my-images"
}
```

**Delete a bucket:**
```bash
DELETE buckets/{bucketName}
```

**List all buckets:**
```bash
GET /buckets
```

### File Operations

**Upload a file to a bucket:**
```bash
POST /buckets/{bucketName}/files
Content-Type: application/json

{
    "fileName": "photo.jpg",
    "fileData": "base64-encoded-file-content"
}
```

**Delete a file:**
```bash
DELETE buckets/{bucketName}/files/{fileName}
```

**List files in a bucket:**
```bash
GET /buckets/{bucketName}/files
```

**Retrieve a specific file:**
```bash
GET /buckets/{bucketName}/files/{fileName}
```

The file retrieval endpoint serves files directly with proper content-type headers, making it perfect for web applications.

## Practical Examples

### Setting Up Your First Bucket

Create a bucket for storing user profile images:

```bash
curl -X POST http://localhost:8080/buckets/create \
  -H "Content-Type: application/json" \
  -d '{"bucketName": "profile-pics"}'
```

### Uploading an Image

Convert your image to base64 and upload it:

```bash
# Convert image to base64 (on macOS/Linux)
base64 -i avatar.png -o avatar.b64

# Upload the file
curl -X POST http://localhost:8080/buckets/profile-pics/files \
  -H "Content-Type: application/json" \
  -d '{
    "fileName": "avatar.png",
    "fileData": "'$(cat avatar.b64)'"
  }'
```

### Accessing Your Files

Once uploaded, access your files directly through the browser:

```
http://localhost:8080/buckets/profile-pics/files/avatar.png
```

The service automatically detects content types, so images display properly in browsers, PDFs open correctly, and media files stream as expected.

### Integration Example

Here's how you might integrate this with a web application:

```html
<!-- Direct image embedding -->
<img src="http://localhost:8080/buckets/profile-pics/files/avatar.png" 
     alt="User Avatar">

<!-- PDF viewer -->
<iframe src="http://localhost:8080/buckets/documents/files/report.pdf">
</iframe>
```

## Content Type Support

The service automatically detects and serves files with appropriate content types:

- **Images**: JPEG, PNG, GIF, WebP, SVG, ICO
- **Audio**: MP3, WAV, OGG, AAC  
- **Video**: MP4, WebM, OGV
- **Documents**: PDF, TXT, HTML, CSS, JavaScript, JSON, XML, CSV
- **Archives**: ZIP, TAR, GZIP
- **Default**: application/octet-stream for unknown types

## Architecture Notes

The application follows a clean separation of concerns:

- **main.go**: HTTP routing and request handling
- **store/store.go**: Database operations and MySQL interaction
- **Bucket-as-Table**: Each bucket creates a corresponding MySQL table
- **Binary Storage**: Files are stored as LONGBLOB data with metadata

This design provides the flexibility of NoSQL-style bucket storage while maintaining ACID compliance and leveraging existing MySQL infrastructure.

## Error Handling

All endpoints return consistent JSON error responses:

```json
{
    "message": "Human-readable error description",
    "details": "Technical details for debugging"
}
```

Common HTTP status codes you'll encounter:
- **200**: Success
- **201**: Resource created
- **400**: Bad request (invalid JSON, missing parameters)
- **404**: Resource not found
- **405**: Method not allowed
- **500**: Internal server error

## Conclusion

Blobstore bridges the gap between traditional file storage and database reliability. Whether you're building a content management system, handling user uploads, or need a simple file API for your microservices, this project provides a solid foundation.

The combination of Go's performance, MySQL's reliability, and REST API simplicity makes this an ideal solution for teams already invested in MySQL infrastructure who want to add file storage capabilities without introducing additional complexity.

Ready to store some files? Start by creating your first bucket and see how easy file management can be when your database handles everything!