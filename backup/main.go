package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/backup"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/kms"
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
			},
			WriteCapacity: pulumi.Int(20),
		})
		// Create the AWS KMS key to encrypt backups.
		kmsKey, err := kms.NewKey(ctx, "demo-kms-key", &kms.KeyArgs{
			DeletionWindowInDays: pulumi.Int(10),
			Description:          pulumi.String("KMS key to encrypt backups"),
			EnableKeyRotation:    pulumi.Bool(true),
			KeyUsage:             pulumi.String("ENCRYPT_DECRYPT"),
			MultiRegion:          pulumi.Bool(false),
			Tags: pulumi.StringMap{
				"Environment": pulumi.String("dev"),
			},
		})
		_, err = kms.NewAlias(ctx, "alias/demo-kms-key", &kms.AliasArgs{
			TargetKeyId: kmsKey.KeyId,
		})
		// Provide a Backup Vault.
		vault, err := backup.NewVault(ctx, "demo-backup-vault", &backup.VaultArgs{
			KmsKeyArn:    kmsKey.Arn,
			ForceDestroy: pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Environment": pulumi.String("dev"),
			},
		})
		// Define a Backup Plan.
		plan, err := backup.NewPlan(ctx, "demo-backup-plan", &backup.PlanArgs{
			Rules: backup.PlanRuleArray{
				&backup.PlanRuleArgs{
					RuleName:         pulumi.String("DemoDailyBackups"),
					TargetVaultName:  vault.Name,
					Schedule:         pulumi.String("cron(0 12 * * ? *)"),
					StartWindow:      pulumi.Int(480),
					CompletionWindow: pulumi.Int(10080),
					Lifecycle: &backup.PlanRuleLifecycleArgs{
						ColdStorageAfter: pulumi.Int(7),
						DeleteAfter:      pulumi.Int(97),
					},
				},
			},
			Tags: pulumi.StringMap{
				"Environment": pulumi.String("dev"),
			},
		})
		// IAM Role that AWS Backup uses to authenticate when restoring and backing up resources.
		assumeRole, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: pulumi.StringRef("Allow"),
					Principals: []iam.GetPolicyDocumentStatementPrincipal{
						{
							Type: "Service",
							Identifiers: []string{
								"backup.amazonaws.com",
							},
						},
					},
					Actions: []string{
						"sts:AssumeRole",
					},
				},
			},
		}, nil)
		backupRole, err := iam.NewRole(ctx, "demo-backup-iam-role", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(assumeRole.Json),
			Tags: pulumi.StringMap{
				"Environment": pulumi.String("dev"),
			},
		})
		_, err = iam.NewRolePolicyAttachment(ctx, "demo-role-policy-attachment", &iam.RolePolicyAttachmentArgs{
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForBackup"),
			Role:      backupRole.Name,
		})
		// Assign AWS resources to a backup plan.
		_, err = backup.NewSelection(ctx, "demo-backup-selection", &backup.SelectionArgs{
			IamRoleArn: backupRole.Arn,
			PlanId:     plan.ID(),
			Resources: pulumi.StringArray{
				tbl.Arn,
			},
		})
		if err != nil {
			return err
		}
		return nil
	})

}
