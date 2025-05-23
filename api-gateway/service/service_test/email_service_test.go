package service_test

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestOrderConfirmationTemplate(t *testing.T) {
	tmpl := `
	<table>
		{{range .Items}}
		<tr><td>{{.Name}}</td><td>{{.Quantity}}</td><td>{{.Price}}</td><td>{{.Subtotal}}</td></tr>
		{{end}}
	</table>`

	orderDetails := map[string]interface{}{
		"OrderID": "ORDER123",
		"Date":    "May 22, 2025",
		"Items": []map[string]interface{}{
			{"Name": "Bike A", "Quantity": 1, "Price": 500.0, "Subtotal": 500.0},
			{"Name": "Bike B", "Quantity": 2, "Price": 300.0, "Subtotal": 600.0},
		},
		"Total": 1100.0,
	}

	tmplt, err := template.New("test").Parse(tmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	err = tmplt.Execute(&buf, orderDetails)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Bike A") || !strings.Contains(out, "Bike B") {
		t.Errorf("template output missing expected content: %s", out)
	}
}

func TestEmailVerificationCodeTemplate(t *testing.T) {
	tmpl := `Your code is {{.Code}} for user {{.Username}}`

	data := map[string]string{
		"Code":     "123456",
		"Username": "testuser",
	}

	tmplt, err := template.New("verify").Parse(tmpl)
	if err != nil {
		t.Fatalf("template parse error: %v", err)
	}

	var buf bytes.Buffer
	err = tmplt.Execute(&buf, data)
	if err != nil {
		t.Fatalf("template execute error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "123456") || !strings.Contains(out, "testuser") {
		t.Errorf("output missing expected content: %s", out)
	}
}
