package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"io"
	"net/http"
	"strings"
	"testing"
)

//go:generate go tool oapi-codegen -config oapi-codegen-cfg.yaml ../../api/openapiv2/app.swagger.yaml

const (
	url = "http://localhost:8080"

	dummyLoginEndpoint = "/dummyLogin"
	registerEndpoint   = "/register"
	loginEndpoint      = "/login"
	pvzEndpoint        = "/pvz"
	receptionsEndpoint = "/receptions"
	productsEndpoint   = "/products"
)

func closeLastReceptionEndpoint(pvzId string) string {
	return fmt.Sprintf("/pvz/%s/close_last_reception", pvzId)
}

func deleteLastProductEndpoint(pvzId string) string {
	return fmt.Sprintf("/pvz/%s/delete_last_product", pvzId)
}

func Test_ReceptionCycle(t *testing.T) {
	moderatorToken, err := dummyToken("moderator")
	if err != nil {
		t.Fatalf("dummyToken error: %s", err.Error())
	}
	employeeToken, err := dummyToken("employee")
	if err != nil {
		t.Fatalf("dummyToken error: %s", err.Error())
	}

	pvzResponse, err := newPvzRequest(moderatorToken, "Казань")
	if err != nil {
		t.Fatalf("newPvzRequest error: %s", err.Error())
	}

	var pvz struct {
		ID string `json:"id"`
	}
	dec := json.NewDecoder(pvzResponse.Body)
	err = dec.Decode(&pvz)
	if err != nil {
		t.Fatalf("decode error: %s", err.Error())
	}

	newReceptionResponse, err := newReceptionRequest(employeeToken, pvz.ID)
	if err != nil {
		t.Fatalf("newReceptionRequest error: %s", err.Error())
	} else if newReceptionResponse.StatusCode != http.StatusCreated {
		t.Fatalf("newReceptionRequest returned wrong status code: %d", newReceptionResponse.StatusCode)
	}

	productTypes := []string{"электроника", "одежда", "обувь"}
	for i := range 50 {
		resp, err := newProductRequest(employeeToken, pvz.ID, productTypes[i%len(productTypes)])
		if err != nil {
			t.Fatalf("failed to create product: %s", err.Error())
		} else if resp.StatusCode != http.StatusCreated {
			t.Fatalf("failed to create product: %s", resp.Status)
		}
	}

	closeResp, err := closeReceptionRequest(employeeToken, pvz.ID)
	if err != nil {
		t.Fatalf("closeReceptionRequest error: %s", err.Error())
	} else if closeResp.StatusCode != http.StatusOK {
		t.Fatalf("closeReceptionRequest returned wrong status code: %d", closeResp.StatusCode)
	}

	var reception Reception
	dec = json.NewDecoder(closeResp.Body)
	err = dec.Decode(&reception)
	if err != nil {
		t.Fatalf("decode error: %s", err.Error())
	}

	t.Log(reception)
	if reception.PvzId.String() != pvz.ID {
		t.Fatalf("reception returned with wrong pvz id: %s", reception.PvzId.String())
	} else if reception.Status != "closed" {
		t.Fatalf("reception returned with wrong status: %s", reception.Status)
	} else if reception.Id == nil {
		t.Fatalf("reception returned with empty id")
	}

}

func Test_DummyLogin(t *testing.T) {
	employeeToken, err := dummyToken("employee")
	if err != nil {
		t.Fatal(err)
	}
	moderatorToken, err := dummyToken("moderator")
	if err != nil {
		t.Fatal(err)
	}

	testTokenPermissions(t, employeeToken, moderatorToken)
}

