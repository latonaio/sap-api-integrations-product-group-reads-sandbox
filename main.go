package main

import (
	sap_api_caller "sap-api-integrations-product-group-reads-rmq-kube/SAP_API_Caller"  
	sap_api_input_reader "sap-api-integrations-product-group-reads-rmq-kube/SAP_API_Input_Reader"
	"sap-api-integrations-product-group-reads-rmq-kube/config"

	"github.com/latonaio/golang-logging-library-for-sap/logger"  
	rabbitmq "github.com/latonaio/rabbitmq-golang-client"
	"golang.org/x/xerrors"
)

func main() {
	l := logger.NewLogger()
	conf := config.NewConf()
	rmq, err := rabbitmq.NewRabbitmqClient(conf.RMQ.URL(), conf.RMQ.QueueFrom(), conf.RMQ.QueueTo())
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rmq.Close()

	caller := sap_api_caller.NewSAPAPICaller(
		conf.SAP.BaseURL(),
		conf.RMQ.QueueTo(),
		rmq,
		l,
	)

	iter, err := rmq.Iterator()
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rmq.Stop()

	for msg := range iter {
		err = callProductGroupProductGroup(caller, msg)
		if err != nil {
			msg.Fail()
			l.Error(err)
			continue
		}
		msg.Success()
	}
}

func callProductGroupProductGroup(caller *sap_api_caller.SAPAPICaller, msg rabbitmq.RabbitmqMessage) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = xerrors.Errorf("error occurred: %w", e)
			return
		}
	}()
	materialGroup, language, materialGroupName := extractData(msg.Data())
	accepter := getAccepter(msg.Data())
	caller.AsyncGetProductGroup(materialGroup, language, productGroupName, accepter)
	return nil
}

func extractData(data map[string]interface{}) (materialGroup, language, materialGroupName string) {
	sdc := sap_api_input_reader.ConvertToSDC(data)
	materialGroup = sdc.ProductGroup.MaterialGroup
	language = sdc.ProductGroup.ProductGroupText.Language
	materialGroupName = sdc.ProductGroup.ProductGroupText.ProductGroupName
	return
}

func getAccepter(data map[string]interface{}) []string {
	sdc := sap_api_input_reader.ConvertToSDC(data)
	accepter := sdc.Accepter
	if len(sdc.Accepter) == 0 {
		accepter = []string{"All"}
	}

	if accepter[0] == "All" {
		accepter = []string{
			"ProductGroup", "ProductGroupName",
		}
	}
	return accepter
}

