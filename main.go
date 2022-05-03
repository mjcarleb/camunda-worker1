package main

import (
	"context"
	"fmt"
	"github.com/camunda/zeebe/clients/go/v8/pkg/pb"
	"github.com/camunda/zeebe/clients/go/v8/pkg/zbc"
)

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

	client, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:      "df4c514a-5a6b-45a8-8bf3-1c147540a2f4.bru-2.zeebe.camunda.io:443",
		CredentialsProvider: credsProvider,
	})

	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	topology, err := client.NewTopologyCommand().Send(ctx)
	if err != nil {
		panic(err)
	}

	for _, broker := range topology.Brokers {
		fmt.Println("Broker", broker.Host, ":", broker.Port)
		for _, partition := range broker.Partitions {
			fmt.Println("  Partition", partition.PartitionId, ":", roleToString(partition.Role))
		}
	}
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
