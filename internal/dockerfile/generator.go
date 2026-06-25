package dockerfile

import (
	"fmt"
	"github.com/asim9115/containerix/internal/detector"
)


func GenerateNode(detected detector.DetectResult) (string, error) {
	return "", nil
}


func GenerateGo(detected detector.DetectResult) (string, error) {

	version := detected.Version
	if version == "" {
		version = "1.26.2"
	}

	content := fmt.Sprintf(`FROM golang:%s

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/server

EXPOSE 8000

CMD ["./app"]
`, version)

	return content, nil

}

func GeneratePython(detected detector.DetectResult) (string, error) {
	version := detected.Version

	if version == "" {
		version = "3.12"
	}

	content := fmt.Sprintf(`FROM python:%s

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "main.py"]
`, version)

	return content, nil
}