package main

import (
	"bytes"
	"encoding/json"
	"github.com/Minimalist-RestAPI-Golang/config"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a = App{}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
							(
								id SERIAL,
								name TEXT NOT NULL,
								price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
								CONSTRAINT products_pkey PRIMARY KEY (id)
							)`

func init() {
	//loads value from .env into the system
	if err := godotenv.Load("application.env"); err != nil {
		log.Print("No .env file found")
	}
}

func TestMain(m *testing.M) {
	conf := config.NewConfig()

	a.Initialize(conf.DbUsername, conf.DbPassword, conf.DbName)
	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

//func TestEmptyTable(t *testing.T) {
//	clearTable()
//
//	var tests = []struct {
//		name           string
//		in             *http.Request
//		out            *httptest.ResponseRecorder
//		expectedStatus int
//		expectedBody   string
//	}{
//		{
//			name: "clearTableTest",
//			in: httptest.NewRequest("GET", "/", nil),
//			out: httptest.NewRecorder(),
//			expectedStatus: http.StatusOK,
//			expectedBody: "[]",
//		},
//
//	}
//	for _, test := range tests {
//		test := test
//		t.Run(test.name, func(t *testing.T) {
//			a.Router.ServeHTTP(test.out, test.in)
//
//			if test.out.Code != test.expectedStatus{
//				t.Logf("expected %d\ngot:%d\n", test.expectedStatus, test.out.Code)
//				t.Fail()
//			}
//
//			body := test.out.Body.String()
//			if body != test.expectedBody{
//				t.Logf("expected %s\ngot: %s\n", test.expectedBody, test.out.Body)
//				t.Fail()
//			}
//		})
//	}
//}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	tests := []struct {
		name                   string
		in                     *http.Request
		out                    *httptest.ResponseRecorder
		expectedStatus         int
		expectedMapErrorKeyVal string
	}{
		{
			name:                   "GetNonExistentProductTest",
			in:                     httptest.NewRequest("GET", "/product/12", nil),
			out:                    httptest.NewRecorder(),
			expectedStatus:         http.StatusNotFound,
			expectedMapErrorKeyVal: "Product not found",
		},
	}
	for _, test := range tests {
		//test := test
		t.Run(test.name, func(t *testing.T) {
			a.Router.ServeHTTP(test.out, test.in)

			if test.out.Code != test.expectedStatus {
				t.Logf("expected %d\ngot:%d\n", test.expectedStatus, test.out.Code)
				t.Fail()
			}

			var m map[string]string
			if err := json.Unmarshal(test.out.Body.Bytes(), &m); err != nil {
				log.Println("an error occurred while trying to unmarshal the json")
			}
			if m["error"] != test.expectedMapErrorKeyVal {
				t.Logf("expected the error key of the response to be set to %s\ngot: %s\n", test.expectedMapErrorKeyVal, m["error"])
				t.Fail()
			}
		})
	}
}

//
func TestCreateProduct(t *testing.T) {
	clearTable()

	jsonStr := []byte(`{"name":"product 1", "price":"11.90"}`)

	tests := []struct {
		name           string
		in             *http.Request
		out            *httptest.ResponseRecorder
		expectedStatus int
	}{
		{
			name:           "CreateProductTest",
			in:             httptest.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr)),
			out:            httptest.NewRecorder(),
			expectedStatus: http.StatusCreated,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.in.Header.Set("Content-Type", "application/json")
			a.Router.ServeHTTP(test.out, test.in)

			if test.out.Code != test.expectedStatus {
				t.Logf("expected %d\ngot:%d\n", test.expectedStatus, test.out.Code)
				t.Logf("the error is: " + test.out.Body.String())
				t.Fail()
			}

			var m map[string]interface{}
			if err := json.Unmarshal(test.out.Body.Bytes(), &m); err != nil {
				log.Println("an error occurred while trying to unmarshal the json")
			}
			if m["name"] != "product 1" {
				t.Logf("expected the product name to be %s\ngot: %s\n", "product 1", m["name"])
				t.Fail()
			}

			if m["price"] != "11.90" {
				t.Logf("expected the product price to be %s\ngot: %s\n", "11.90", m["price"])
				t.Fail()
			}

			// the id is compared to 1.0 because JSON unmarshaling converts numbers to
			// floats, when the target is a map[string]interface{}
			if m["id"] != 1.0 {
				t.Logf("expected the id to be %s\ngot: %s\n", "1.0", m["price"])
				t.Fail()
			}
		})
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(2)

	tests := []struct {
		name           string
		in             *http.Request
		out            *httptest.ResponseRecorder
		expectedStatus int
	}{
		{
			name:           "GetProductTest",
			in:             httptest.NewRequest("GET", "/product/1", nil),
			out:            httptest.NewRecorder(),
			expectedStatus: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a.Router.ServeHTTP(test.out, test.in)

			if test.out.Code != test.expectedStatus {
				t.Logf("expected %d\ngot:%d\n", test.expectedStatus, test.out.Code)
				t.Fail()
			}
		})
	}
}

func TestGetProducts(t *testing.T) {
	clearTable()
	addProducts(7)

	tests := []struct {
		name          string
		in            *http.Request
		out           *httptest.ResponseRecorder
		expectedValue int
	}{
		{
			name:          "GetProductsTest",
			in:            httptest.NewRequest("GET", "/products?count=2&start=1", nil),
			out:           httptest.NewRecorder(),
			expectedValue: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a.Router.ServeHTTP(test.out, test.in)

			if test.out.Code != test.expectedValue {
				t.Logf("expected %d\ngot%d\n", test.expectedValue, test.out.Code)
				t.Fail()
			}
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProducts(2)

	jsonStr := []byte(`{"name":"product 1 updated", "price":"14.00"}`)

	tests := []struct {
		name           string
		in             *http.Request
		out            *httptest.ResponseRecorder
		expectedStatus int
	}{
		{
			name:           "UpdateProductTest",
			in:             httptest.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr)),
			out:            httptest.NewRecorder(),
			expectedStatus: http.StatusAccepted,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			x := httptest.NewRequest("GET", "/product/1", nil)
			a.Router.ServeHTTP(test.out, x)

			var originalProduct map[string]interface{}
			if err := json.Unmarshal(test.out.Body.Bytes(), &originalProduct); err != nil {
				log.Println("an error occurred while trying to unmarshal the json")
			}

			test.in.Header.Set("Content-Type", "application/json")
			g := httptest.NewRecorder()
			a.Router.ServeHTTP(g, test.in)
			//fmt.Println(test.out.Body.String())

			if g.Code != test.expectedStatus {
				t.Logf("expected %d\ngot:%d\n", test.expectedStatus, test.out.Code)
				t.Fail()
			}

			var m map[string]interface{}
			if err := json.Unmarshal(g.Body.Bytes(), &m); err != nil {
				log.Println("an error occurred while trying to unmarshal the json: " + err.Error())
			}
			if m["name"] == originalProduct["name"] {
				t.Logf("expected the product name to change from %s\n t0 %s\ngot: %s\n", originalProduct["name"], m["name"], m["name"])
				t.Fail()
			}

			if m["price"] == originalProduct["price"] {
				t.Logf("expected the product price to change from %s\n to %s\ngot: %s\n", originalProduct["price"], m["price"], m["price"])
				t.Fail()
			}

			// the id is compared to 1.0 because JSON unmarshaling converts numbers to
			// floats, when the target is a map[string]interface{}
			if m["id"] != originalProduct["id"] {
				t.Logf("expected the id to be the same %.1f\n got: %.1f\n", originalProduct["id"], m["id"])
				t.Fail()
			}
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(2)

	tests := []struct {
		name           string
		in             *http.Request
		out            *httptest.ResponseRecorder
		expectedStatus int
	}{
		{
			name:           "DeleteProductTest",
			in:             httptest.NewRequest("DELETE", "/product/1", nil),
			out:            httptest.NewRecorder(),
			expectedStatus: http.StatusOK,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			x := httptest.NewRequest("GET", "/product/1", nil)
			a.Router.ServeHTTP(test.out, x)

			if test.out.Code != http.StatusOK {
				t.Logf("while getting product: expected %d\ngot:%d\n", http.StatusOK, test.out.Code)
				t.Fail()
			}

			c := httptest.NewRecorder()
			a.Router.ServeHTTP(c, test.in)
			if c.Code != test.expectedStatus {
				t.Logf("while deleting product: expected %d\ngot:%d\n", test.expectedStatus, c.Code)
				t.Fail()
			}

			z := httptest.NewRecorder()
			o := httptest.NewRequest("GET", "/product/1", nil)
			a.Router.ServeHTTP(z, o)

			if z.Code != http.StatusNotFound {
				t.Logf("while getting product: expected %d\ngot:%d\n", http.StatusNotFound, z.Code)
				t.Fail()
			}
		})
	}
}

func ensureTableExists() {
	if _, err := a.Db.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.Db.Exec("DELETE FROM products")
	a.Db.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 1; i < count; i++ {
		a.Db.Exec("INSERT INTO products (name, price) VALUES ($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}
