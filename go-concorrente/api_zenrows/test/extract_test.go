package api_zenrows

import (
	"io/ioutil"
	"testing"
	"encoding/json"
	internal "api_zenrows/internal"
)

func TestExtractObraDataFromLog(t *testing.T) {
	data, err := ioutil.ReadFile("html_rrt.log")
	if err != nil {
		t.Fatalf("Erro ao ler html_rrt.log: %v", err)
	}

       result, err := internal.ExtractObraData(string(data))
       if err != nil {
	       t.Errorf("Erro na extração: %v", err)
       } else {
	       jsonResult, err := json.MarshalIndent(result, "", "  ")
	       if err != nil {
		       t.Errorf("Erro ao converter para JSON: %v", err)
	       } else {
		       t.Logf("Resultado extraído em JSON:\n%s", string(jsonResult))
	       }
       }
}
