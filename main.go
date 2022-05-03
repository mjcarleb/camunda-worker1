package main

import (
	"context"
	"github.com/camunda/zeebe/clients/go/v8/pkg/entities"
	"github.com/camunda/zeebe/clients/go/v8/pkg/pb"
	"github.com/camunda/zeebe/clients/go/v8/pkg/worker"
	"github.com/camunda/zeebe/clients/go/v8/pkg/zbc"
	"log"
)

var readyClose = make(chan struct{})

func main() {

	credsProvider, err := zbc.NewOAuthCredentialsProvider(&zbc.OAuthProviderConfig{
		//AuthorizationServerURL: "https://login.cloud.camunda.io/oauth/token",                       //os.Getenv("ZEEBE_AUTHORIZATION_SERVER_URL"),
		Audience:     "df4c514a-5a6b-45a8-8bf3-1c147540a2f4.bru-2.zeebe.camunda.io",      //os.Getenv("ZEEBE_ADDRESS"),
		ClientID:     "F6twwTJrs9Vo77b8CEf~DT6ga1wqY~0I",                                 //os.Getenv("ZEEBE_CLIENT_ID"),
		ClientSecret: "vsQVRdmL7vPSfjDbouwKDatNe_N~oOVEpHzpQw-SF7Oo3kIo-vucAy-e615rHtKy", //os.Getenv("ZEEBE_CLIENT_SECRET"),
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
	gatewayAddr := "df4c514a-5a6b-45a8-8bf3-1c147540a2f4.bru-2.zeebe.camunda.io:443"
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

	//headers, err := job.GetCustomHeadersAsMap()
	//if err != nil {
	//	// failed to handle job as we require the custom job headers
	//	failJob(client, job)
	//	return
	//}

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	log.Println("Complete job", jobKey, "of type", job.Type)
	log.Println("tran_ref", variables["tran_ref"])
	log.Println("account", variables["account"])
	log.Println("security", variables["security"])
	log.Println("qty", variables["qty"])
	log.Println("tran_type", variables["tran_type"])
	log.Println("counter_party", variables["counter_party"])
	//log.Println("Processing order:", variables["orderId"])
	//log.Println("Collect money using payment method:", headers["method"])

	ctx := context.Background()
	_, err = request.Send(ctx)
	if err != nil {
		panic(err)
	}

	log.Println("Successfully completed job")
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
