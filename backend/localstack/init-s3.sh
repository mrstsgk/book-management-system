#!/bin/bash
# LocalStack runs this every time S3 is ready (mounted into /etc/localstack/init/ready.d).
# LocalStack keeps objects in memory only, so the seed covers are uploaded again on
# every start; backend/seed/seed.sql points the sample books at these keys.
set -euo pipefail
awslocal s3api create-bucket \
  --bucket book-images \
  --create-bucket-configuration LocationConstraint=ap-northeast-1 >/dev/null 2>&1 || true
awslocal s3 cp /seed/books s3://book-images/seed/books --recursive --content-type image/png --only-show-errors
echo "bucket book-images is ready (seed covers uploaded)"
