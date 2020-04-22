![](https://github.com/StevenWeathers/exothermic-story-mapping/workflows/Go/badge.svg)
![](https://github.com/StevenWeathers/exothermic-story-mapping/workflows/Node.js%20CI/badge.svg)
![](https://github.com/StevenWeathers/exothermic-story-mapping/workflows/Docker/badge.svg)

# Exothermic Story Mapping

Exothermic is an open source agile story mapping tool, though storyboards are generic enough to be used for multiple excercises.

Each storyboard has the ability to add goals (rows), columns, and stories.

### **Uses WebSockets and [Svelte](https://svelte.dev/) frontend framework for a truly Reactive UI experience**

![image](https://user-images.githubusercontent.com/846933/77712629-b8933500-6faa-11ea-9b9f-b64a1f648f98.png)


## Building and running with docker-compose (easiest solution)

Prefered way of building and running the application with Postgres DB

```
docker-compose up --build
```

## Building

To run without docker you will need to first build, then setup the postgres DB,
and pass the user, pass, name, host, and port to the application as environment variables

```
DB_HOST=
DB_PORT=
DB_USER=
DB_PASS=
DB_NAME=
```

### Install dependencies
```
go get
go go install github.com/markbates/pkger/cmd/pkger
npm install
```

## Build with Make
```
make build
```
### OR manual steps

### Build static assets
```
npm run build
```

### bundle up static assets
```
pkger
```

### Build for current OS
```
go build
```

# Run Locally

Run the server and visit [http://localhost:8080](http://localhost:8080)