package main

import (
	"context"
	"fmt"
	"github.com/camunda/zeebe/clients/go/v8/pkg/entities"
	"github.com/camunda/zeebe/clients/go/v8/pkg/pb"
	"github.com/camunda/zeebe/clients/go/v8/pkg/worker"
	"github.com/camunda/zeebe/clients/go/v8/pkg/zbc"
	"github.com/xuri/excelize/v2"
	"log"
	"reflect"
	"strconv"
)

//type trade struct {
//	tran_ref      string
//	account       string
//	security      string
//	qty           string
//	tran_type     string
//	counter_party string
//}

var readyClose = make(chan struct{})

func main() {

	credsProvider, err := zbc.NewOAuthCredentialsProvider(&zbc.OAuthProviderConfig{
		//AuthorizationServerURL: "https://login.cloud.camunda.io/oauth/token",                       //os.Getenv("ZEEBE_AUTHORIZATION_SERVER_URL"),
		Audience:     "241fa57d-bec2-4fec-9968-8f651682b023.bru-2.zeebe.camunda.io",      //os.Getenv("ZEEBE_ADDRESS"),
		ClientID:     "Dk0MPLoP_F0CmECfoidErdBdcLzZxLr.",                                 //os.Getenv("ZEEBE_CLIENT_ID"),
		ClientSecret: "_2VaLaDvXpFqtTlUbbinhs_oT_9e.8epWXIAMe5STottheBev9293zxQHG6jGaM~", //os.Getenv("ZEEBE_CLIENT_SECRET"),
	})

	if err != nil {
		panic(err)
	}

	//client, err := zbc.NewClient(&zbc.ClientConfig{
	//	GatewayAddress:      "df4c514a-5a6b-45a8-8bf3-1c147540a2f4.bru-2.zeebe.camunda.io:443",
	//	CredentialsProvider: credsProvider,
	//})
	//
	//if err != nil {
	//	panic(err)
	//}

	//ctx := context.Background()
	//topology, err := client.NewTopologyCommand().Send(ctx)
	//if err != nil {
	//	panic(err)
	//}

	//for _, broker := range topology.Brokers {
	//	fmt.Println("Broker", broker.Host, ":", broker.Port)
	//	for _, partition := range broker.Partitions {
	//		fmt.Println("  Partition", partition.PartitionId, ":", roleToString(partition.Role))
	//	}
	//}

	// Add code to do some work
	gatewayAddr := "241fa57d-bec2-4fec-9968-8f651682b023.bru-2.zeebe.camunda.io:443"
	plainText := false

	zbClient, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         gatewayAddr,
		UsePlaintextConnection: plainText,
		CredentialsProvider:    credsProvider,
	})

	if err != nil {
		panic(err)
	}
	jobWorker := zbClient.NewJobWorker().JobType("trade_match_worker").Handler(handleJob).Open()

	<-readyClose
	jobWorker.Close()
	jobWorker.AwaitClose()
}

func roleToString(role pb.Partition_PartitionBrokerRole) string {
	switch role {
	case pb.Partition_LEADER:
		return "Leader"
	case pb.Partition_FOLLOWER:
		return "Follower"
	default:
		return "Unknown"
	}
}

func handleJob(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	// create firm trade from variables passed into process as json
	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}
	firm_trade := map[string]string{
		"tran_ref":      variables["tran_ref"].(string),
		"account":       variables["account"].(string),
		"security":      variables["security"].(string),
		"qty":           variables["qty"].(string),
		"tran_type":     variables["tran_type"].(string),
		"counter_party": variables["counter_party"].(string),
	}

	//headers, err := job.GetCustomHeadersAsMap()
	//if err != nil {
	//	// failed to handle job as we require the custom job headers
	//	failJob(client, job)
	//	return
	//}

	// Assume no match
	match := false

	// Read in street master trades from excel
	f, err := excelize.OpenFile("master_trades.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := f.Rows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}

	street_trade := map[string]string{}

	row_cnt := 1 // excel rows are indexed starting at 1 in excelize package
	for rows.Next() {

		row, err := rows.Columns()
		if err != nil {
			fmt.Println(err)
		}
		for i, colCell := range row {
			switch i {
			case 0:
				street_trade["tran_ref"] = colCell
			case 1:
				street_trade["account"] = colCell
			case 2:
				street_trade["security"] = colCell
			case 3:
				street_trade["qty"] = colCell
			case 4:
				street_trade["tran_type"] = colCell
			case 5:
				street_trade["counter_party"] = colCell
			}
		}

		if reflect.DeepEqual(firm_trade, street_trade) {

			match = true

			//pos := "G" + "3" // starts at 1 and is actual row in spreadsheet  rows.seekRow
			pos := "G" + strconv.Itoa(row_cnt)

			err := f.SetCellValue("Sheet1", pos, "matched")

			if err != nil {
				fmt.Println(err)
				return
			}

			// save the file
			f.Save()
			f.Close()

			// leave iterating over the rows
			break

			fmt.Println("Never get here!")

		}
		row_cnt = row_cnt + 1
	}

	if err = rows.Close(); err != nil {
		fmt.Println(err)
	}

	// Send match_result back to Zeebe
	if match {
		variables["match_result"] = "matched"
	} else {
		variables["match_result"] = "no_match"
	}

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	ctx := context.Background()
	_, err = request.Send(ctx)
	if err != nil {
		panic(err)
	}

	// finish up
	log.Println("")
	log.Println("*********")
	log.Println("Worker completed job with match result = ", match)
	log.Println("Complete job", jobKey, "of type", job.Type)
	log.Println("tran_ref", variables["tran_ref"])
	log.Println("account", variables["account"])
	log.Println("security", variables["security"])
	log.Println("qty", variables["qty"])
	log.Println("tran_type", variables["tran_type"])
	log.Println("counter_party", variables["counter_party"])
	log.Println("*********")
	log.Println("")
	//log.Println("Processing order:", variables["orderId"])
	//log.Println("Collect money using payment method:", headers["method"])

	close(readyClose)
}

func failJob(client worker.JobClient, job entities.Job) {
	log.Println("Failed to complete job", job.GetKey())

	ctx := context.Background()
	_, err := client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send(ctx)
	if err != nil {
		panic(err)
	}
}
