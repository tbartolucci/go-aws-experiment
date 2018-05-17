#!/bin/sh

aws dynamodb create-table \
    --profile terraform \
    --region us-east-1 \
    --table-name PhotosAppUsers \
    --attribute-definitions AttributeName=ID,AttributeType=S AttributeName=Username,AttributeType=S \
    --key-schema KeyType=HASH,AttributeName=ID \
    --global-secondary-indexes 'IndexName=Username-index,KeySchema=[{AttributeName=Username,KeyType=HASH}],ProvisionedThroughput={ReadCapacityUnits=1,WriteCapacityUnits=1},Projection={ProjectionType=ALL}' \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1
 
aws dynamodb create-table \
    --profile terraform \
    --region us-east-1 \
    --table-name PhotosAppPhotos \
    --attribute-definitions AttributeName=ID,AttributeType=S AttributeName=UserID,AttributeType=S\
    --key-schema KeyType=HASH,AttributeName=ID \
    --global-secondary-indexes 'IndexName=UserID-index,KeySchema=[{AttributeName=UserID,KeyType=HASH}],ProvisionedThroughput={ReadCapacityUnits=1,WriteCapacityUnits=1},Projection={ProjectionType=ALL}' \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1
 
aws dynamodb create-table \
    --profile terraform \
    --region us-east-1 \
    --table-name PhotosAppComments \
    --attribute-definitions AttributeName=CreatedAt,AttributeType=S AttributeName=PhotoID,AttributeType=S \
    --key-schema KeyType=HASH,AttributeName=PhotoID KeyType=RANGE,AttributeName=CreatedAt \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1

aws dynamodb create-table \
    --profile terraform \
    --region us-east-1 \
    --table-name PhotosAppFollowers \
    --attribute-definitions AttributeName=UserID,AttributeType=S AttributeName=FollowerID,AttributeType=S \
    --key-schema KeyType=HASH,AttributeName=UserID KeyType=RANGE,AttributeName=FollowerID \
    --global-secondary-indexes 'IndexName=FollowerID-index,KeySchema=[{AttributeName=FollowerID,KeyType=HASH}],ProvisionedThroughput={ReadCapacityUnits=1,WriteCapacityUnits=1},Projection={ProjectionType=ALL}' \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1