func Test_RegisterLogin(t *testing.T) {
	employeeCreds := fmt.Sprintf(`{"email":"user-%s@example.com","password":"password","role":"employee"}`, lo.RandomString(5, lo.AlphanumericCharset))
	moderatorCreds := fmt.Sprintf(`{"email":"moder-%s@example.com","password":"password","role":"moderator"}`, lo.RandomString(5, lo.AlphanumericCharset))

	// register
	employee1RegisterResponse, err := registerRequest(employeeCreds)
	if err != nil {
		t.Fatal(err)
	} else if employee1RegisterResponse.StatusCode != http.StatusCreated {
		t.Fatalf("Unexpected status code: %d", employee1RegisterResponse.StatusCode)
	}
	employee2RegisterResponse, err := registerRequest(employeeCreds)
	if err != nil {
		t.Fatal(err)
	} else if employee2RegisterResponse.StatusCode == http.StatusOK {
		t.Fatalf("Creating second user with same email should fail")
	}
	moderatorRegisterResponse, err := registerRequest(moderatorCreds)
	if err != nil {
		t.Fatal(err)
	} else if moderatorRegisterResponse.StatusCode != http.StatusCreated {
		t.Fatalf("Unexpected status code: %d", moderatorRegisterResponse.StatusCode)
	}

	// login
	employee1LoginResponse, err := loginRequest(`{"email":"user@example.com","password":"wrongPassword"}`)
	if err != nil {
		t.Fatal(err)
	}
	if employee1LoginResponse.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password should lead to unauthorized request")
	}
	employee1LoginResponse, err = loginRequest(employeeCreds)
	if err != nil {
		t.Fatal(err)
	}
	if employee1LoginResponse.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code: %d", employee1LoginResponse.StatusCode)
	}
	employeeToken, err := tokenFromBody(employee1LoginResponse.Body)
	if err != nil {
		t.Fatal(err)
	}

	moderatorLoginResponse, err := loginRequest(moderatorCreds)
	if err != nil {
		t.Fatal(err)
	}
	if moderatorLoginResponse.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code: %d", moderatorLoginResponse.StatusCode)
	}
	moderatorToken, err := tokenFromBody(moderatorLoginResponse.Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate roles
	testTokenPermissions(t, employeeToken, moderatorToken)
}

func testTokenPermissions(t *testing.T, employeeToken, moderatorToken string) {
	employeeResponse, err := newPvzRequest(employeeToken, "Москва")
	if err != nil {
		t.Fatal(err)
	}
	if employeeResponse.StatusCode == http.StatusOK {
		t.Fatalf("Employee should not have been permitted to create new pvz")
	}

	moderatorResponse, err := newPvzRequest(moderatorToken, "Москва")
	if err != nil {
		t.Fatal(err)
	}
	if moderatorResponse.StatusCode != http.StatusCreated {
		t.Fatalf("Employee should have been permitted to create new pvz, code: %d", moderatorResponse.StatusCode)
	}
}

func dummyToken(role string) (string, error) {
	request, err := http.NewRequest(http.MethodPost, url+dummyLoginEndpoint, strings.NewReader(`{"role":"`+role+`"}`))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to send %s request: %s", dummyLoginEndpoint, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got unexpected status code %s in %s response", resp.Status, dummyLoginEndpoint)
	}

	return tokenFromBody(resp.Body)
}

func tokenFromBody(body io.ReadCloser) (string, error) {
	if token, err := io.ReadAll(body); err != nil {
		return "", fmt.Errorf("failed to decode response body: %s", err)
	} else {
		return string(token[1 : len(token)-1]), nil
	}
}

func newPvzRequest(token, city string) (*http.Response, error) {
	data := `{"city":"` + city + `"}`
	request, err := http.NewRequest(http.MethodPost, url+pvzEndpoint, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)

	return http.DefaultClient.Do(request)
}

func newReceptionRequest(token, pvzId string) (*http.Response, error) {
	data := `{"pvzId":"` + pvzId + `"}`
	request, err := http.NewRequest(http.MethodPost, url+receptionsEndpoint, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)

	return http.DefaultClient.Do(request)
}

func closeReceptionRequest(token, pvzId string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPost, url+closeLastReceptionEndpoint(pvzId), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)

	return http.DefaultClient.Do(request)
}

func newProductRequest(token, pvzId, productType string) (*http.Response, error) {
	data := `{"pvzId":"` + pvzId + `", "type":"` + productType + `"}`
	request, err := http.NewRequest(http.MethodPost, url+productsEndpoint, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)

	return http.DefaultClient.Do(request)
}

func registerRequest(creds string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPost, url+registerEndpoint, strings.NewReader(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	return http.DefaultClient.Do(request)
}

func loginRequest(creds string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPost, url+loginEndpoint, strings.NewReader(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	return http.DefaultClient.Do(request)
}
