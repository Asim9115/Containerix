package templates

import (
	"strings"
	"testing"
)

func TestGeneratePython(t *testing.T) {
	got := GeneratePython(		"RUN python manage.py collectstatic\n",
		"/app",
		"python manage.py runserver 0.0.0.0:8000",
		8000,)

		expected := []string{
		"FROM python:3.13",
		"WORKDIR /app",
		"COPY requirements.txt ./",
		"RUN pip install --no-cache-dir -r requirements.txt",
		"RUN python manage.py collectstatic",
		"EXPOSE 8000",
		`CMD ["sh", "-c", "python manage.py runserver 0.0.0.0:8000"]`,
	}

	for _, line := range expected {
		if !strings.Contains(got, line) {
			t.Errorf("expected Dockerfile to contain %q\nGot:\n%s", line, got)
		}
	}
}