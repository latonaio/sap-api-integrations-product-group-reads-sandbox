package sap_api_caller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	sap_api_output_formatter "sap-api-integrations-product-group-reads/SAP_API_Output_Formatter"
	"strings"
	"sync"

	"github.com/latonaio/golang-logging-library-for-sap/logger"
	"golang.org/x/xerrors"
)

type SAPAPICaller struct {
	baseURL string
	apiKey  string
	log     *logger.Logger
}

func NewSAPAPICaller(baseUrl string, l *logger.Logger) *SAPAPICaller {
	return &SAPAPICaller{
		baseURL: baseUrl,
		apiKey:  GetApiKey(),
		log:     l,
	}
}

func (c *SAPAPICaller) AsyncGetProductGroup(materialGroup, language, materialGroupName string, accepter []string) {
	wg := &sync.WaitGroup{}
	wg.Add(len(accepter))
	for _, fn := range accepter {
		switch fn {
		case "ProductGroup":
			func() {
				c.ProductGroup(materialGroup)
				wg.Done()
			}()
		case "ProductGroupName":
			func() {
				c.ProductGroupName(language, materialGroupName)
				wg.Done()
			}()
		default:
			wg.Done()
		}
	}

	wg.Wait()
}

func (c *SAPAPICaller) ProductGroup(materialGroup string) {
	productGroupData, err := c.callProductGroupSrvAPIRequirementProductGroup("A_ProductGroup", materialGroup)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(productGroupData)

	productGroupNameData, err := c.callToProductGroupName(productGroupData[0].ToProductGroupText)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(productGroupNameData)

}

func (c *SAPAPICaller) callProductGroupSrvAPIRequirementProductGroup(api, materialGroup string) ([]sap_api_output_formatter.ProductGroup, error) {
	url := strings.Join([]string{c.baseURL, "API_PRODUCTGROUP_SRV", api}, "/")
	req, _ := http.NewRequest("GET", url, nil)

	c.setHeaderAPIKeyAccept(req)
	c.getQueryWithProductGroup(req, materialGroup)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToProductGroup(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) callToProductGroupName(url string) ([]sap_api_output_formatter.ToProductGroupText, error) {
	req, _ := http.NewRequest("GET", url, nil)
	c.setHeaderAPIKeyAccept(req)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToToProductGroupText(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) ProductGroupName(language, materialGroupName string) {
	productGroupNameData, err := c.callProductGroupSrvAPIRequirementProductGroupName("A_ProductGroupText", language, materialGroupName)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(productGroupNameData)

}

func (c *SAPAPICaller) callProductGroupSrvAPIRequirementProductGroupName(api, language, materialGroupName string) ([]sap_api_output_formatter.ProductGroupText, error) {
	url := strings.Join([]string{c.baseURL, "API_PRODUCTGROUP_SRV", api}, "/")
	req, _ := http.NewRequest("GET", url, nil)

	c.setHeaderAPIKeyAccept(req)
	c.getQueryWithProductGroupName(req, language, materialGroupName)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToProductGroupText(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) setHeaderAPIKeyAccept(req *http.Request) {
	req.Header.Set("APIKey", c.apiKey)
	req.Header.Set("Accept", "application/json")
}

func (c *SAPAPICaller) getQueryWithProductGroup(req *http.Request, materialGroup string) {
	params := req.URL.Query()
	params.Add("$filter", fmt.Sprintf("MaterialGroup eq '%s'", materialGroup))
	req.URL.RawQuery = params.Encode()
}

func (c *SAPAPICaller) getQueryWithProductGroupName(req *http.Request, language, materialGroupName string) {
	params := req.URL.Query()
	params.Add("$filter", fmt.Sprintf("Language eq '%s' and substringof('%s', MaterialGroupName)", language, materialGroupName))
	req.URL.RawQuery = params.Encode()
}
