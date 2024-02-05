package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Provide a DynamoDB table resource.
		tbl, err := dynamodb.NewTable(ctx, "demo-dynamodb-table", &dynamodb.TableArgs{
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("CustomerId"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("ProductName"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("PurchaseAmount"),
					Type: pulumi.String("N"),
				},
			},
			BillingMode: pulumi.String("PROVISIONED"),
			GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray{
				&dynamodb.TableGlobalSecondaryIndexArgs{
					HashKey: pulumi.String("ProductName"),
					Name:    pulumi.String("ProductNameIndex"),
					NonKeyAttributes: pulumi.StringArray{
						pulumi.String("CustomerId"),
					},
					ProjectionType: pulumi.String("INCLUDE"),
					RangeKey:       pulumi.String("PurchaseAmount"),
					ReadCapacity:   pulumi.Int(10),
					WriteCapacity:  pulumi.Int(10),
				},
			},
			HashKey:      pulumi.String("CustomerId"),
			RangeKey:     pulumi.String("ProductName"),
			ReadCapacity: pulumi.Int(20),
			Tags: pulumi.StringMap{
				"Environment": pulumi.String("dev"),
				"Name":        pulumi.String("dynamodb-table-1"),
			},
			Ttl: &dynamodb.TableTtlArgs{
				AttributeName: pulumi.String("TimeToExist"),
				Enabled:       pulumi.Bool(false),
			},
			WriteCapacity: pulumi.Int(20),
		})
		if err != nil {
			return err
		}
		// Export output data.
		ctx.Export("demoTableArn", tbl.Arn)
		return nil
	})

}
