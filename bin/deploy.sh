#!/usr/bin/env bash
die() {
  echo "${1:-argh}"
  ls ./cmd/bin && rm -rf ./cmd/bin/
  exit "${2:-1}"
}

hash sam  2>/dev/null || die "missing dep: sam"
hash aws  2>/dev/null || die "missing dep: aws"
hash ./bin/parse-yaml.sh || die "parse-yaml.sh not found."

profile=$1
[[ -z $profile ]] && die "Usage: $0 <profile>"

STACK_NAME="Recipe-api"

tags=$(./bin/parse-yaml.sh ./cf/tags.yaml) || die "failed to parse tags"
bucket_name=$(aws ssm get-parameter --profile "$profile" --name /s3/cfn-bucket/name --query "Parameter.Value" --output text) || die "failed to get name of cfn bucket"

echo "~~~ Build code using iterative golang compiling"
for path in ./cmd/*/; do
  dirname=$(basename "$path")

  echo "Building $dirname..."

  cd "./cmd/$dirname" || die "failed to cd into ./cmd/$dirname"
  GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o "../bin/$dirname/bootstrap" "$dirname.go"
  cd ../..
done

echo "~~~ Deploy infra stack"
sam deploy \
  --tags "$tags" \
  --no-fail-on-empty-changeset \
  --s3-bucket "${bucket_name}" \
  --stack-name "${STACK_NAME}-infra" \
  --s3-prefix "${STACK_NAME}-infra" \
  --capabilities "CAPABILITY_IAM" "CAPABILITY_NAMED_IAM" \
  --template "./cf/infra.yaml" \
  --region "ap-southeast-2" \
  --profile "${profile}" || die "failed to deploy stack "$STACK_NAME"-infra"

echo "~~~ Deploy processor stack"
sam deploy \
  --tags "$tags" \
  --no-fail-on-empty-changeset \
  --s3-bucket "${bucket_name}" \
  --stack-name "${STACK_NAME}-processors" \
  --s3-prefix "${STACK_NAME}-processors" \
  --capabilities "CAPABILITY_IAM" "CAPABILITY_NAMED_IAM" \
  --template "./cf/processors.yaml" \
  --region "ap-southeast-2" \
  --profile "${profile}" || die "failed to deploy stack "$STACK_NAME"-processors"

echo "~~~ Deploy gateway stack"
sam deploy \
  --tags "$tags" \
  --no-fail-on-empty-changeset \
  --s3-bucket "${bucket_name}" \
  --stack-name "${STACK_NAME}-gateway" \
  --s3-prefix "${STACK_NAME}-gateway" \
  --capabilities "CAPABILITY_IAM" "CAPABILITY_NAMED_IAM" \
  --template "./cf/gateway.yaml" \
  --region "ap-southeast-2" \
  --profile "${profile}" || die "failed to deploy stack "$STACK_NAME"-gateway"

echo "~~ cleaning up"
rm -rf ./cmd/bin/