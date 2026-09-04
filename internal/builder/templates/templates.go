package templates

import "fmt"

func GeneratePython(buildCommand, directory, startCommand string, port int) string {
	return fmt.Sprintf(`FROM python:3.13

WORKDIR %s

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

RUN %s

EXPOSE %d

CMD ["sh", "-c", "%s"]
`, directory, buildCommand, port, startCommand)
}

func GenerateGo(buildCommand, directory, startCommand string, port int) string {
	return fmt.Sprintf(`FROM golang:1.24

WORKDIR %s

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN %s

EXPOSE %d

CMD ["sh", "-c", "%s"]
`, directory, buildCommand, port, startCommand)
}

func GenerateNode(buildCommand, directory, startCommand string, port int) string  {
	return fmt.Sprintf(`FROM node:22

WORKDIR %s

COPY package*.json ./
RUN npm install

COPY . .

RUN %s

EXPOSE %d

CMD ["sh", "-c", "%s"]
`, directory, buildCommand, port, startCommand)
}