#!/bin/bash
# LocalStack runs this once S3 is ready (mounted into /etc/localstack/init/ready.d).
set -euo pipefail
awslocal s3api create-bucket \
  --bucket book-images \
  --create-bucket-configuration LocationConstraint=ap-northeast-1 >/dev/null 2>&1 || true
echo "bucket book-images is ready"
