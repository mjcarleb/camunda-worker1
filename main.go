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
		Audience:     "df4c514a-5a6b-45a8-8bf3-1c147540a2f4.bru-2.zeebe.camunda.io:26500", //os.Getenv("ZEEBE_ADDRESS"),
		ClientID:     "zFshZ2p7iXjCucJIzjKwZ8wt66pjiF0e",                                  //os.Getenv("ZEEBE_CLIENT_ID"),
		ClientSecret: "q9SY4wmWnuiUGn9E_2LA4q.TrLLVy_5B_tXFHEWKNm0IHfAZl0vPoFTzwA70r9i6",  //os.Getenv("ZEEBE_CLIENT_SECRET"),
	})

	if err != nil {
		panic(err)
	}

	client, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:      "df4c514a-5a6b-45a8-8bf3-1c147540a2f4.bru-2.zeebe.camunda.io:26500",
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
