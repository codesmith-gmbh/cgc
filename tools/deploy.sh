#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
VPC_ID=$(aws ec2 describe-vpcs | jq -r '.Vpcs | map(select(.IsDefault == true))[0].VpcId')

aws cloudformation deploy \
    --template-file ${SCRIPT_DIR}/CGCTestStack.yaml \
    --stack-name CGCTestStack \
    --parameter-overrides \
        VpcId=${VPC_ID}
