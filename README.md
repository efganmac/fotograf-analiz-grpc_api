# Project Architecture Diagram
![Untitled](https://github.com/efganmac/fotograf-analiz-grpc_api/assets/95138129/df032786-7983-41a8-ab53-4c48e99f07de)

# Project Setup Documentation

## Prerequisites
Before running this project, make sure the following tools are installed on your system:

- Go (1.15 or higher)
- Google Cloud SDK
- Docker (optional for Kafka)
- gRPC and Protocol Buffers

## Step 1: Cloning the Repo
To clone the project to your local system, use the following git command:

```bash
git clone https://github.com/efganmac/fotograf-analiz-grpc_api.git
cd fotograf-analiz-grpc_api
```

## Step 2: Setting Environment Variables
The project uses certain environment variables. To set these variables, create a `.env` file in the project's root directory and fill in the following variables:

```makefile
GOOGLE_APPLICATION_CREDENTIALS=<Path to Google Cloud service account key file>
KAFKA_BOOTSTRAP_SERVERS=<Kafka server addresses>
KAFKA_USERNAME=<Kafka username>
KAFKA_PASSWORD=<Kafka password>
GCP_STORAGE_BUCKET=<Google Cloud Storage bucket name>
GCP_PROJECT_ID=<Google Cloud project ID>
FIRESTORE_DB_ID=<Firestore database ID>
```

## Step 3: Installing Dependencies
To install Go dependencies, run the following command in the project's root directory:

```go
go mod tidy
```

## Step 4: Compiling gRPC Protocol Files
To compile the gRPC services and messages to Go code, run the following command:

```css
protoc --go_out=. --go-grpc_out=. api/proto/*.proto
```

## Step 5: Running the Application
To run the main Go application:

```go
go run main.go
```

## Step 6: Kafka and Google Cloud Settings
To use Kafka and Google Cloud Vision API, set up and configure the relevant services. If using Docker, you can create a Docker Compose file for Kafka.

Create a project in the Google Cloud console for Google Cloud Vision API and Google Cloud Storage, and enable the necessary APIs.

## Additional Notes
- To properly operate the application, you must use a service account with appropriate permissions for Google Cloud Storage and Firestore.
- Kafka configuration may vary according to the project's requirements.
- These steps cover the basic setup and running of the project for local development. For more details or specific configurations related to the project, check the project's README file or relevant documentation.
